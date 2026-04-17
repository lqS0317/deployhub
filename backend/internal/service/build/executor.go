package build

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/repository"
	"deployhub/internal/service/cluster"
	"deployhub/internal/service/crypto"
	"deployhub/internal/service/notification"
	"deployhub/internal/service/setting"
	"deployhub/internal/ws"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// BuildExecutor 负责在 K8s 集群中创建 Kaniko Job 执行镜像构建
type BuildExecutor struct {
	clientPool      *cluster.ClientsetPool
	buildRepo       repository.BuildRepository
	serviceRepo     repository.ServiceRepository
	registryRepo    repository.RegistryRepository
	gitRepoRepo     repository.GitRepoRepository
	clusterRepo     repository.ClusterRepository
	cryptoSvc       *crypto.CryptoService
	settingSvc      *setting.SettingService
	wsHub           *ws.Hub
	notifDispatcher *notification.Dispatcher
}

// NewBuildExecutor 创建构建执行器
func NewBuildExecutor(
	clientPool *cluster.ClientsetPool,
	buildRepo repository.BuildRepository,
	serviceRepo repository.ServiceRepository,
	registryRepo repository.RegistryRepository,
	gitRepoRepo repository.GitRepoRepository,
	clusterRepo repository.ClusterRepository,
	cryptoSvc *crypto.CryptoService,
	settingSvc *setting.SettingService,
	wsHub *ws.Hub,
	notifDispatcher *notification.Dispatcher,
) *BuildExecutor {
	return &BuildExecutor{
		clientPool:      clientPool,
		buildRepo:       buildRepo,
		serviceRepo:     serviceRepo,
		registryRepo:    registryRepo,
		gitRepoRepo:     gitRepoRepo,
		clusterRepo:     clusterRepo,
		cryptoSvc:       cryptoSvc,
		settingSvc:      settingSvc,
		wsHub:           wsHub,
		notifDispatcher: notifDispatcher,
	}
}

// resolveJobNamespace 返回构建 Job 运行的命名空间（复用系统设置的 helm_job_namespace）
func (e *BuildExecutor) resolveJobNamespace() string {
	if e.settingSvc != nil {
		return e.settingSvc.GetHelmJobNamespace()
	}
	return "default"
}

// Execute 异步执行构建：创建 Kaniko Job 并监听日志
func (e *BuildExecutor) Execute(buildID uint) {
	go e.run(buildID)
}

