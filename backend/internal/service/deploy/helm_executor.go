package deploy

import (
	"context"
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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// HelmExecutor 通过 Helm Runner Job 执行部署
type HelmExecutor struct {
	clientPool      *cluster.ClientsetPool
	deployRepo      repository.DeploymentRepository
	gitRepoRepo     repository.GitRepoRepository
	helmValuesRepo  repository.HelmValuesRepository
	cryptoSvc       *crypto.CryptoService
	wsHub           ProgressBroadcaster
	clusterRepo     repository.ClusterRepository
	settingSvc      *setting.SettingService
	notifDispatcher *notification.Dispatcher
}

// NewHelmExecutor 创建 Helm 执行器
func NewHelmExecutor(
	clientPool *cluster.ClientsetPool,
	deployRepo repository.DeploymentRepository,
	gitRepoRepo repository.GitRepoRepository,
	helmValuesRepo repository.HelmValuesRepository,
	cryptoSvc *crypto.CryptoService,
	wsHub ProgressBroadcaster,
	clusterRepo repository.ClusterRepository,
	settingSvc *setting.SettingService,
	notifDispatcher *notification.Dispatcher,
) *HelmExecutor {
	return &HelmExecutor{
		clientPool:      clientPool,
		deployRepo:      deployRepo,
		gitRepoRepo:     gitRepoRepo,
		helmValuesRepo:  helmValuesRepo,
		cryptoSvc:       cryptoSvc,
		wsHub:           wsHub,
		clusterRepo:     clusterRepo,
		settingSvc:      settingSvc,
		notifDispatcher: notifDispatcher,
	}
}

// resolveEnvSuffix 根据集群 env 查找 ENV_VALUES_MAP 映射（从系统配置动态读取）
func (e *HelmExecutor) resolveEnvSuffix(clusterID uint) string {
	if e.clusterRepo == nil || e.settingSvc == nil {
		return ""
	}
	c, err := e.clusterRepo.FindByID(clusterID)
	if err != nil {
		return ""
	}
	envValuesMap := e.settingSvc.GetEnvValuesMap()
	if suffix, ok := envValuesMap[c.Env]; ok {
		return suffix
	}
	return ""
}

// resolveJobNamespace 返回 Runner Job 运行的命名空间（从系统配置动态读取）
func (e *HelmExecutor) resolveJobNamespace() string {
	if e.settingSvc != nil {
		return e.settingSvc.GetHelmJobNamespace()
	}
	return "deployhub-jobs"
}

// Execute 异步执行 Helm 部署
func (e *HelmExecutor) Execute(deployment *model.Deployment, service *model.Service) error {
	go e.run(deployment.ID, service)
	return nil
}

func (e *HelmExecutor) run(deploymentID uint, service *model.Service) {
	log.Printf("[HelmExecutor] 开始执行 Helm 部署 %d", deploymentID)

	dep, err := e.deployRepo.FindByID(deploymentID)
	if err != nil {
		log.Printf("[HelmExecutor] 部署 %d 不存在: %v", deploymentID, err)
		return
	}

	// 获取 Helm Chart Git 仓库凭证（优先 deployment，fallback service）
	helmRepoID := dep.HelmRepoID
	if helmRepoID == nil {
		helmRepoID = service.HelmRepoID
	}
	if helmRepoID == nil {
		e.failDeploy(deploymentID, "未配置 Helm Chart Git 仓库")
		return
	}

	gitRepo, err := e.gitRepoRepo.FindByID(*helmRepoID)
	if err != nil {
		e.failDeploy(deploymentID, fmt.Sprintf("Helm Chart Git 仓库不存在: %v", err))
		return
	}

	credential, err := e.cryptoSvc.Decrypt(gitRepo.CredentialEncrypted)
	if err != nil {
		e.failDeploy(deploymentID, fmt.Sprintf("解密 Git 凭证失败: %v", err))
		return
	}

	clientset, err := e.clientPool.GetClientset(dep.ClusterID)
	if err != nil {
		e.failDeploy(deploymentID, fmt.Sprintf("获取集群连接失败: %v", err))
		return
	}

	ctx := context.Background()
	jobNs := e.resolveJobNamespace()

	// 构造带凭证的 Git URL
	gitURL := buildGitURLForHelm(gitRepo.URL, gitRepo.AuthType, credential)
	gitBranch := resolveHelmChartBranch(dep, service)

	// 创建系统 values ConfigMap（放在 Job 命名空间）
	valuesConfigMapName := fmt.Sprintf("helm-values-%d", deploymentID)
	e.ensureValuesConfigMap(ctx, clientset, jobNs, valuesConfigMapName, service.ID, dep.ClusterID)

	// 解析 ServiceAccount 和创建 Runner Job
	clusterSA := ""
	if cluster, cerr := e.clusterRepo.FindByID(dep.ClusterID); cerr == nil {
		clusterSA = cluster.HelmServiceAccount
	}
	sa := resolveHelmServiceAccount(dep, clusterSA)
	jobName := fmt.Sprintf("helm-deploy-%d-%d", deploymentID, time.Now().Unix())
	envSuffix := e.resolveEnvSuffix(dep.ClusterID)
	// Job 运行在 jobNs，helm 命令通过 --namespace 指向服务命名空间 dep.Namespace
	job := buildHelmUpgradeJob(jobName, jobNs, dep, service, gitURL, gitBranch, "", valuesConfigMapName, envSuffix, sa)

	// 保存执行命令到部署记录并广播（追加到已有的 preview 命令后面）
	helmCmd := job.Spec.Template.Spec.Containers[0].Args[0]
	upgradeCmd := fmt.Sprintf("\n\n# helm upgrade --install (deploy)\n# Runner Job NS: %s | Target NS: %s\n# SA: %s | Job: %s\n\n%s", jobNs, dep.Namespace, sa, jobName, helmCmd)
	existing := dep.DeployCommand
	e.updateDeployCommand(deploymentID, existing+upgradeCmd)
	e.broadcastLog(deploymentID, fmt.Sprintf("[DeployHub] Runner Job 命名空间: %s\n[DeployHub] 服务目标命名空间: %s\n", jobNs, dep.Namespace))
	e.broadcastLog(deploymentID, fmt.Sprintf("[DeployHub] 执行命令:\n%s\n---\n", helmCmd))

	_, err = clientset.BatchV1().Jobs(jobNs).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		e.failDeploy(deploymentID, fmt.Sprintf("创建 Helm Runner Job 失败: %v", err))
		return
	}

	log.Printf("[HelmExecutor] 部署 %d: Job %s 已创建 (jobNs=%s, targetNs=%s)", deploymentID, jobName, jobNs, dep.Namespace)

	// 监听 Job 状态并流式获取日志
	e.watchHelmJob(ctx, clientset, deploymentID, jobName, jobNs)

	// 清理临时资源
	e.cleanup(ctx, clientset, jobNs, valuesConfigMapName)
}

