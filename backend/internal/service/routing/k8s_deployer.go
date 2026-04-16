package routing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"deployhub/internal/service/cluster"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

// K8sRouteDeployer 路由资源 K8s 部署器
type K8sRouteDeployer struct {
	clientPool *cluster.ClientsetPool
}

// NewK8sRouteDeployer 创建路由部署器
func NewK8sRouteDeployer(clientPool *cluster.ClientsetPool) *K8sRouteDeployer {
	return &K8sRouteDeployer{clientPool: clientPool}
}

// ApplyYAML 将 YAML 通过 server-side apply 应用到指定集群
func (d *K8sRouteDeployer) ApplyYAML(clusterID uint, yamlContent string) error {
	if yamlContent == "" {
		return fmt.Errorf("YAML 内容为空")
	}

	dynClient, mapper, err := d.buildDynamicClient(clusterID)
	if err != nil {
		return err
	}

	docs := splitDocs(yamlContent)
	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		obj, err := decodeUnstructuredObj([]byte(doc))
		if err != nil {
			return fmt.Errorf("解析第 %d 个 YAML 文档失败: %w", i+1, err)
		}

		gvk := obj.GroupVersionKind()
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return fmt.Errorf("无法映射资源类型 %s: %w", gvk.String(), err)
		}

		resource := dynClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
		obj.SetManagedFields(nil)
		_, applyErr := resource.Apply(
			context.Background(),
			obj.GetName(),
			obj,
			metav1.ApplyOptions{FieldManager: "deployhub-routing", Force: true},
		)
		if applyErr != nil {
			return fmt.Errorf("应用资源 %s/%s 失败: %w", gvk.Kind, obj.GetName(), applyErr)
		}
		log.Printf("[K8sRouteDeployer] 已应用 %s/%s (namespace=%s)", gvk.Kind, obj.GetName(), obj.GetNamespace())
	}
	return nil
}

// DeleteResource 从指定集群删除资源
func (d *K8sRouteDeployer) DeleteResource(clusterID uint, gvr schema.GroupVersionResource, name, namespace string) error {
	dynClient, _, err := d.buildDynamicClient(clusterID)
	if err != nil {
		return err
	}

	err = dynClient.Resource(gvr).Namespace(namespace).Delete(
		context.Background(),
		name,
		metav1.DeleteOptions{},
	)
	if err != nil {
		return fmt.Errorf("删除资源 %s/%s 失败: %w", gvr.Resource, name, err)
	}
	log.Printf("[K8sRouteDeployer] 已删除 %s/%s (namespace=%s)", gvr.Resource, name, namespace)
	return nil
}

func (d *K8sRouteDeployer) buildDynamicClient(clusterID uint) (dynamic.Interface, *restmapper.DeferredDiscoveryRESTMapper, error) {
	cfg, err := d.clientPool.GetRestConfig(clusterID)
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

// splitDocs 按 --- 分隔多文档 YAML
func splitDocs(raw string) []string {
	var docs []string
	var current strings.Builder
	for _, line := range strings.Split(raw, "\n") {
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

// decodeUnstructuredObj 将单个 YAML 文档解码为 Unstructured 对象
func decodeUnstructuredObj(data []byte) (*unstructured.Unstructured, error) {
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