func (e *BuildExecutor) run(buildID uint) {
	log.Printf("[BuildExecutor] 开始执行构建 %d", buildID)

	build, err := e.buildRepo.FindByID(buildID)
	if err != nil {
		log.Printf("[BuildExecutor] 构建 %d 不存在: %v", buildID, err)
		return
	}

	svc, err := e.serviceRepo.FindByID(build.ServiceID)
	if err != nil {
		e.failByID(buildID, fmt.Sprintf("服务不存在: %v", err))
		return
	}

	gitRepo, err := e.gitRepoRepo.FindByID(svc.GitRepoID)
	if err != nil {
		e.failByID(buildID, fmt.Sprintf("Git 仓库不存在: %v", err))
		return
	}

	gitCredential, err := e.cryptoSvc.Decrypt(gitRepo.CredentialEncrypted)
	if err != nil {
		e.failByID(buildID, fmt.Sprintf("解密 Git 凭证失败: %v", err))
		return
	}

	// 优先从 build 自身读取 registry_id，fallback 到 service
	var registryID uint
	if build.RegistryID != nil {
		registryID = *build.RegistryID
	} else if svc.RegistryID != nil {
		registryID = *svc.RegistryID
	}
	if registryID == 0 {
		e.failByID(buildID, "未配置镜像仓库")
		return
	}
	registry, err := e.registryRepo.FindByID(registryID)
	if err != nil {
		e.failByID(buildID, fmt.Sprintf("镜像仓库不存在: %v", err))
		return
	}

	clientset, err := e.clientPool.GetClientset(build.BuildClusterID)
	if err != nil {
		e.failByID(buildID, fmt.Sprintf("获取集群连接失败: %v", err))
		return
	}

	// 更新为 building 状态
	now := time.Now()
	if err := e.updateBuildFields(buildID, map[string]interface{}{
		"status":     model.BuildStatusBuilding,
		"started_at": now,
	}); err != nil {
		log.Printf("[BuildExecutor] 构建 %d 更新状态失败: %v", buildID, err)
	}

	gitURL := buildGitURL(gitRepo.URL, gitRepo.AuthType, gitCredential)

	// 构造完整镜像地址：优先 build.ImageRepo，fallback svc.ImageRepo
	imageRepo := build.ImageRepo
	if imageRepo == "" {
		imageRepo = svc.ImageRepo
	}
	fullImage := build.ImageTag
	if !strings.Contains(fullImage, "/") {
		// tag 不含 repo 路径，需要拼接
		if imageRepo != "" {
			fullImage = fmt.Sprintf("%s:%s", imageRepo, build.ImageTag)
		}
	}
	if fullImage == "" || !strings.Contains(fullImage, "/") {
		e.failByID(buildID, "镜像地址无效：请确保配置了镜像路径（image_repo）")
		return
	}

	jobName := fmt.Sprintf("build-%d-%d", buildID, now.Unix())
	_ = e.updateBuildFields(buildID, map[string]interface{}{"kaniko_job_name": jobName})

	namespace := e.resolveJobNamespace()
	ctx := context.Background()

	// ECR 通过 IRSA 认证，不需要 Docker config Secret
	isECR := registry.Provider == "ecr" || strings.Contains(registry.URL, ".ecr.")
	secretName := ""
	if !isECR {
		secretName = fmt.Sprintf("build-%d-docker-cfg", buildID)
		if err := e.ensureDockerConfigSecret(ctx, clientset, namespace, secretName, registry); err != nil {
			e.failByID(buildID, fmt.Sprintf("创建镜像仓库凭证 Secret 失败: %v", err))
			return
		}
	}

	// 优先从 build 自身读取 dockerfile_path，fallback 到 service
	dockerfilePath := build.DockerfilePath
	if dockerfilePath == "" || dockerfilePath == "./Dockerfile" {
		dockerfilePath = svc.DockerfilePath
	}
	buildCtx := build.BuildContext
	if buildCtx == "" {
		buildCtx = "."
	}

	// 读取构建集群配置的 ServiceAccount（用于 IRSA 等场景，如推送 ECR 镜像）
	buildSA := ""
	if c, cerr := e.clusterRepo.FindByID(build.BuildClusterID); cerr == nil {
		buildSA = c.BuildServiceAccount
	}
	job := e.buildKanikoJob(jobName, namespace, gitURL, build.GitBranch, build.GitCommit, dockerfilePath, fullImage, buildCtx, registry, secretName, buildSA)

	_, err = clientset.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		e.failByID(buildID, fmt.Sprintf("创建 Kaniko Job 失败: %v", err))
		return
	}

	log.Printf("[BuildExecutor] 构建 %d: Job %s 已创建", buildID, jobName)
	e.appendLogByID(buildID, fmt.Sprintf("[DeployHub] 构建 Job %s 已创建，等待执行...\n", jobName))

	e.watchJob(ctx, clientset, buildID, jobName, namespace)
}