// Preview 执行 helm template dry-run，返回渲染后的 YAML
func (e *HelmExecutor) Preview(deployment *model.Deployment, service *model.Service) (string, error) {
	helmRepoID := resolveHelmRepoID(deployment, service)
	if helmRepoID == nil {
		return "", fmt.Errorf("未配置 Helm Chart Git 仓库")
	}

	gitRepo, err := e.gitRepoRepo.FindByID(*helmRepoID)
	if err != nil {
		return "", fmt.Errorf("Helm Chart Git 仓库不存在: %w", err)
	}

	credential, err := e.cryptoSvc.Decrypt(gitRepo.CredentialEncrypted)
	if err != nil {
		return "", fmt.Errorf("解密 Git 凭证失败: %w", err)
	}

	clientset, err := e.clientPool.GetClientset(deployment.ClusterID)
	if err != nil {
		return "", fmt.Errorf("获取集群连接失败: %w", err)
	}

	ctx := context.Background()
	jobNs := e.resolveJobNamespace()
	gitURL := buildGitURLForHelm(gitRepo.URL, gitRepo.AuthType, credential)
	gitBranch := resolveHelmChartBranch(deployment, service)
	if gitBranch == "" {
		gitBranch = "main"
	}

	valuesConfigMapName := fmt.Sprintf("helm-preview-values-%d", deployment.ID)
	e.ensureValuesConfigMap(ctx, clientset, jobNs, valuesConfigMapName, service.ID, deployment.ClusterID)

	// 解析 ServiceAccount
	clusterSA := ""
	if cluster, cerr := e.clusterRepo.FindByID(deployment.ClusterID); cerr == nil {
		clusterSA = cluster.HelmServiceAccount
	}
	sa := resolveHelmServiceAccount(deployment, clusterSA)
	jobName := fmt.Sprintf("helm-preview-%d-%d", deployment.ID, time.Now().Unix())
	envSuffix := e.resolveEnvSuffix(deployment.ClusterID)
	job := buildHelmTemplateJob(jobName, jobNs, deployment, service, gitURL, gitBranch, valuesConfigMapName, envSuffix, sa)

	// 保存 helm template 命令到部署记录
	templateCmd := job.Spec.Template.Spec.Containers[0].Args[0]
	deployCmd := fmt.Sprintf("# helm template (preview)\n# Runner Job NS: %s | Target NS: %s\n# SA: %s\n\n%s", jobNs, deployment.Namespace, sa, templateCmd)
	e.updateDeployCommand(deployment.ID, deployCmd)

	_, err = clientset.BatchV1().Jobs(jobNs).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		e.cleanup(ctx, clientset, jobNs, valuesConfigMapName)
		return "", fmt.Errorf("创建 Helm Template Job 失败: %w", err)
	}

	// 同步等待 Job 完成并收集输出
	result := e.waitAndCollectOutput(ctx, clientset, jobName, jobNs)
	e.cleanup(ctx, clientset, jobNs, valuesConfigMapName)

	return result, nil
}

