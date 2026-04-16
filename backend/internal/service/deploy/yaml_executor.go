package deploy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"deployhub/internal/model"
	"deployhub/internal/service/cluster"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/yaml"
)

// YamlExecutor 通过 K8s dynamic client 执行 server-side apply 的执行器
type YamlExecutor struct {
	clientPool *cluster.ClientsetPool
}

// NewYamlExecutor 创建 YAML 执行器
func NewYamlExecutor(clientPool *cluster.ClientsetPool) *YamlExecutor {
	return &YamlExecutor{clientPool: clientPool}
}

// ExecuteRaw 将提供的 YAML 应用到目标集群/命名空间
func (e *YamlExecutor) ExecuteRaw(rawYAML string, deployment *model.Deployment, service *model.Service) error {
	if rawYAML == "" {
		return fmt.Errorf("YAML 内容为空")
	}
	return e.executeYAML(rawYAML, deployment)
}

// Execute 将部署的 DeployCommand 中的 YAML 应用到目标集群/命名空间
func (e *YamlExecutor) Execute(deployment *model.Deployment, service *model.Service) error {
	rawYAML := deployment.DeployCommand
	if rawYAML == "" {
		return fmt.Errorf("YAML 内容为空")
	}
	return e.executeYAML(rawYAML, deployment)
}

func (e *YamlExecutor) executeYAML(rawYAML string, deployment *model.Deployment) error {

	dynClient, mapper, err := e.buildDynamicClient(deployment.ClusterID)
	if err != nil {
		return err
	}

	docs := splitYAMLDocuments(rawYAML)
	var applied []resourceSummary

	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		obj, err := decodeUnstructured([]byte(doc))
		if err != nil {
			return fmt.Errorf("解析第 %d 个 YAML 文档失败: %w", i+1, err)
		}

		if obj.GetNamespace() == "" {
			obj.SetNamespace(deployment.Namespace)
		}

		gvk := obj.GroupVersionKind()
		gvr, err := gvkToGVR(mapper, gvk)
		if err != nil {
			return fmt.Errorf("无法映射资源类型 %s: %w", gvk.String(), err)
		}

		resource := dynClient.Resource(gvr).Namespace(obj.GetNamespace())
		obj.SetManagedFields(nil)
		_, applyErr := resource.Apply(
			context.Background(),
			obj.GetName(),
			obj,
			metav1.ApplyOptions{FieldManager: "deployhub", Force: true},
		)
		if applyErr != nil {
			return fmt.Errorf("应用资源 %s/%s 失败: %w", gvk.Kind, obj.GetName(), applyErr)
		}

		log.Printf("[YamlExecutor] 已应用 %s/%s (namespace=%s)", gvk.Kind, obj.GetName(), obj.GetNamespace())
		applied = append(applied, resourceSummary{
			Kind:   gvk.Kind,
			Name:   obj.GetName(),
			Action: "applied",
		})
	}

	summaryJSON, _ := json.Marshal(map[string]interface{}{
		"resources":   applied,
		"total_count": len(applied),
	})
	deployment.PreviewSummary = summaryJSON

	return nil
}

// DryRunRaw 校验提供的 YAML（server-side dry-run）
func (e *YamlExecutor) DryRunRaw(rawYAML string, deployment *model.Deployment, service *model.Service) (string, error) {
	if rawYAML == "" {
		return "", fmt.Errorf("YAML 内容为空")
	}
	return e.dryRunYAML(rawYAML, deployment)
}

// DryRun 校验 YAML 但不实际应用（server-side dry-run）
func (e *YamlExecutor) DryRun(deployment *model.Deployment, service *model.Service) (string, error) {
	rawYAML := deployment.DeployCommand
	if rawYAML == "" {
		return "", fmt.Errorf("YAML 内容为空")
	}
	return e.dryRunYAML(rawYAML, deployment)
}

func (e *YamlExecutor) dryRunYAML(rawYAML string, deployment *model.Deployment) (string, error) {

	dynClient, mapper, err := e.buildDynamicClient(deployment.ClusterID)
	if err != nil {
		return "", err
	}

	docs := splitYAMLDocuments(rawYAML)
	var resultParts []string
	var applied []resourceSummary

	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		obj, err := decodeUnstructured([]byte(doc))
		if err != nil {
			return "", fmt.Errorf("解析第 %d 个 YAML 文档失败: %w", i+1, err)
		}

		if obj.GetNamespace() == "" {
			obj.SetNamespace(deployment.Namespace)
		}

		gvk := obj.GroupVersionKind()
		gvr, err := gvkToGVR(mapper, gvk)
		if err != nil {
			return "", fmt.Errorf("无法映射资源类型 %s: %w", gvk.String(), err)
		}

		resource := dynClient.Resource(gvr).Namespace(obj.GetNamespace())
		obj.SetManagedFields(nil)
		result, applyErr := resource.Apply(
			context.Background(),
			obj.GetName(),
			obj,
			metav1.ApplyOptions{FieldManager: "deployhub", Force: true, DryRun: []string{metav1.DryRunAll}},
		)
		if applyErr != nil {
			return "", fmt.Errorf("dry-run 资源 %s/%s 失败: %w", gvk.Kind, obj.GetName(), applyErr)
		}

		resultYAML, _ := yaml.Marshal(result.Object)
		resultParts = append(resultParts, string(resultYAML))
		applied = append(applied, resourceSummary{
			Kind:   gvk.Kind,
			Name:   obj.GetName(),
			Action: "dry-run",
		})
	}

	summaryJSON, _ := json.Marshal(map[string]interface{}{
		"resources":   applied,
		"total_count": len(applied),
	})
	deployment.PreviewSummary = summaryJSON

	previewYAML := strings.Join(resultParts, "---\n")
	return SanitizeSecrets(previewYAML), nil
}

// --- 内部辅助 ---

type resourceSummary struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Action string `json:"action"`
}

// buildDynamicClient 根据集群 ID 构建 dynamic client 和 RESTMapper
func (e *YamlExecutor) buildDynamicClient(clusterID uint) (dynamic.Interface, *restmapper.DeferredDiscoveryRESTMapper, error) {
	cfg, err := e.clientPool.GetRestConfig(clusterID)
	if err != nil {
		return nil, nil, fmt.Errorf("获取集群配置失败: %w", err)
	}

	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("创建 dynamic client 失败: %w", err)
	}

	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("创建 discovery client 失败: %w", err)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	return dynClient, mapper, nil
}

// splitYAMLDocuments 按 --- 分隔多文档 YAML
func splitYAMLDocuments(raw string) []string {
	var docs []string
	scanner := bufio.NewScanner(strings.NewReader(raw))
	var current strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			if current.Len() > 0 {
				docs = append(docs, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteString(line)
		current.WriteString("\n")
	}
	if current.Len() > 0 {
		docs = append(docs, current.String())
	}
	return docs
}

// decodeUnstructured 将单个 YAML 文档解码为 Unstructured 对象
func decodeUnstructured(data []byte) (*unstructured.Unstructured, error) {
	reader := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	obj := &unstructured.Unstructured{}
	if err := reader.Decode(obj); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("空的 YAML 文档")
		}
		return nil, err
	}
	return obj, nil
}

// gvkToGVR 通过 RESTMapper 将 GVK 映射为 GVR
func gvkToGVR(mapper *restmapper.DeferredDiscoveryRESTMapper, gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return mapping.Resource, nil
}
