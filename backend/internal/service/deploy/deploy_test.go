package deploy

import (
	"testing"

	"deployhub/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ==================== Mock 定义 ====================

type mockDeployRepo struct{ mock.Mock }

func (m *mockDeployRepo) Create(d *model.Deployment) error { return m.Called(d).Error(0) }
func (m *mockDeployRepo) FindByID(id uint) (*model.Deployment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Deployment), args.Error(1)
}
func (m *mockDeployRepo) Update(d *model.Deployment) error { return m.Called(d).Error(0) }
func (m *mockDeployRepo) List(page, pageSize int, serviceID *uint) ([]model.Deployment, int64, error) {
	args := m.Called(page, pageSize, serviceID)
	return args.Get(0).([]model.Deployment), args.Get(1).(int64), args.Error(2)
}
func (m *mockDeployRepo) FindActiveByService(serviceID uint) (*model.Deployment, error) {
	args := m.Called(serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Deployment), args.Error(1)
}
func (m *mockDeployRepo) FindLastSuccessful(serviceID uint) (*model.Deployment, error) {
	args := m.Called(serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Deployment), args.Error(1)
}
func (m *mockDeployRepo) UpdateStatus(id uint, status string) error {
	return m.Called(id, status).Error(0)
}
func (m *mockDeployRepo) UpdateStatusWithReason(id uint, status, reason string) error {
	return m.Called(id, status, reason).Error(0)
}
func (m *mockDeployRepo) UpdatePodStatus(id uint, status, podStatus, podMessage string) error {
	return m.Called(id, status, podStatus, podMessage).Error(0)
}
func (m *mockDeployRepo) UpdateField(id uint, field string, value interface{}) error {
	return m.Called(id, field, value).Error(0)
}
func (m *mockDeployRepo) Delete(id uint) error { return m.Called(id).Error(0) }

type mockServiceRepo struct{ mock.Mock }

func (m *mockServiceRepo) Create(svc *model.Service) error { return m.Called(svc).Error(0) }
func (m *mockServiceRepo) FindByID(id uint) (*model.Service, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Service), args.Error(1)
}
func (m *mockServiceRepo) FindByName(name string) (*model.Service, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Service), args.Error(1)
}
func (m *mockServiceRepo) Update(svc *model.Service) error   { return m.Called(svc).Error(0) }
func (m *mockServiceRepo) Delete(id uint) error              { return m.Called(id).Error(0) }
func (m *mockServiceRepo) BatchCreate(s []*model.Service) error { return m.Called(s).Error(0) }
func (m *mockServiceRepo) List(page, pageSize int) ([]model.Service, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]model.Service), args.Get(1).(int64), args.Error(2)
}

type mockBuildRepo struct{ mock.Mock }

func (m *mockBuildRepo) Create(b *model.Build) error { return m.Called(b).Error(0) }
func (m *mockBuildRepo) FindByID(id uint) (*model.Build, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Build), args.Error(1)
}
func (m *mockBuildRepo) Update(b *model.Build) error                              { return m.Called(b).Error(0) }
func (m *mockBuildRepo) UpdateFields(id uint, fields map[string]interface{}) error { return m.Called(id, fields).Error(0) }
func (m *mockBuildRepo) UpdateStatus(id uint, status string) error                 { return m.Called(id, status).Error(0) }
func (m *mockBuildRepo) AppendLog(id uint, logChunk string) error  { return m.Called(id, logChunk).Error(0) }
func (m *mockBuildRepo) List(page, pageSize int, serviceID *uint) ([]model.Build, int64, error) {
	args := m.Called(page, pageSize, serviceID)
	return args.Get(0).([]model.Build), args.Get(1).(int64), args.Error(2)
}
func (m *mockBuildRepo) Delete(id uint) error { return nil }

// ==================== 辅助函数 ====================

func newTestDeployService() (*DeployService, *mockDeployRepo, *mockServiceRepo, *mockBuildRepo) {
	dr := new(mockDeployRepo)
	sr := new(mockServiceRepo)
	br := new(mockBuildRepo)
	svc := NewDeployService(dr, sr, br)
	return svc, dr, sr, br
}

// ==================== 测试 ====================