// waitAndCollectOutput 等待 Job 完成并收集 Pod stdout
func (e *HelmExecutor) waitAndCollectOutput(ctx context.Context, clientset *kubernetes.Clientset, jobName, namespace string) string {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.After(3 * time.Minute)

	for {
		select {
		case <-timeout:
			return "# helm template 超时（3分钟）"
		case <-ticker.C:
			// 检查 Pod 状态（ImagePullBackOff 等不会让 Job Failed）
			pods, podErr := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("job-name=%s", jobName),
			})
			if podErr == nil && len(pods.Items) > 0 {
				pod := pods.Items[0]
				for _, cs := range pod.Status.ContainerStatuses {
					if cs.State.Waiting != nil {
						reason := cs.State.Waiting.Reason
						if reason == "ImagePullBackOff" || reason == "ErrImagePull" || reason == "CrashLoopBackOff" {
							return fmt.Sprintf("# Pod 启动失败: %s - %s", reason, cs.State.Waiting.Message)
						}
					}
				}
				for _, cs := range pod.Status.InitContainerStatuses {
					if cs.State.Waiting != nil {
						reason := cs.State.Waiting.Reason
						if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
							return fmt.Sprintf("# Init 容器启动失败: %s - %s", reason, cs.State.Waiting.Message)
						}
					}
				}
			}

			job, err := clientset.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
			if err != nil {
				continue
			}

			if job.Status.Succeeded > 0 || job.Status.Failed > 0 {
				pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
					LabelSelector: fmt.Sprintf("job-name=%s", jobName),
				})
				if err != nil || len(pods.Items) == 0 {
					return "# 无法获取 Pod 输出"
				}

				req := clientset.CoreV1().Pods(namespace).GetLogs(pods.Items[0].Name, &corev1.PodLogOptions{Container: "helm-template"})
				stream, err := req.Stream(ctx)
				if err != nil {
					return "# 获取日志失败"
				}
				defer stream.Close()

				data, _ := io.ReadAll(stream)
				return string(data)
			}
		}
	}
}

// Rollback 通过 Helm Runner Job 执行回滚
func (e *HelmExecutor) Rollback(deploymentID uint, service *model.Service, namespace string, clusterID uint, revision int) error {
	clientset, err := e.clientPool.GetClientset(clusterID)
	if err != nil {
		return fmt.Errorf("获取集群连接失败: %w", err)
	}

	dep, err := e.deployRepo.FindByID(deploymentID)
	if err != nil {
		return fmt.Errorf("部署记录不存在: %w", err)
	}

	ctx := context.Background()

	// 解析 ServiceAccount
	clusterSA := ""
	if cluster, cerr := e.clusterRepo.FindByID(clusterID); cerr == nil {
		clusterSA = cluster.HelmServiceAccount
	}
	sa := resolveHelmServiceAccount(dep, clusterSA)
	jobNs := e.resolveJobNamespace()
	jobName := fmt.Sprintf("helm-rollback-%d-%d", deploymentID, time.Now().Unix())
	job := buildHelmRollbackJob(jobName, jobNs, dep, service, revision, sa)

	_, err = clientset.BatchV1().Jobs(jobNs).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建 Helm Rollback Job 失败: %w", err)
	}

	log.Printf("[HelmExecutor] 部署 %d: Rollback Job %s 已创建 (jobNs=%s, revision=%d)", deploymentID, jobName, jobNs, revision)

	go e.watchHelmJob(ctx, clientset, deploymentID, jobName, jobNs)

	return nil
}

