package cluster

import (
	"testing"

	"deployhub/internal/model"
	"deployhub/internal/service/crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockClusterRepo struct {
	mock.Mock
}

func (m *mockClusterRepo) Create(c *model.Cluster) error { return m.Called(c).Error(0) }
func (m *mockClusterRepo) FindByID(id uint) (*model.Cluster, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Cluster), args.Error(1)
}
func (m *mockClusterRepo) FindByName(name string) (*model.Cluster, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Cluster), args.Error(1)
}
func (m *mockClusterRepo) Update(c *model.Cluster) error { return m.Called(c).Error(0) }
func (m *mockClusterRepo) Delete(id uint) error          { return m.Called(id).Error(0) }
func (m *mockClusterRepo) List(p, ps int) ([]model.Cluster, int64, error) {
	args := m.Called(p, ps)
	return args.Get(0).([]model.Cluster), args.Get(1).(int64), args.Error(2)
}

func newTestService(t *testing.T) (*ClusterService, *mockClusterRepo) {
	repo := new(mockClusterRepo)
	cryptoSvc, err := crypto.NewCryptoService("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	require.NoError(t, err)
	svc := NewClusterService(repo, cryptoSvc)
	return svc, repo
}

func TestCreateCluster(t *testing.T) {
	svc, repo := newTestService(t)
	repo.On("FindByName", "test-cluster").Return(nil, gorm.ErrRecordNotFound)
	repo.On("Create", mock.AnythingOfType("*model.Cluster")).Return(nil)

	cluster, err := svc.Create("test-cluster", "测试集群", "dev", "https://k8s:6443", "kubeconfig-content", "")
	require.NoError(t, err)
	assert.Equal(t, "test-cluster", cluster.Name)
	assert.NotEqual(t, "kubeconfig-content", cluster.KubeconfigEncrypted)
	repo.AssertExpectations(t)
}

func TestCreateClusterDuplicate(t *testing.T) {
	svc, repo := newTestService(t)
	existing := &model.Cluster{Name: "test-cluster"}
	repo.On("FindByName", "test-cluster").Return(existing, nil)

	_, err := svc.Create("test-cluster", "测试集群", "dev", "", "kubeconfig", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "已存在")
}

func TestGetCluster(t *testing.T) {
	svc, repo := newTestService(t)
	cryptoSvc, _ := crypto.NewCryptoService("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	encrypted, _ := cryptoSvc.Encrypt("kubeconfig-data")

	cluster := &model.Cluster{ID: 1, Name: "test", KubeconfigEncrypted: encrypted}
	repo.On("FindByID", uint(1)).Return(cluster, nil)

	result, err := svc.GetByID(1)
	require.NoError(t, err)
	assert.Equal(t, uint(1), result.ID)
}

func TestListClusters(t *testing.T) {
	svc, repo := newTestService(t)
	clusters := []model.Cluster{{ID: 1, Name: "cluster-1"}, {ID: 2, Name: "cluster-2"}}
	repo.On("List", 1, 20).Return(clusters, int64(2), nil)

	result, total, err := svc.List(1, 20)
	require.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
}