func (e *BuildExecutor) buildKanikoJob(jobName, namespace, gitURL, branch, commit, dockerfile, destination, buildContext string, registry *model.Registry, dockerSecretName, serviceAccount string) *batchv1.Job {
	backoffLimit := int32(0)
	ttl := int32(3600)

	gitContext := fmt.Sprintf("git://%s#refs/heads/%s", strings.TrimPrefix(strings.TrimPrefix(gitURL, "https://"), "http://"), branch)
	if commit != "" {
		gitContext = fmt.Sprintf("git://%s#%s", strings.TrimPrefix(strings.TrimPrefix(gitURL, "https://"), "http://"), commit)
	}

	args := []string{
		fmt.Sprintf("--context=%s", gitContext),
		fmt.Sprintf("--dockerfile=%s", dockerfile),
		fmt.Sprintf("--destination=%s", destination),
		"--cache=true",
		"--snapshot-mode=redo",
	}
	// 支持 monorepo：指定构建子目录
	if buildContext != "" && buildContext != "." && buildContext != "./" {
		args = append(args, fmt.Sprintf("--context-sub-path=%s", buildContext))
	}

	// 如果使用非 HTTPS 的 registry 则加 insecure 标志
	if strings.HasPrefix(registry.URL, "http://") {
		args = append(args, fmt.Sprintf("--insecure-registry=%s", strings.TrimPrefix(registry.URL, "http://")))
	}

	container := corev1.Container{
		Name:  "kaniko",
		Image: "gcr.io/kaniko-project/executor:latest",
		Args:  args,
	}

	var volumes []corev1.Volume
	if dockerSecretName != "" {
		container.VolumeMounts = []corev1.VolumeMount{
			{Name: "docker-config", MountPath: "/kaniko/.docker"},
		}
		volumes = append(volumes, corev1.Volume{
			Name: "docker-config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: dockerSecretName,
					Items:      []corev1.KeyToPath{{Key: ".dockerconfigjson", Path: "config.json"}},
				},
			},
		})
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        "deployhub-build",
				"managed-by": "deployhub",
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        "deployhub-build",
						"build-id":   fmt.Sprintf("%d", 0),
						"managed-by": "deployhub",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccount,
					RestartPolicy:      corev1.RestartPolicyNever,
					Containers:         []corev1.Container{container},
					Volumes:            volumes,
				},
			},
		},
	}
}

// watchJob 轮询 Job 状态并流式获取日志
func (e *BuildExecutor) watchJob(ctx context.Context, clientset *kubernetes.Clientset, buildID uint, jobName, namespace string) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	logStreamed := false
	timeout := time.After(30 * time.Minute)

	for {
		select {
		case <-timeout:
			e.failByID(buildID, "构建超时（30分钟）")
			return
		case <-ticker.C:
			job, err := clientset.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
			if err != nil {
				log.Printf("[BuildExecutor] 构建 %d: 查询 Job 状态失败: %v", buildID, err)
				continue
			}

			// 尝试获取 Pod 日志
			if !logStreamed {
				pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
					LabelSelector: fmt.Sprintf("job-name=%s", jobName),
				})
				if err == nil && len(pods.Items) > 0 {
					pod := pods.Items[0]
					if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
						go e.streamPodLog(ctx, clientset, buildID, pod.Name, namespace)
						logStreamed = true
					}
				}
			}

		if job.Status.Succeeded > 0 {
			now := time.Now()
			_ = e.updateBuildFields(buildID, map[string]interface{}{
				"status":      model.BuildStatusSuccess,
				"finished_at": now,
			})
			e.appendLogByID(buildID, "\n[DeployHub] 构建成功 ✓\n")
			log.Printf("[BuildExecutor] 构建 %d 成功", buildID)
			e.dispatchBuildEvent(buildID, model.EventBuildSuccess, "")
			return
		}

		if job.Status.Failed > 0 {
			// 获取 Pod 的失败原因和日志
			reason := e.collectFailureInfo(ctx, clientset, buildID, jobName, namespace, logStreamed)
			e.failByID(buildID, reason)
			return
		}
		}
	}
}