// watchHelmJob 轮询 Job 状态并流式获取 Pod 日志
func (e *HelmExecutor) watchHelmJob(ctx context.Context, clientset *kubernetes.Clientset, deploymentID uint, jobName, namespace string) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	logStreamed := false
	timeout := time.After(12 * time.Minute)

	for {
		select {
		case <-timeout:
			e.failDeploy(deploymentID, "Helm 部署超时（12分钟）")
			return
		case <-ticker.C:
			job, err := clientset.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
			if err != nil {
				log.Printf("[HelmExecutor] 部署 %d: 查询 Job 状态失败: %v", deploymentID, err)
				continue
			}

			// 尝试获取 Pod 日志
			if !logStreamed {
				pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
					LabelSelector: fmt.Sprintf("job-name=%s", jobName),
				})
				if err == nil && len(pods.Items) > 0 {
					pod := pods.Items[0]
					// init container 和 main container 都需要日志
					for _, cs := range pod.Status.InitContainerStatuses {
						if cs.State.Running != nil || cs.State.Terminated != nil {
							go e.streamPodContainerLog(ctx, clientset, deploymentID, pod.Name, namespace, "git-clone")
							break
						}
					}
					if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
						go e.streamPodContainerLog(ctx, clientset, deploymentID, pod.Name, namespace, "helm")
						logStreamed = true
					}
				}
			}

		if job.Status.Succeeded > 0 {
			now := time.Now()
			_ = e.deployRepo.UpdateStatus(deploymentID, model.DeployStatusSuccess)
			_ = e.updateDeployFields(deploymentID, map[string]interface{}{
				"finished_at": now,
			})
			e.broadcastLog(deploymentID, "\n[DeployHub] Helm 部署成功 ✓\n")
			log.Printf("[HelmExecutor] 部署 %d 成功", deploymentID)
			e.dispatchHelmEvent(deploymentID, model.EventDeploySuccess, "")
			return
		}

		if job.Status.Failed > 0 {
			// 读取 Pod 日志
			e.collectFailureLogs(ctx, clientset, deploymentID, jobName, namespace, logStreamed)
			e.failDeploy(deploymentID, "Helm Runner Job 执行失败")
			return
		}
		}
	}
}

// streamPodContainerLog 流式读取指定容器的日志
func (e *HelmExecutor) streamPodContainerLog(ctx context.Context, clientset *kubernetes.Clientset, deploymentID uint, podName, namespace, container string) {
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: container,
		Follow:    true,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		e.broadcastLog(deploymentID, fmt.Sprintf("[DeployHub] 获取 %s 日志流失败: %v\n", container, err))
		return
	}
	defer stream.Close()

	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			e.broadcastLog(deploymentID, string(buf[:n]))
		}
		if err != nil {
			if err != io.EOF {
				e.broadcastLog(deploymentID, fmt.Sprintf("\n[DeployHub] %s 日志流中断: %v\n", container, err))
			}
			return
		}
	}
}

// collectFailureLogs 失败时收集 Pod 日志
func (e *HelmExecutor) collectFailureLogs(ctx context.Context, clientset *kubernetes.Clientset, deploymentID uint, jobName, namespace string, logAlreadyStreamed bool) {
	if logAlreadyStreamed {
		time.Sleep(2 * time.Second)
		return
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil || len(pods.Items) == 0 {
		return
	}

	pod := pods.Items[0]
	for _, containerName := range []string{"git-clone", "helm"} {
		e.readContainerLogOnce(ctx, clientset, deploymentID, pod.Name, namespace, containerName)
	}
}

func (e *HelmExecutor) readContainerLogOnce(ctx context.Context, clientset *kubernetes.Clientset, deploymentID uint, podName, namespace, container string) {
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{Container: container})
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
		e.broadcastLog(deploymentID, fmt.Sprintf("\n--- %s 日志 ---\n%s", container, string(data)))
	}
}

