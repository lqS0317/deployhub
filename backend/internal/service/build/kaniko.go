package build

import (
	"fmt"

	"deployhub/internal/model"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateKanikoJob 生成 Kaniko 构建 K8s Job 规格
// gitCredential: Git 仓库访问凭据（token/password）
// registryAuth: 镜像仓库认证信息（base64 编码的 docker config JSON）
// namespace: Job 运行的命名空间
func GenerateKanikoJob(build *model.Build, service *model.Service, gitCredential, registryAuth, namespace string) *batchv1.Job {
	jobName := fmt.Sprintf("kaniko-build-%d", build.ID)
	backoffLimit := int32(0)

	gitURL := ""
	if service.GitRepo != nil {
		gitURL = service.GitRepo.URL
	}

	labels := map[string]string{
		"app":        "deployhub-build",
		"build-id":   fmt.Sprintf("%d", build.ID),
		"managed-by": "deployhub",
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy:  corev1.RestartPolicyNever,
					InitContainers: []corev1.Container{buildGitCloneContainer(gitURL, build.GitBranch, build.GitCommit, gitCredential)},
					Containers:     []corev1.Container{buildKanikoContainer(service.DockerfilePath, build.ImageTag)},
					Volumes:        buildVolumes(registryAuth),
				},
			},
		},
	}
}

// buildGitCloneContainer 构建 Git 拉取初始化容器
func buildGitCloneContainer(repoURL, branch, commit, credential string) corev1.Container {
	cloneScript := fmt.Sprintf(
		`git clone --single-branch --branch %s https://x-access-token:%s@%s /workspace && cd /workspace && git checkout %s`,
		branch, credential, stripScheme(repoURL), commit,
	)

	return corev1.Container{
		Name:    "git-clone",
		Image:   "alpine/git:latest",
		Command: []string{"sh", "-c", cloneScript},
		VolumeMounts: []corev1.VolumeMount{
			{Name: "workspace", MountPath: "/workspace"},
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		},
	}
}

// buildKanikoContainer 构建 Kaniko 执行容器
func buildKanikoContainer(dockerfilePath, imageTag string) corev1.Container {
	return corev1.Container{
		Name:  "kaniko",
		Image: "gcr.io/kaniko-project/executor:latest",
		Args: []string{
			fmt.Sprintf("--dockerfile=%s", dockerfilePath),
			"--context=dir:///workspace",
			fmt.Sprintf("--destination=%s", imageTag),
			"--cache=true",
			"--snapshot-mode=redo",
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: "workspace", MountPath: "/workspace"},
			{Name: "docker-config", MountPath: "/kaniko/.docker", ReadOnly: true},
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2"),
				corev1.ResourceMemory: resource.MustParse("4Gi"),
			},
		},
	}
}

// buildVolumes 构建 Job 所需的卷定义
func buildVolumes(registryAuth string) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "workspace",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "docker-config",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: registryAuth,
					Items: []corev1.KeyToPath{
						{Key: ".dockerconfigjson", Path: "config.json"},
					},
				},
			},
		},
	}
}

// stripScheme 去除 URL 的协议前缀
func stripScheme(url string) string {
	for _, prefix := range []string{"https://", "http://"} {
		if len(url) > len(prefix) && url[:len(prefix)] == prefix {
			return url[len(prefix):]
		}
	}
	return url
}