// streamPodLog 流式读取 Pod 日志并追加到构建记录
func (e *BuildExecutor) streamPodLog(ctx context.Context, clientset *kubernetes.Clientset, buildID uint, podName, namespace string) {
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: "kaniko",
		Follow:    true,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		e.appendLogByID(buildID, fmt.Sprintf("[DeployHub] 获取日志流失败: %v\n", err))
		return
	}
	defer stream.Close()

	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			e.appendLogByID(buildID, string(buf[:n]))
		}
		if err != nil {
			if err != io.EOF {
				e.appendLogByID(buildID, fmt.Sprintf("\n[DeployHub] 日志流中断: %v\n", err))
			}
			return
		}
	}
}

// collectFailureInfo 收集 Pod 失败原因和日志
func (e *BuildExecutor) collectFailureInfo(ctx context.Context, clientset *kubernetes.Clientset, buildID uint, jobName, namespace string, logAlreadyStreamed bool) string {
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil || len(pods.Items) == 0 {
		return "Kaniko Job 执行失败（无法获取 Pod 信息）"
	}

	pod := pods.Items[0]
	var reason string

	// 从容器状态中提取退出原因
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == "kaniko" && cs.State.Terminated != nil {
			t := cs.State.Terminated
			reason = fmt.Sprintf("退出码 %d", t.ExitCode)
			if t.Reason != "" {
				reason += fmt.Sprintf(", 原因: %s", t.Reason)
			}
			if t.Message != "" {
				reason += fmt.Sprintf(", 信息: %s", t.Message)
			}
		}
	}

	// 如果日志还没流式获取过，主动读取一次完整日志
	if !logAlreadyStreamed {
		e.readPodLogOnce(ctx, clientset, buildID, pod.Name, namespace)
	} else {
		// 等一下让流式日志写完
		time.Sleep(2 * time.Second)
	}

	if reason == "" {
		reason = "Kaniko Job 执行失败"
	}
	return reason
}

// readPodLogOnce 一次性读取 Pod 日志（用于 Pod 已结束的场景）
func (e *BuildExecutor) readPodLogOnce(ctx context.Context, clientset *kubernetes.Clientset, buildID uint, podName, namespace string) {
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: "kaniko",
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		return
	}
	defer stream.Close()

	data, err := io.ReadAll(stream)
	if err != nil {
		return
	}

	if len(data) > 0 {
		e.appendLogByID(buildID, string(data))
	}
}

// failByID 通过 ID 直接更新失败状态（避免 Preload 对象的 Save 问题）
func (e *BuildExecutor) failByID(buildID uint, message string) {
	now := time.Now()
	_ = e.updateBuildFields(buildID, map[string]interface{}{
		"status":      model.BuildStatusFailed,
		"finished_at": now,
	})
	e.appendLogByID(buildID, fmt.Sprintf("\n[DeployHub] 构建失败: %s\n", message))
	log.Printf("[BuildExecutor] 构建 %d 失败: %s", buildID, message)
	e.dispatchBuildEvent(buildID, model.EventBuildFailed, message)
}

// appendLogByID 通过 ID 追加日志
func (e *BuildExecutor) appendLogByID(buildID uint, chunk string) {
	_ = e.buildRepo.AppendLog(buildID, chunk)
	if e.wsHub != nil {
		e.wsHub.Broadcast(fmt.Sprintf("build:%d", buildID), []byte(chunk))
	}
}

// updateBuildFields 直接按字段更新，避免 GORM Save 关联对象问题
func (e *BuildExecutor) updateBuildFields(buildID uint, fields map[string]interface{}) error {
	return e.buildRepo.UpdateFields(buildID, fields)
}

