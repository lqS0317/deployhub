package cluster

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TestConnectionResult 连接测试结果
type TestConnectionResult struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	ServerVersion string `json:"server_version,omitempty"`
}

// TestConnection 测试集群连接
func TestConnection(cs *kubernetes.Clientset) *TestConnectionResult {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	version, err := cs.Discovery().ServerVersion()
	if err != nil {
		return &TestConnectionResult{
			Success: false,
			Message: fmt.Sprintf("连接失败: %s", err.Error()),
		}
	}

	// 尝试列出命名空间以验证权限
	_, err = cs.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return &TestConnectionResult{
			Success:       true,
			Message:       fmt.Sprintf("连接成功但权限受限: %s", err.Error()),
			ServerVersion: version.GitVersion,
		}
	}

	return &TestConnectionResult{
		Success:       true,
		Message:       "连接成功",
		ServerVersion: version.GitVersion,
	}
}