// ensureValuesConfigMap 创建包含系统 values 的 ConfigMap
func (e *HelmExecutor) ensureValuesConfigMap(ctx context.Context, clientset *kubernetes.Clientset, namespace, configMapName string, serviceID, clusterID uint) {
	hv, err := e.helmValuesRepo.FindByServiceAndCluster(serviceID, clusterID)
	if err != nil || hv.Content == "" {
		// 创建空 ConfigMap（helm -f 会带 || true 容错）
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapName,
				Namespace: namespace,
				Labels:    map[string]string{"managed-by": "deployhub"},
			},
			Data: map[string]string{"values.yaml": "# empty"},
		}
		_, _ = clientset.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
		return
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
			Labels:    map[string]string{"managed-by": "deployhub"},
		},
		Data: map[string]string{"values.yaml": hv.Content},
	}
	_, _ = clientset.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
}

// cleanup 清理临时 ConfigMap
func (e *HelmExecutor) cleanup(ctx context.Context, clientset *kubernetes.Clientset, namespace, configMapName string) {
	_ = clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
}

func (e *HelmExecutor) failDeploy(deploymentID uint, message string) {
	now := time.Now()
	_ = e.deployRepo.UpdateStatus(deploymentID, model.DeployStatusFailed)
	_ = e.updateDeployFields(deploymentID, map[string]interface{}{"finished_at": now})
	e.broadcastLog(deploymentID, fmt.Sprintf("\n[DeployHub] Helm 部署失败: %s\n", message))
	log.Printf("[HelmExecutor] 部署 %d 失败: %s", deploymentID, message)
	e.dispatchHelmEvent(deploymentID, model.EventDeployFailed, message)
}

func (e *HelmExecutor) updateDeployCommand(deploymentID uint, command string) {
	_ = e.deployRepo.UpdateField(deploymentID, "deploy_command", command)
}

func (e *HelmExecutor) updateDeployFields(deploymentID uint, fields map[string]interface{}) error {
	dep, err := e.deployRepo.FindByID(deploymentID)
	if err != nil {
		return err
	}
	if t, ok := fields["finished_at"]; ok {
		if ft, ok := t.(time.Time); ok {
			dep.FinishedAt = &ft
		}
	}
	return e.deployRepo.Update(dep)
}

func (e *HelmExecutor) broadcastLog(deploymentID uint, chunk string) {
	if e.wsHub != nil {
		e.wsHub.Broadcast(fmt.Sprintf("deployment:%d", deploymentID), []byte(chunk))
	}
}

// dispatchHelmEvent 发送 Helm 部署相关通知
func (e *HelmExecutor) dispatchHelmEvent(deploymentID uint, eventType, failReason string) {
	if e.notifDispatcher == nil {
		return
	}
	dep, err := e.deployRepo.FindByID(deploymentID)
	if err != nil {
		return
	}
	svcName := ""
	if dep.Service != nil {
		svcName = dep.Service.Name
	}
	payload := notification.NotificationPayload{
		ServiceName: svcName,
		Namespace:   dep.Namespace,
		ImageTag:    dep.ImageTag,
		FailReason:  failReason,
		DeployID:    deploymentID,
	}
	if c, err := e.clusterRepo.FindByID(dep.ClusterID); err == nil {
		payload.ClusterName = c.Name
		payload.Env = c.Env
	}
	e.notifDispatcher.Dispatch(dep.ServiceID, eventType, payload)
}

// buildGitURLForHelm 构造带凭证的 Git URL
func buildGitURLForHelm(repoURL, authType, credential string) string {
	if authType != "token" || credential == "" {
		return repoURL
	}
	if strings.HasPrefix(repoURL, "https://") {
		return "https://" + credential + "@" + strings.TrimPrefix(repoURL, "https://")
	}
	if strings.HasPrefix(repoURL, "http://") {
		return "http://" + credential + "@" + strings.TrimPrefix(repoURL, "http://")
	}
	return repoURL
}
