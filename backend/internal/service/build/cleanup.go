package build

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CleanupBuildResources 清理构建完成后的 Kaniko Job 及关联 Pod
func CleanupBuildResources(clientset kubernetes.Interface, namespace, jobName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	propagation := metav1.DeletePropagationBackground
	opts := metav1.DeleteOptions{PropagationPolicy: &propagation}

	if err := clientset.BatchV1().Jobs(namespace).Delete(ctx, jobName, opts); err != nil {
		return fmt.Errorf("删除 Job %s 失败: %w", jobName, err)
	}

	return nil
}
