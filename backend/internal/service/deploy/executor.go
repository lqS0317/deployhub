package deploy

import (
	"context"
	"fmt"

	"deployhub/internal/model"
	"deployhub/internal/service/cluster"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

// DeployExecutor 部署执行器接口
type DeployExecutor interface {
	Execute(deployment *model.Deployment, service *model.Service) error
}

// KubeAppsClient 封装 K8s Apps/v1 客户端接口，便于测试
type KubeAppsClient interface {
	Deployments(namespace string) v1.DeploymentInterface
	StatefulSets(namespace string) v1.StatefulSetInterface
}

// DirectExecutor 直接通过 client-go 操作 K8s 工作负载的执行器
// 根据 DirectMode 路由到 YAML 执行器、配置执行器或传统执行器
type DirectExecutor struct {
	getClientset   func(clusterID uint) (KubeAppsClient, error)
	yamlExecutor   *YamlExecutor
	configExecutor *ConfigExecutor
}

// NewDirectExecutor 创建直接部署执行器
func NewDirectExecutor(clientPool *cluster.ClientsetPool) *DirectExecutor {
	return &DirectExecutor{
		getClientset: func(clusterID uint) (KubeAppsClient, error) {
			cs, err := clientPool.GetClientset(clusterID)
			if err != nil {
				return nil, err
			}
			return cs.AppsV1(), nil
		},
		yamlExecutor:   NewYamlExecutor(clientPool),
		configExecutor: NewConfigExecutor(clientPool),
	}
}

// SetConfigDeployHelper 注入配置中心部署助手
func (e *DirectExecutor) SetConfigDeployHelper(h *ConfigDeployHelper) {
	if e.configExecutor != nil {
		e.configExecutor.SetConfigDeployHelper(h)
	}
}

// Execute 使用 config executor 生成 YAML 并执行部署，fallback 到传统模式
func (e *DirectExecutor) Execute(deployment *model.Deployment, service *model.Service) error {
	if e.configExecutor != nil {
		return e.configExecutor.Execute(deployment, service)
	}
	return e.executeLegacy(deployment, service)
}

// DryRun 使用 config executor 生成 YAML 并做 dry-run，fallback 到传统模式
func (e *DirectExecutor) DryRun(deployment *model.Deployment, service *model.Service) (string, []byte, error) {
	if e.configExecutor != nil {
		previewYAML, err := e.configExecutor.DryRun(deployment, service)
		return previewYAML, deployment.PreviewSummary, err
	}
	return DryRunDirect(deployment, service, e)
}

// executeLegacy 向后兼容：DirectMode 为空时走原有的 Deployment/StatefulSet 逻辑
func (e *DirectExecutor) executeLegacy(deployment *model.Deployment, service *model.Service) error {
	client, err := e.getClientset(deployment.ClusterID)
	if err != nil {
		return fmt.Errorf("获取集群客户端失败: %w", err)
	}

	if resolveWorkloadType(deployment, service) == "statefulset" {
		return e.executeStatefulSet(client, deployment, service)
	}
	return e.executeDeployment(client, deployment, service)
}

// --- Deployment ---

func (e *DirectExecutor) executeDeployment(client KubeAppsClient, deployment *model.Deployment, service *model.Service) error {
	ctx := context.Background()
	deploymentsClient := client.Deployments(deployment.Namespace)

	existing, err := deploymentsClient.Get(ctx, service.Name, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("查询 K8s Deployment 失败: %w", err)
		}
		return e.createK8sDeployment(ctx, deploymentsClient, deployment, service)
	}
	return e.updateK8sDeployment(ctx, deploymentsClient, existing, deployment, service)
}

func (e *DirectExecutor) createK8sDeployment(ctx context.Context, client v1.DeploymentInterface, deployment *model.Deployment, service *model.Service) error {
	replicas := int32(service.Replicas)
	image := resolveImage(deployment, service)
	port := resolvePort(deployment, service)
	labels := map[string]string{"app": service.Name, "managed-by": "deployhub"}

	k8sDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: service.Name, Namespace: deployment.Namespace, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: buildPodTemplate(labels, service.Name, image, port),
		},
	}

	_, err := client.Create(ctx, k8sDep, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建 K8s Deployment 失败: %w", err)
	}
	return nil
}

func (e *DirectExecutor) updateK8sDeployment(ctx context.Context, client v1.DeploymentInterface, existing *appsv1.Deployment, deployment *model.Deployment, service *model.Service) error {
	replicas := int32(service.Replicas)
	image := resolveImage(deployment, service)
	port := resolvePort(deployment, service)
	labels := map[string]string{"app": service.Name, "managed-by": "deployhub"}

	existing.Spec.Replicas = &replicas
	existing.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	existing.Spec.Template = buildPodTemplate(labels, service.Name, image, port)

	_, err := client.Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新 K8s Deployment 失败: %w", err)
	}
	return nil
}