func TestCreateDeployment_Success(t *testing.T) {
	svc, dr, sr, _ := newTestDeployService()

	clusterID := uint(10)
	service := &model.Service{
		ID:        1,
		ClusterID: &clusterID,
		Namespace: "production",
		Replicas:  3,
		ImageRepo: "registry.example.com/app",
	}
	sr.On("FindByID", uint(1)).Return(service, nil)

	// 无正在进行的部署
	dr.On("FindActiveByService", uint(1)).Return(nil, gorm.ErrRecordNotFound)

	// 记录上一次成功部署的镜像标签
	lastDeploy := &model.Deployment{ID: 5, ServiceID: 1, ImageTag: "v1.0.0", Status: model.DeployStatusSuccess}
	dr.On("FindLastSuccessful", uint(1)).Return(lastDeploy, nil)

	dr.On("Create", mock.AnythingOfType("*model.Deployment")).Return(nil).Run(func(args mock.Arguments) {
		d := args.Get(0).(*model.Deployment)
		d.ID = 100
	})

	dep, err := svc.CreateDeployment(1, 42, 10, "production", nil, "v2.0.0", "build", "", DeployConfig{})
	require.NoError(t, err)
	assert.Equal(t, uint(1), dep.ServiceID)
	assert.Equal(t, uint(42), dep.TriggerUserID)
	assert.Equal(t, uint(10), dep.ClusterID)
	assert.Equal(t, "production", dep.Namespace)
	assert.Equal(t, "v2.0.0", dep.ImageTag)
	assert.Equal(t, model.DeployStatusPreviewing, dep.Status)
	assert.Equal(t, "v1.0.0", dep.PreviousImageTag)
	dr.AssertExpectations(t)
	sr.AssertExpectations(t)
}

func TestCreateDeployment_WithBuild(t *testing.T) {
	svc, dr, sr, br := newTestDeployService()

	clusterID := uint(10)
	service := &model.Service{
		ID: 1, ClusterID: &clusterID, Namespace: "staging", Replicas: 2,
		ImageRepo: "registry.example.com/app",
	}
	sr.On("FindByID", uint(1)).Return(service, nil)

	buildID := uint(5)
	build := &model.Build{ID: 5, ServiceID: 1, Status: model.BuildStatusSuccess, ImageTag: "build-abc123"}
	br.On("FindByID", uint(5)).Return(build, nil)

	dr.On("FindActiveByService", uint(1)).Return(nil, gorm.ErrRecordNotFound)
	dr.On("FindLastSuccessful", uint(1)).Return(nil, gorm.ErrRecordNotFound)
	dr.On("Create", mock.AnythingOfType("*model.Deployment")).Return(nil)

	dep, err := svc.CreateDeployment(1, 42, 10, "staging", &buildID, "", "build", "", DeployConfig{})
	require.NoError(t, err)
	assert.Equal(t, "build-abc123", dep.ImageTag)
	assert.Equal(t, &buildID, dep.BuildID)
	br.AssertExpectations(t)
}

func TestCreateDeployment_BuildNotSuccess(t *testing.T) {
	svc, _, sr, br := newTestDeployService()

	clusterID := uint(10)
	service := &model.Service{ID: 1, ClusterID: &clusterID, Namespace: "staging", Replicas: 2}
	sr.On("FindByID", uint(1)).Return(service, nil)

	buildID := uint(5)
	build := &model.Build{ID: 5, ServiceID: 1, Status: model.BuildStatusFailed, ImageTag: "build-abc123"}
	br.On("FindByID", uint(5)).Return(build, nil)

	_, err := svc.CreateDeployment(1, 42, 10, "staging", &buildID, "", "build", "", DeployConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "构建未成功")
}

func TestCreateDeployment_ConcurrentLock(t *testing.T) {
	svc, dr, sr, _ := newTestDeployService()

	clusterID := uint(10)
	service := &model.Service{ID: 1, ClusterID: &clusterID, Namespace: "production", Replicas: 3}
	sr.On("FindByID", uint(1)).Return(service, nil)

	// 已有活跃部署
	active := &model.Deployment{ID: 99, ServiceID: 1, Status: model.DeployStatusDeploying}
	dr.On("FindActiveByService", uint(1)).Return(active, nil)

	_, err := svc.CreateDeployment(1, 42, 10, "production", nil, "v2.0.0", "build", "", DeployConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "存在进行中的部署")
}

func TestCreateDeployment_ServiceNotFound(t *testing.T) {
	svc, _, sr, _ := newTestDeployService()
	sr.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.CreateDeployment(999, 42, 10, "ns", nil, "v1.0.0", "build", "", DeployConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "服务不存在")
}

