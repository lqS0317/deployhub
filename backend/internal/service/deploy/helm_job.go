package deploy

import (
	"fmt"
	"strings"

	"deployhub/internal/model"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// buildHelmUpgradeJob 生成 Helm upgrade --install 的 Runner Job spec
// cicd 仓库结构：charts/general-chart/ + services/{svc-name}/app.yaml + app-{env}.yaml
func buildHelmUpgradeJob(jobName, namespace string, deployment *model.Deployment, service *model.Service, gitURL, gitBranch, gitSecretName, valuesConfigMapName, envSuffix, serviceAccount string) *batchv1.Job {
	backoffLimit := int32(0)
	ttl := int32(3600)

	releaseName := resolveHelmReleaseName(deployment, service)
	if releaseName == "" {
		releaseName = service.Name
	}

	chartPath := resolveHelmChartPath(deployment, service)
	if chartPath == "" {
		chartPath = "charts/general-chart"
	}

	// 构造 helm upgrade 命令
	helmCmd := buildHelmUpgradeCmd(releaseName, chartPath, service, deployment, valuesConfigMapName, envSuffix)

	// git clone 命令（shallow clone 加速）
	gitCloneCmd := fmt.Sprintf("git clone --depth 1 -b %s %s /workspace", gitBranch, gitURL)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":           "deployhub-helm",
				"managed-by":    "deployhub",
				"deployment-id": fmt.Sprintf("%d", deployment.ID),
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        "deployhub-helm",
						"managed-by": "deployhub",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccount,
					RestartPolicy:      corev1.RestartPolicyNever,
					InitContainers: []corev1.Container{
						{
							Name:    "git-clone",
							Image:   "alpine/git:latest",
							Command: []string{"sh", "-c"},
							Args:    []string{gitCloneCmd},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "workspace", MountPath: "/workspace"},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "helm",
							Image:   "alpine/helm:3.16",
							Command: []string{"sh", "-c"},
							Args:    []string{helmCmd},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "workspace", MountPath: "/workspace"},
								{Name: "system-values", MountPath: "/tmp/system-values"},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "workspace",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "system-values",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: valuesConfigMapName,
									},
									Optional: boolPtr(true),
								},
							},
						},
					},
				},
			},
		},
	}
}