// --- StatefulSet ---

func (e *DirectExecutor) executeStatefulSet(client KubeAppsClient, deployment *model.Deployment, service *model.Service) error {
	ctx := context.Background()
	stsClient := client.StatefulSets(deployment.Namespace)

	existing, err := stsClient.Get(ctx, service.Name, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("查询 K8s StatefulSet 失败: %w", err)
		}
		return e.createK8sStatefulSet(ctx, stsClient, deployment, service)
	}
	return e.updateK8sStatefulSet(ctx, stsClient, existing, deployment, service)
}

func (e *DirectExecutor) createK8sStatefulSet(ctx context.Context, client v1.StatefulSetInterface, deployment *model.Deployment, service *model.Service) error {
	replicas := int32(service.Replicas)
	image := resolveImage(deployment, service)
	port := resolvePort(deployment, service)
	labels := map[string]string{"app": service.Name, "managed-by": "deployhub"}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: service.Name, Namespace: deployment.Namespace, Labels: labels},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: service.Name,
			Selector:    &metav1.LabelSelector{MatchLabels: labels},
			Template:    buildPodTemplate(labels, service.Name, image, port),
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
	}

	_, err := client.Create(ctx, sts, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建 K8s StatefulSet 失败: %w", err)
	}
	return nil
}

func (e *DirectExecutor) updateK8sStatefulSet(ctx context.Context, client v1.StatefulSetInterface, existing *appsv1.StatefulSet, deployment *model.Deployment, service *model.Service) error {
	replicas := int32(service.Replicas)
	image := resolveImage(deployment, service)
	port := resolvePort(deployment, service)
	labels := map[string]string{"app": service.Name, "managed-by": "deployhub"}

	existing.Spec.Replicas = &replicas
	existing.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	existing.Spec.Template = buildPodTemplate(labels, service.Name, image, port)

	_, err := client.Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新 K8s StatefulSet 失败: %w", err)
	}
	return nil
}

// --- 公共辅助 ---

// resolvePort 从 service 解析端口
func resolvePort(deployment *model.Deployment, service *model.Service) int {
	if service.DefaultPort > 0 {
		return service.DefaultPort
	}
	if service.Port > 0 {
		return service.Port
	}
	return 8080
}

// resolveWorkloadType 从 deployment 或 service 解析工作负载类型
func resolveWorkloadType(deployment *model.Deployment, service *model.Service) string {
	if deployment.WorkloadType != "" {
		return deployment.WorkloadType
	}
	return service.WorkloadType
}

// resolveHelmRepoID 从 deployment 或 service 解析 Helm 仓库 ID
func resolveHelmRepoID(deployment *model.Deployment, service *model.Service) *uint {
	if deployment.HelmRepoID != nil {
		return deployment.HelmRepoID
	}
	return service.HelmRepoID
}

func resolveHelmChartPath(deployment *model.Deployment, service *model.Service) string {
	if deployment.HelmChartPath != "" {
		return deployment.HelmChartPath
	}
	return service.HelmChartPath
}

func resolveHelmReleaseName(deployment *model.Deployment, service *model.Service) string {
	if deployment.HelmReleaseName != "" {
		return deployment.HelmReleaseName
	}
	return service.HelmReleaseName
}

func resolveHelmChartBranch(deployment *model.Deployment, service *model.Service) string {
	if deployment.HelmChartBranch != "" && deployment.HelmChartBranch != "main" {
		return deployment.HelmChartBranch
	}
	if service.HelmChartBranch != "" {
		return service.HelmChartBranch
	}
	return "main"
}

// resolveHelmServiceAccount 解析 Helm Job 使用的 ServiceAccount
// 优先级: deployment > cluster > "default"
func resolveHelmServiceAccount(deployment *model.Deployment, clusterSA string) string {
	if deployment.HelmServiceAccount != "" {
		return deployment.HelmServiceAccount
	}
	if clusterSA != "" {
		return clusterSA
	}
	return "default"
}

// resolveImage 根据 image_source 决定最终镜像地址
func resolveImage(deployment *model.Deployment, service *model.Service) string {
	if deployment.ImageSource == "external" && deployment.ExternalImage != "" {
		return deployment.ExternalImage
	}
	return fmt.Sprintf("%s:%s", service.ImageRepo, deployment.ImageTag)
}

func buildPodTemplate(labels map[string]string, containerName, image string, port int) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{Labels: labels},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  containerName,
					Image: image,
					Ports: []corev1.ContainerPort{
						{ContainerPort: int32(port)},
					},
				},
			},
		},
	}
}
