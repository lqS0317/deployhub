package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"deployhub/internal/model"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// DryRunDirect 传统模式的 K8s server-side dry-run（Deployment/StatefulSet）
func DryRunDirect(deployment *model.Deployment, service *model.Service, e *DirectExecutor) (string, []byte, error) {
	client, err := e.getClientset(deployment.ClusterID)
	if err != nil {
		return "", nil, fmt.Errorf("获取集群客户端失败: %w", err)
	}

	if service.WorkloadType == "statefulset" {
		return e.dryRunStatefulSet(client, deployment, service)
	}
	return e.dryRunDeployment(client, deployment, service)
}

func (e *DirectExecutor) dryRunDeployment(client KubeAppsClient, deployment *model.Deployment, service *model.Service) (string, []byte, error) {
	ctx := context.Background()
	deploymentsClient := client.Deployments(deployment.Namespace)
	image := resolveImage(deployment, service)
	port := resolvePort(deployment, service)
	replicas := int32(service.Replicas)
	labels := map[string]string{"app": service.Name, "managed-by": "deployhub"}

	k8sDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: service.Name, Namespace: deployment.Namespace, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: buildPodTemplate(labels, service.Name, image, port),
		},
	}

	// 检查是否已存在
	existing, getErr := deploymentsClient.Get(ctx, service.Name, metav1.GetOptions{})
	var action string

	if getErr != nil && errors.IsNotFound(getErr) {
		// server-side dry-run 创建
		result, err := deploymentsClient.Create(ctx, k8sDep, metav1.CreateOptions{
			DryRun: []string{metav1.DryRunAll},
		})
		if err != nil {
			return "", nil, fmt.Errorf("dry-run 创建失败: %w", err)
		}
		action = "create"
		return buildPreviewResult(result, "Deployment", action, deployment)
	} else if getErr != nil {
		return "", nil, fmt.Errorf("查询 K8s Deployment 失败: %w", getErr)
	}

	// server-side dry-run 更新
	existing.Spec.Replicas = &replicas
	existing.Spec.Template = buildPodTemplate(labels, service.Name, image, port)
	result, err := deploymentsClient.Update(ctx, existing, metav1.UpdateOptions{
		DryRun: []string{metav1.DryRunAll},
	})
	if err != nil {
		return "", nil, fmt.Errorf("dry-run 更新失败: %w", err)
	}
	action = "update"
	return buildPreviewResult(result, "Deployment", action, deployment)
}

func (e *DirectExecutor) dryRunStatefulSet(client KubeAppsClient, deployment *model.Deployment, service *model.Service) (string, []byte, error) {
	ctx := context.Background()
	stsClient := client.StatefulSets(deployment.Namespace)
	image := resolveImage(deployment, service)
	port := resolvePort(deployment, service)
	replicas := int32(service.Replicas)
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

	existing, getErr := stsClient.Get(ctx, service.Name, metav1.GetOptions{})
	var action string

	if getErr != nil && errors.IsNotFound(getErr) {
		result, err := stsClient.Create(ctx, sts, metav1.CreateOptions{
			DryRun: []string{metav1.DryRunAll},
		})
		if err != nil {
			return "", nil, fmt.Errorf("dry-run 创建 StatefulSet 失败: %w", err)
		}
		action = "create"
		return buildPreviewResult(result, "StatefulSet", action, deployment)
	} else if getErr != nil {
		return "", nil, fmt.Errorf("查询 K8s StatefulSet 失败: %w", getErr)
	}

	existing.Spec.Replicas = &replicas
	existing.Spec.Template = buildPodTemplate(labels, service.Name, image, port)
	result, err := stsClient.Update(ctx, existing, metav1.UpdateOptions{
		DryRun: []string{metav1.DryRunAll},
	})
	if err != nil {
		return "", nil, fmt.Errorf("dry-run 更新 StatefulSet 失败: %w", err)
	}
	action = "update"
	return buildPreviewResult(result, "StatefulSet", action, deployment)
}

// buildPreviewResult 将 dry-run 结果序列化为 YAML + 摘要
func buildPreviewResult(obj interface{}, kind, action string, deployment *model.Deployment) (string, []byte, error) {
	yamlBytes, err := yaml.Marshal(obj)
	if err != nil {
		return "", nil, fmt.Errorf("序列化 YAML 失败: %w", err)
	}

	previewYAML := SanitizeSecrets(string(yamlBytes))

	summary := map[string]interface{}{
		"resources": []map[string]string{
			{"kind": kind, "name": deployment.Namespace + "/" + kind, "action": action},
		},
		"changes": map[string]string{
			"image_tag": deployment.ImageTag,
		},
	}
	summaryJSON, _ := json.Marshal(summary)

	return previewYAML, summaryJSON, nil
}

// SanitizeSecrets 脱敏 YAML 中的 Secret data 字段
func SanitizeSecrets(yamlContent string) string {
	lines := strings.Split(yamlContent, "\n")
	inData := false
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "data:" || trimmed == "stringData:" {
			inData = true
			result = append(result, line)
			continue
		}

		if inData {
			if strings.HasPrefix(trimmed, "-") || (len(trimmed) > 0 && trimmed[0] != ' ' && !strings.HasPrefix(line, "  ")) {
				inData = false
				result = append(result, line)
			} else if strings.Contains(trimmed, ":") {
				key := strings.SplitN(trimmed, ":", 2)[0]
				indent := strings.Repeat(" ", len(line)-len(strings.TrimLeft(line, " ")))
				result = append(result, indent+key+": \"***\"")
			} else {
				result = append(result, line)
			}
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
