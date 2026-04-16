package cluster

import (
	"fmt"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientsetPool 管理集群 clientset 缓存
type ClientsetPool struct {
	mu         sync.RWMutex
	clients    map[uint]*kubernetes.Clientset
	clusterSvc *ClusterService
}

// NewClientsetPool 创建 clientset 池
func NewClientsetPool(clusterSvc *ClusterService) *ClientsetPool {
	return &ClientsetPool{
		clients:    make(map[uint]*kubernetes.Clientset),
		clusterSvc: clusterSvc,
	}
}

// GetClientset 获取指定集群的 clientset（带缓存）
func (p *ClientsetPool) GetClientset(clusterID uint) (*kubernetes.Clientset, error) {
	p.mu.RLock()
	if cs, ok := p.clients[clusterID]; ok {
		p.mu.RUnlock()
		return cs, nil
	}
	p.mu.RUnlock()

	return p.buildAndCache(clusterID)
}

// InvalidateCache 清除指定集群的缓存
func (p *ClientsetPool) InvalidateCache(clusterID uint) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.clients, clusterID)
}

func (p *ClientsetPool) buildAndCache(clusterID uint) (*kubernetes.Clientset, error) {
	kubeconfig, err := p.clusterSvc.GetDecryptedKubeconfig(clusterID)
	if err != nil {
		return nil, fmt.Errorf("获取集群凭证失败: %w", err)
	}

	cs, err := BuildClientset(kubeconfig)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.clients[clusterID] = cs
	p.mu.Unlock()

	return cs, nil
}

// GetRestConfig 获取指定集群的 rest.Config（用于构建 dynamic client 等）
func (p *ClientsetPool) GetRestConfig(clusterID uint) (*rest.Config, error) {
	kubeconfig, err := p.clusterSvc.GetDecryptedKubeconfig(clusterID)
	if err != nil {
		return nil, fmt.Errorf("获取集群凭证失败: %w", err)
	}
	cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, fmt.Errorf("解析 kubeconfig 失败: %w", err)
	}
	return cfg, nil
}

// BuildClientset 从 kubeconfig 内容构建 clientset
func BuildClientset(kubeconfigData string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigData))
	if err != nil {
		return nil, fmt.Errorf("解析 kubeconfig 失败: %w", err)
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 Kubernetes clientset 失败: %w", err)
	}

	return cs, nil
}