// buildHelmUpgradeCmd 构造 helm upgrade --install 命令
// values 优先级：chart defaults < services/{name}/app.yaml < services/{name}/app-{env}.yaml < system values < --set
func buildHelmUpgradeCmd(releaseName, chartPath string, service *model.Service, deployment *model.Deployment, valuesConfigMapName, envSuffix string) string {
	svcValuesDir := fmt.Sprintf("services/%s", service.Name)

	// 构建可选 values 文件的条件检查变量
	parts := []string{
		"set -e",
		"echo '[DeployHub] Helm 部署开始'",
		"echo '[DeployHub] Release: " + releaseName + "'",
		"echo '[DeployHub] Chart: /workspace/" + chartPath + "'",
		"echo '[DeployHub] Image Source: " + deployment.ImageSource + "'",
		"echo '[DeployHub] Namespace: " + deployment.Namespace + "'",
		"echo '---'",
		// 预检查可选 values 文件是否存在，生成动态参数
		"VALUES_ARGS=''",
		fmt.Sprintf("[ -f /workspace/%s/app.yaml ] && VALUES_ARGS=\"$VALUES_ARGS -f /workspace/%s/app.yaml\"", svcValuesDir, svcValuesDir),
	}

	if envSuffix != "" {
		envFile := fmt.Sprintf("/workspace/%s/app-%s.yaml", svcValuesDir, envSuffix)
		parts = append(parts, fmt.Sprintf("[ -f %s ] && VALUES_ARGS=\"$VALUES_ARGS -f %s\"", envFile, envFile))
	}

	if service.HelmValuesPath != "" {
		parts = append(parts, fmt.Sprintf("[ -f /workspace/%s ] && VALUES_ARGS=\"$VALUES_ARGS -f /workspace/%s\"", service.HelmValuesPath, service.HelmValuesPath))
	}

	if valuesConfigMapName != "" {
		parts = append(parts, "[ -f /tmp/system-values/values.yaml ] && VALUES_ARGS=\"$VALUES_ARGS -f /tmp/system-values/values.yaml\"")
	}

	// 构造核心 helm 命令
	helmCmd := fmt.Sprintf("helm upgrade --install %s /workspace/%s --namespace %s $VALUES_ARGS",
		releaseName, chartPath, deployment.Namespace)

	// 根据 image_source 决定镜像参数
	switch deployment.ImageSource {
	case "external":
		if deployment.ExternalImage != "" {
			repo, tag := splitImageRef(deployment.ExternalImage)
			helmCmd += " --set image.repository=" + repo + " --set image.tag=" + tag
		}
	case "env_file":
		if service.HelmEnvFilePath != "" {
			parts = append(parts, fmt.Sprintf("[ -f /workspace/%s ] && VALUES_ARGS=\"$VALUES_ARGS -f /workspace/%s\"", service.HelmEnvFilePath, service.HelmEnvFilePath))
		}
	default:
		if deployment.ImageTag != "" {
			helmCmd += " --set image.tag=" + deployment.ImageTag
		}
	}
	helmCmd += " --wait --timeout 10m"

	parts = append(parts,
		"echo \"[DeployHub] VALUES_ARGS: $VALUES_ARGS\"",
		helmCmd,
		"echo '---'",
		"echo '[DeployHub] Helm 部署完成'",
		"helm status "+releaseName+" --namespace "+deployment.Namespace,
	)

	return strings.Join(parts, "\n")
}

// buildHelmTemplateJob 生成 helm template dry-run 的 Runner Job spec（不连接集群，纯本地渲染）
func buildHelmTemplateJob(jobName, namespace string, deployment *model.Deployment, service *model.Service, gitURL, gitBranch, valuesConfigMapName, envSuffix, serviceAccount string) *batchv1.Job {
	backoffLimit := int32(0)
	ttl := int32(1800)

	releaseName := resolveHelmReleaseName(deployment, service)
	if releaseName == "" {
		releaseName = service.Name
	}

	chartPath := resolveHelmChartPath(deployment, service)
	if chartPath == "" {
		chartPath = "charts/general-chart"
	}

	// helm template 命令（纯渲染，不连接集群）
	templateCmd := buildHelmTemplateCmd(releaseName, chartPath, service, deployment, valuesConfigMapName, envSuffix)
	gitCloneCmd := fmt.Sprintf("git clone --depth 1 -b %s %s /workspace", gitBranch, gitURL)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName, Namespace: namespace,
			Labels: map[string]string{"app": "deployhub-helm-preview", "managed-by": "deployhub"},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccount,
					RestartPolicy:      corev1.RestartPolicyNever,
					InitContainers: []corev1.Container{{
						Name: "git-clone", Image: "alpine/git:latest",
						Command: []string{"sh", "-c"}, Args: []string{gitCloneCmd},
						VolumeMounts: []corev1.VolumeMount{{Name: "workspace", MountPath: "/workspace"}},
					}},
					Containers: []corev1.Container{{
						Name: "helm-template", Image: "alpine/helm:3.16",
						Command: []string{"sh", "-c"}, Args: []string{templateCmd},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "workspace", MountPath: "/workspace"},
							{Name: "system-values", MountPath: "/tmp/system-values"},
						},
					}},
					Volumes: []corev1.Volume{
						{Name: "workspace", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
						{Name: "system-values", VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: valuesConfigMapName},
								Optional:             boolPtr(true),
							},
						}},
					},
				},
			},
		},
	}
}

