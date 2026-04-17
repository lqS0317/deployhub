package build

import (
	"errors"
	"testing"

	"deployhub/internal/model"
	"deployhub/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBuildRepo 构建仓储 mock 实现
type mockBuildRepo struct {
	builds  map[uint]*model.Build
	nextID  uint
	createFn func(build *model.Build) error
}

func newMockBuildRepo() *mockBuildRepo {
	return &mockBuildRepo{builds: make(map[uint]*model.Build), nextID: 1}
}

func (m *mockBuildRepo) Create(build *model.Build) error {
	if m.createFn != nil {
		return m.createFn(build)
	}
	build.ID = m.nextID
	m.nextID++
	m.builds[build.ID] = build
	return nil
}

func (m *mockBuildRepo) FindByID(id uint) (*model.Build, error) {
	if b, ok := m.builds[id]; ok {
		return b, nil
	}
	return nil, errors.New("记录不存在")
}

func (m *mockBuildRepo) Update(build *model.Build) error {
	if _, ok := m.builds[build.ID]; !ok {
		return errors.New("记录不存在")
	}
	m.builds[build.ID] = build
	return nil
}

func (m *mockBuildRepo) List(page, pageSize int, serviceID *uint) ([]model.Build, int64, error) {
	var result []model.Build
	for _, b := range m.builds {
		if serviceID != nil && b.ServiceID != *serviceID {
			continue
		}
		result = append(result, *b)
	}
	total := int64(len(result))
	start := (page - 1) * pageSize
	if start >= len(result) {
		return nil, total, nil
	}
	end := start + pageSize
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

func (m *mockBuildRepo) UpdateStatus(id uint, status string) error {
	b, ok := m.builds[id]
	if !ok {
		return errors.New("记录不存在")
	}
	b.Status = status
	return nil
}

func (m *mockBuildRepo) UpdateFields(id uint, fields map[string]interface{}) error {
	b, ok := m.builds[id]
	if !ok {
		return errors.New("记录不存在")
	}
	if s, ok := fields["status"]; ok {
		b.Status = s.(string)
	}
	return nil
}

func (m *mockBuildRepo) AppendLog(id uint, logChunk string) error {
	b, ok := m.builds[id]
	if !ok {
		return errors.New("记录不存在")
	}
	b.Log += logChunk
	return nil
}

func (m *mockBuildRepo) Delete(id uint) error { return nil }

// mockServiceRepo 服务仓储 mock 实现
type mockServiceRepo struct {
	services map[uint]*model.Service
}

func newMockServiceRepo() *mockServiceRepo {
	return &mockServiceRepo{services: make(map[uint]*model.Service)}
}

func (m *mockServiceRepo) Create(_ *model.Service) error          { return nil }
func (m *mockServiceRepo) FindByName(_ string) (*model.Service, error) { return nil, errors.New("不存在") }
func (m *mockServiceRepo) Update(_ *model.Service) error          { return nil }
func (m *mockServiceRepo) Delete(_ uint) error                    { return nil }
func (m *mockServiceRepo) BatchCreate(_ []*model.Service) error   { return nil }
func (m *mockServiceRepo) List(_, _ int) ([]model.Service, int64, error) {
	return nil, 0, nil
}
func (m *mockServiceRepo) FindByID(id uint) (*model.Service, error) {
	if s, ok := m.services[id]; ok {
		return s, nil
	}
	return nil, errors.New("服务不存在")
}

// 确保 mock 实现了接口
var _ repository.BuildRepository = (*mockBuildRepo)(nil)
var _ repository.ServiceRepository = (*mockServiceRepo)(nil)

func TestCreateBuild(t *testing.T) {
	buildRepo := newMockBuildRepo()
	serviceRepo := newMockServiceRepo()
	serviceRepo.services[1] = &model.Service{
		ID:        1,
		Name:      "test-svc",
		ImageRepo: "registry.example.com/test-svc",
	}

	svc := NewBuildService(buildRepo, serviceRepo)

	t.Run("成功创建构建", func(t *testing.T) {
		b, err := svc.CreateBuild(1, 10, 2, "main", "abc123", "", "", "", nil, "", "")
		require.NoError(t, err)
		assert.Equal(t, uint(1), b.ServiceID)
		assert.Equal(t, uint(10), b.TriggerUserID)
		assert.Equal(t, uint(2), b.BuildClusterID)
		assert.Equal(t, "main", b.GitBranch)
		assert.Equal(t, "abc123", b.GitCommit)
		assert.Equal(t, model.BuildStatusPending, b.Status)
		assert.NotEmpty(t, b.ImageTag)
		assert.Contains(t, b.ImageTag, "main-")
	})

	t.Run("服务不存在时创建失败", func(t *testing.T) {
		_, err := svc.CreateBuild(999, 10, 2, "main", "abc123", "", "", "", nil, "", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "服务不存在")
	})
}

func TestGetByID(t *testing.T) {
	buildRepo := newMockBuildRepo()
	serviceRepo := newMockServiceRepo()
	svc := NewBuildService(buildRepo, serviceRepo)

	buildRepo.builds[1] = &model.Build{ID: 1, Status: model.BuildStatusPending}

	t.Run("获取已存在的构建", func(t *testing.T) {
		b, err := svc.GetByID(1)
		require.NoError(t, err)
		assert.Equal(t, uint(1), b.ID)
	})

	t.Run("获取不存在的构建", func(t *testing.T) {
		_, err := svc.GetByID(999)
		assert.Error(t, err)
	})
}

func TestList(t *testing.T) {
	buildRepo := newMockBuildRepo()
	serviceRepo := newMockServiceRepo()
	svc := NewBuildService(buildRepo, serviceRepo)

	buildRepo.builds[1] = &model.Build{ID: 1, ServiceID: 1, Status: model.BuildStatusPending}
	buildRepo.builds[2] = &model.Build{ID: 2, ServiceID: 2, Status: model.BuildStatusSuccess}

	t.Run("列出全部构建", func(t *testing.T) {
		builds, total, err := svc.List(1, 20, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, builds, 2)
	})

	t.Run("按服务 ID 过滤", func(t *testing.T) {
		sid := uint(1)
		builds, total, err := svc.List(1, 20, &sid)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, builds, 1)
		assert.Equal(t, uint(1), builds[0].ServiceID)
	})
}

func TestCancel(t *testing.T) {
	buildRepo := newMockBuildRepo()
	serviceRepo := newMockServiceRepo()
	svc := NewBuildService(buildRepo, serviceRepo)

	t.Run("取消 pending 状态的构建", func(t *testing.T) {
		buildRepo.builds[1] = &model.Build{ID: 1, Status: model.BuildStatusPending}
		err := svc.Cancel(1)
		require.NoError(t, err)
		assert.Equal(t, model.BuildStatusCancelled, buildRepo.builds[1].Status)
	})

	t.Run("取消 building 状态的构建", func(t *testing.T) {
		buildRepo.builds[2] = &model.Build{ID: 2, Status: model.BuildStatusBuilding}
		err := svc.Cancel(2)
		require.NoError(t, err)
		assert.Equal(t, model.BuildStatusCancelled, buildRepo.builds[2].Status)
	})

	t.Run("取消已完成的构建应失败", func(t *testing.T) {
		buildRepo.builds[3] = &model.Build{ID: 3, Status: model.BuildStatusSuccess}
		err := svc.Cancel(3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无法取消")
	})

	t.Run("取消已失败的构建应失败", func(t *testing.T) {
		buildRepo.builds[4] = &model.Build{ID: 4, Status: model.BuildStatusFailed}
		err := svc.Cancel(4)
		assert.Error(t, err)
	})

	t.Run("取消不存在的构建应失败", func(t *testing.T) {
		err := svc.Cancel(999)
		assert.Error(t, err)
	})
}

func TestGetLog(t *testing.T) {
	buildRepo := newMockBuildRepo()
	serviceRepo := newMockServiceRepo()
	svc := NewBuildService(buildRepo, serviceRepo)

	buildRepo.builds[1] = &model.Build{ID: 1, Log: "step 1\nstep 2\n"}

	t.Run("获取构建日志", func(t *testing.T) {
		log, err := svc.GetLog(1)
		require.NoError(t, err)
		assert.Equal(t, "step 1\nstep 2\n", log)
	})

	t.Run("获取不存在构建的日志", func(t *testing.T) {
		_, err := svc.GetLog(999)
		assert.Error(t, err)
	})
}

func TestUpdateStatus(t *testing.T) {
	buildRepo := newMockBuildRepo()
	serviceRepo := newMockServiceRepo()
	svc := NewBuildService(buildRepo, serviceRepo)

	buildRepo.builds[1] = &model.Build{ID: 1, Status: model.BuildStatusPending}

	t.Run("更新构建状态", func(t *testing.T) {
		err := svc.UpdateStatus(1, model.BuildStatusBuilding)
		require.NoError(t, err)
		assert.Equal(t, model.BuildStatusBuilding, buildRepo.builds[1].Status)
	})
}