func TestGetByID(t *testing.T) {
	svc, dr, _, _ := newTestDeployService()

	dep := &model.Deployment{ID: 1, ServiceID: 1, Status: model.DeployStatusSuccess}
	dr.On("FindByID", uint(1)).Return(dep, nil)

	result, err := svc.GetByID(1)
	require.NoError(t, err)
	assert.Equal(t, uint(1), result.ID)
}

func TestGetByID_NotFound(t *testing.T) {
	svc, dr, _, _ := newTestDeployService()
	dr.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.GetByID(999)
	assert.Error(t, err)
}

func TestList(t *testing.T) {
	svc, dr, _, _ := newTestDeployService()

	deployments := []model.Deployment{
		{ID: 2, ServiceID: 1, Status: model.DeployStatusSuccess},
		{ID: 1, ServiceID: 1, Status: model.DeployStatusFailed},
	}
	svcID := uint(1)
	dr.On("List", 1, 20, &svcID).Return(deployments, int64(2), nil)

	result, total, err := svc.List(1, 20, &svcID)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
}

func TestUpdateStatus(t *testing.T) {
	svc, dr, _, _ := newTestDeployService()
	dr.On("UpdateStatus", uint(1), model.DeployStatusSuccess).Return(nil)

	err := svc.UpdateStatus(1, model.DeployStatusSuccess)
	assert.NoError(t, err)
	dr.AssertExpectations(t)
}

func TestStartDeploy(t *testing.T) {
	svc, dr, _, _ := newTestDeployService()

	dep := &model.Deployment{ID: 1, ServiceID: 1, Status: model.DeployStatusApproved}
	dr.On("FindByID", uint(1)).Return(dep, nil)
	dr.On("Update", mock.AnythingOfType("*model.Deployment")).Return(nil)

	err := svc.StartDeploy(1)
	require.NoError(t, err)
	dr.AssertExpectations(t)
}

func TestStartDeploy_InvalidStatus(t *testing.T) {
	svc, dr, _, _ := newTestDeployService()

	dep := &model.Deployment{ID: 1, ServiceID: 1, Status: model.DeployStatusPendingApproval}
	dr.On("FindByID", uint(1)).Return(dep, nil)

	err := svc.StartDeploy(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不允许执行")
}

func TestRollback_Success(t *testing.T) {
	svc, dr, sr, _ := newTestDeployService()

	original := &model.Deployment{
		ID:               10,
		ServiceID:        1,
		ClusterID:        5,
		Namespace:        "production",
		ImageTag:         "v2.0.0",
		PreviousImageTag: "v1.0.0",
		Status:           model.DeployStatusSuccess,
	}
	dr.On("FindByID", uint(10)).Return(original, nil)

	cid5 := uint(5)
	service := &model.Service{ID: 1, ClusterID: &cid5, Namespace: "production", Replicas: 3}
	sr.On("FindByID", uint(1)).Return(service, nil)

	dr.On("FindActiveByService", uint(1)).Return(nil, gorm.ErrRecordNotFound)
	dr.On("Create", mock.AnythingOfType("*model.Deployment")).Return(nil)

	dep, err := svc.Rollback(10, 42)
	require.NoError(t, err)
	assert.True(t, dep.IsRollback)
	assert.Equal(t, &original.ID, dep.RollbackFromID)
	assert.Equal(t, "v1.0.0", dep.ImageTag)
	assert.Equal(t, model.DeployStatusPendingApproval, dep.Status)
}

func TestRollback_NoPreviousImageTag(t *testing.T) {
	svc, dr, sr, _ := newTestDeployService()

	original := &model.Deployment{
		ID:               10,
		ServiceID:        1,
		ClusterID:        5,
		Namespace:        "production",
		ImageTag:         "v2.0.0",
		PreviousImageTag: "",
		Status:           model.DeployStatusSuccess,
	}
	dr.On("FindByID", uint(10)).Return(original, nil)

	cid5 := uint(5)
	service := &model.Service{ID: 1, ClusterID: &cid5, Namespace: "production", Replicas: 3}
	sr.On("FindByID", uint(1)).Return(service, nil)
	dr.On("FindActiveByService", uint(1)).Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.Rollback(10, 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无可回滚的镜像")
}