// buildHelmTemplateCmd 构造 helm template 命令
func buildHelmTemplateCmd(releaseName, chartPath string, service *model.Service, deployment *model.Deployment, valuesConfigMapName, envSuffix string) string {
	svcValuesDir := fmt.Sprintf("services/%s", service.Name)

	parts := []string{
		"VALUES_ARGS=''",
		fmt.Sprintf("[ -f /workspace/%s/app.yaml ] && VALUES_ARGS=\"$VALUES_ARGS -f /workspace/%s/app.yaml\"", svcValuesDir, svcValuesDir),
	}

	if envSuffix != "" {
		envFile := fmt.Sprintf("/workspace/%s/app-%s.yaml", svcValuesDir, envSuffix)
		parts = append(parts, fmt.Sprintf("[ -f %s ] && VALUES_ARGS=\"$VALUES_ARGS -f %s\"", envFile, envFile))
	}

	if service.HelmValuesPath != "" {
		parts = append(parts, fmt.Sprintf("[ -f /workspace/%s ] && VALUES_ARGS=\"$VALUES_ARGS -f /workspace/%s\"", service.HelmValuesPath, service.HelmValuesPath))
	}

	if valuesConfigMapName != "" {
		parts = append(parts, "[ -f /tmp/system-values/values.yaml ] && VALUES_ARGS=\"$VALUES_ARGS -f /tmp/system-values/values.yaml\"")
	}

	helmCmd := fmt.Sprintf("helm template %s /workspace/%s --namespace %s $VALUES_ARGS",
		releaseName, chartPath, deployment.Namespace)

	switch deployment.ImageSource {
	case "external":
		if deployment.ExternalImage != "" {
			repo, tag := splitImageRef(deployment.ExternalImage)
			helmCmd += " --set image.repository=" + repo + " --set image.tag=" + tag
		}
	case "env_file":
		if service.HelmEnvFilePath != "" {
			parts = append(parts, fmt.Sprintf("[ -f /workspace/%s ] && VALUES_ARGS=\"$VALUES_ARGS -f /workspace/%s\"", service.HelmEnvFilePath, service.HelmEnvFilePath))
		}
	default:
		if deployment.ImageTag != "" {
			helmCmd += " --set image.tag=" + deployment.ImageTag
		}
	}

	parts = append(parts, helmCmd)
	return strings.Join(parts, "\n")
}

// buildHelmRollbackJob 生成 Helm rollback 的 Runner Job spec
func buildHelmRollbackJob(jobName, namespace string, deployment *model.Deployment, service *model.Service, revision int, serviceAccount string) *batchv1.Job {
	backoffLimit := int32(0)
	ttl := int32(3600)

	releaseName := resolveHelmReleaseName(deployment, service)
	if releaseName == "" {
		releaseName = service.Name
	}

	rollbackCmd := fmt.Sprintf(
		"set -e\necho '[DeployHub] Helm 回滚到 revision %d'\nhelm rollback %s %d --namespace %s --wait --timeout 5m\necho '[DeployHub] 回滚完成'\nhelm status %s --namespace %s",
		revision, releaseName, revision, namespace, releaseName, namespace,
	)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":           "deployhub-helm",
				"managed-by":    "deployhub",
				"deployment-id": fmt.Sprintf("%d", deployment.ID),
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccount,
					RestartPolicy:      corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "helm",
							Image:   "alpine/helm:3.16",
							Command: []string{"sh", "-c"},
							Args:    []string{rollbackCmd},
						},
					},
				},
			},
		},
	}
}

// splitImageRef 拆分完整镜像地址为 repository 和 tag
func splitImageRef(ref string) (string, string) {
	if idx := strings.LastIndex(ref, ":"); idx > 0 {
		afterColon := ref[idx+1:]
		if !strings.Contains(afterColon, "/") {
			return ref[:idx], afterColon
		}
	}
	return ref, "latest"
}

func boolPtr(b bool) *bool {
	return &b
}