// ensureDockerConfigSecret 创建包含镜像仓库认证的 Secret
func (e *BuildExecutor) ensureDockerConfigSecret(ctx context.Context, clientset *kubernetes.Clientset, namespace, secretName string, registry *model.Registry) error {
	authConfigJSON, err := e.cryptoSvc.Decrypt(registry.AuthConfigEncrypted)
	if err != nil {
		return fmt.Errorf("解密镜像仓库凭证失败: %w", err)
	}

	// 解析 {"username":"xx","password":"xx"}
	var authCfg struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal([]byte(authConfigJSON), &authCfg); err != nil {
		return fmt.Errorf("解析镜像仓库凭证失败: %w", err)
	}

	// 构造 Docker config.json 格式
	registryHost := extractRegistryHost(registry.URL)
	authStr := base64.StdEncoding.EncodeToString([]byte(authCfg.Username + ":" + authCfg.Password))
	dockerConfig := map[string]interface{}{
		"auths": map[string]interface{}{
			registryHost: map[string]string{
				"auth": authStr,
			},
		},
	}
	dockerConfigBytes, _ := json.Marshal(dockerConfig)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        "deployhub-build",
				"managed-by": "deployhub",
			},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": dockerConfigBytes,
		},
	}

	// 先尝试删除已有的同名 Secret
	_ = clientset.CoreV1().Secrets(namespace).Delete(ctx, secretName, metav1.DeleteOptions{})

	_, err = clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

// extractRegistryHost 从 registry URL 提取 host（Docker Hub 用 https://index.docker.io/v1/）
func extractRegistryHost(url string) string {
	host := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	host = strings.TrimSuffix(host, "/")
	// Docker Hub 特殊处理
	if host == "docker.io" || host == "hub.docker.com" || host == "registry-1.docker.io" || host == "" {
		return "https://index.docker.io/v1/"
	}
	return host
}

// DispatchEvent 对外暴露的通知发送方法（供 handler 调用）
func (e *BuildExecutor) DispatchEvent(buildID uint, eventType, failReason string) {
	e.dispatchBuildEvent(buildID, eventType, failReason)
}

// CancelJob 删除 K8s 上的构建 Job 和关联 Pod
func (e *BuildExecutor) CancelJob(buildID uint) {
	build, err := e.buildRepo.FindByID(buildID)
	if err != nil || build.KanikoJobName == "" {
		return
	}
	clientset, err := e.clientPool.GetClientset(build.BuildClusterID)
	if err != nil {
		log.Printf("[BuildExecutor] 取消构建 %d: 获取集群连接失败: %v", buildID, err)
		return
	}

	ctx := context.Background()
	namespace := e.resolveJobNamespace()
	propagation := metav1.DeletePropagationBackground
	err = clientset.BatchV1().Jobs(namespace).Delete(ctx, build.KanikoJobName, metav1.DeleteOptions{
		PropagationPolicy: &propagation,
	})
	if err != nil {
		log.Printf("[BuildExecutor] 取消构建 %d: 删除 Job %s 失败: %v", buildID, build.KanikoJobName, err)
	} else {
		log.Printf("[BuildExecutor] 取消构建 %d: Job %s 已删除", buildID, build.KanikoJobName)
	}
}

// dispatchBuildEvent 发送构建相关通知
func (e *BuildExecutor) dispatchBuildEvent(buildID uint, eventType, failReason string) {
	if e.notifDispatcher == nil {
		return
	}
	b, err := e.buildRepo.FindByID(buildID)
	if err != nil {
		return
	}
	svcName := ""
	if svc, err := e.serviceRepo.FindByID(b.ServiceID); err == nil {
		svcName = svc.Name
	}
	e.notifDispatcher.Dispatch(b.ServiceID, eventType, notification.NotificationPayload{
		ServiceName: svcName,
		ImageTag:    b.ImageTag,
		FailReason:  failReason,
		BuildID:     buildID,
	})
}

// buildGitURL 构造带凭证的 Git URL
func buildGitURL(repoURL, authType, credential string) string {
	if authType != "token" || credential == "" {
		return repoURL
	}
	// https://github.com/owner/repo.git → https://token@github.com/owner/repo.git
	if strings.HasPrefix(repoURL, "https://") {
		return "https://" + credential + "@" + strings.TrimPrefix(repoURL, "https://")
	}
	if strings.HasPrefix(repoURL, "http://") {
		return "http://" + credential + "@" + strings.TrimPrefix(repoURL, "http://")
	}
	return repoURL
}
