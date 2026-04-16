package config

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// syncToK8s 将配置同步到 Kubernetes 集群，支持 ConfigMap 和 Secret
func (s *ConfigService) syncToK8s(clientset *kubernetes.Clientset, namespace, name, configType, content string) error {
	ctx := context.Background()

	switch configType {
	case "configmap":
		return syncConfigMap(ctx, clientset, namespace, name, content)
	case "secret":
		return syncSecret(ctx, clientset, namespace, name, content)
	default:
		return fmt.Errorf("不支持的配置类型: %s", configType)
	}
}

// syncConfigMap 创建或更新 ConfigMap
func syncConfigMap(ctx context.Context, clientset *kubernetes.Clientset, namespace, name, content string) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"managed-by": "deployhub",
			},
		},
		Data: map[string]string{
			"config": content,
		},
	}

	existing, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = clientset.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return fmt.Errorf("查询 ConfigMap 失败: %w", err)
	}

	existing.Data = cm.Data
	existing.Labels = cm.Labels
	_, err = clientset.CoreV1().ConfigMaps(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

// syncSecret 创建或更新 Secret
func syncSecret(ctx context.Context, clientset *kubernetes.Clientset, namespace, name, content string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"managed-by": "deployhub",
			},
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"config": content,
		},
	}

	existing, err := clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return fmt.Errorf("查询 Secret 失败: %w", err)
	}

	existing.StringData = secret.StringData
	existing.Labels = secret.Labels
	_, err = clientset.CoreV1().Secrets(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	return err
}
