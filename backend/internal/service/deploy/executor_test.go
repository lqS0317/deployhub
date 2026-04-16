package deploy

import (
	"context"
	"testing"

	"deployhub/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestExecute_CreateNewDeployment(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	executor := &DirectExecutor{getClientset: func(clusterID uint) (KubeAppsClient, error) {
		return fakeClient.AppsV1(), nil
	}}

	deployment := &model.Deployment{
		ID:        1,
		ClusterID: 10,
		Namespace: "default",
		ImageTag:  "v1.0.0",
	}
	service := &model.Service{
		ID:        1,
		Name:      "test-app",
		ImageRepo: "registry.example.com/test-app",
		Port:      8080,
		Replicas:  3,
	}

	err := executor.Execute(deployment, service)
	require.NoError(t, err)

	// 验证 K8s Deployment 被正确创建
	k8sDep, err := fakeClient.AppsV1().Deployments("default").Get(context.Background(), "test-app", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "test-app", k8sDep.Name)
	assert.Equal(t, int32(3), *k8sDep.Spec.Replicas)
	assert.Equal(t, "registry.example.com/test-app:v1.0.0", k8sDep.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, int32(8080), k8sDep.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort)
}

func TestExecute_UpdateExistingDeployment(t *testing.T) {
	replicas := int32(2)
	existingDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "test-app", Namespace: "production"},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test-app"},
			},
		},
	}
	fakeClient := fake.NewSimpleClientset([]runtime.Object{existingDep}...)

	executor := &DirectExecutor{getClientset: func(clusterID uint) (KubeAppsClient, error) {
		return fakeClient.AppsV1(), nil
	}}

	deployment := &model.Deployment{
		ID:        2,
		ClusterID: 10,
		Namespace: "production",
		ImageTag:  "v2.0.0",
	}
	service := &model.Service{
		ID:        1,
		Name:      "test-app",
		ImageRepo: "registry.example.com/test-app",
		Port:      8080,
		Replicas:  5,
	}

	err := executor.Execute(deployment, service)
	require.NoError(t, err)

	// 验证更新后的镜像和副本数
	k8sDep, err := fakeClient.AppsV1().Deployments("production").Get(context.Background(), "test-app", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, int32(5), *k8sDep.Spec.Replicas)
	assert.Equal(t, "registry.example.com/test-app:v2.0.0", k8sDep.Spec.Template.Spec.Containers[0].Image)
}

func TestExecute_DeploymentSpec(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	executor := &DirectExecutor{getClientset: func(clusterID uint) (KubeAppsClient, error) {
		return fakeClient.AppsV1(), nil
	}}

	deployment := &model.Deployment{
		ID:        3,
		ClusterID: 10,
		Namespace: "staging",
		ImageTag:  "abc123",
	}
	service := &model.Service{
		ID:        2,
		Name:      "api-server",
		ImageRepo: "registry.example.com/api",
		Port:      3000,
		Replicas:  1,
	}

	err := executor.Execute(deployment, service)
	require.NoError(t, err)

	k8sDep, err := fakeClient.AppsV1().Deployments("staging").Get(context.Background(), "api-server", metav1.GetOptions{})
	require.NoError(t, err)

	// 验证标签选择器
	assert.Equal(t, map[string]string{"app": "api-server", "managed-by": "deployhub"}, k8sDep.Spec.Selector.MatchLabels)
	assert.Equal(t, map[string]string{"app": "api-server", "managed-by": "deployhub"}, k8sDep.Spec.Template.Labels)

	// 验证容器配置
	containers := k8sDep.Spec.Template.Spec.Containers
	require.Len(t, containers, 1)
	assert.Equal(t, "api-server", containers[0].Name)
	assert.Equal(t, "registry.example.com/api:abc123", containers[0].Image)
}

func TestExecute_CreateNewStatefulSet(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	executor := &DirectExecutor{getClientset: func(clusterID uint) (KubeAppsClient, error) {
		return fakeClient.AppsV1(), nil
	}}

	deployment := &model.Deployment{ID: 4, ClusterID: 10, Namespace: "default", ImageTag: "v1.0.0"}
	service := &model.Service{ID: 1, Name: "redis", ImageRepo: "redis", Port: 6379, WorkloadType: "statefulset", Replicas: 3}

	err := executor.Execute(deployment, service)
	require.NoError(t, err)

	sts, err := fakeClient.AppsV1().StatefulSets("default").Get(context.Background(), "redis", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "redis", sts.Name)
	assert.Equal(t, "redis", sts.Spec.ServiceName)
	assert.Equal(t, int32(3), *sts.Spec.Replicas)
	assert.Equal(t, "redis:v1.0.0", sts.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, appsv1.RollingUpdateStatefulSetStrategyType, sts.Spec.UpdateStrategy.Type)
}

func TestExecute_UpdateExistingStatefulSet(t *testing.T) {
	replicas := int32(1)
	existingSts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "redis", Namespace: "default"},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: "redis",
			Selector:    &metav1.LabelSelector{MatchLabels: map[string]string{"app": "redis"}},
		},
	}
	fakeClient := fake.NewSimpleClientset([]runtime.Object{existingSts}...)

	executor := &DirectExecutor{getClientset: func(clusterID uint) (KubeAppsClient, error) {
		return fakeClient.AppsV1(), nil
	}}

	deployment := &model.Deployment{ID: 5, ClusterID: 10, Namespace: "default", ImageTag: "v2.0.0"}
	service := &model.Service{ID: 1, Name: "redis", ImageRepo: "redis", Port: 6379, WorkloadType: "statefulset", Replicas: 3}

	err := executor.Execute(deployment, service)
	require.NoError(t, err)

	sts, err := fakeClient.AppsV1().StatefulSets("default").Get(context.Background(), "redis", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, int32(3), *sts.Spec.Replicas)
	assert.Equal(t, "redis:v2.0.0", sts.Spec.Template.Spec.Containers[0].Image)
}
