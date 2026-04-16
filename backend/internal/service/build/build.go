package build

import (
	"fmt"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/repository"
)

// BuildService 构建管理服务
type BuildService struct {
	buildRepo   repository.BuildRepository
	serviceRepo repository.ServiceRepository
}

// NewBuildService 创建构建服务
func NewBuildService(buildRepo repository.BuildRepository, serviceRepo repository.ServiceRepository) *BuildService {
	return &BuildService{
		buildRepo:   buildRepo,
		serviceRepo: serviceRepo,
	}
}

// CreateBuild 创建构建记录，状态初始为 pending
// buildClusterID 为 0 时自动使用服务关联的 cluster_id
// imageTag 为空时自动生成
func (s *BuildService) CreateBuild(serviceID, userID, buildClusterID uint, gitBranch, gitCommit, imageTag, name, dockerfilePath string, registryID *uint, imageRepo, buildContext string) (*model.Build, error) {
	svc, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		return nil, fmt.Errorf("服务不存在: %w", err)
	}

	if buildClusterID == 0 {
		if svc.ClusterID != nil {
			buildClusterID = *svc.ClusterID
		} else {
			return nil, fmt.Errorf("未指定构建集群且服务未绑定集群")
		}
	}

	// 构建配置：优先使用请求参数，fallback 到 Service 值
	effectiveImageRepo := imageRepo
	if effectiveImageRepo == "" {
		effectiveImageRepo = svc.ImageRepo
	}
	effectiveDockerfile := dockerfilePath
	if effectiveDockerfile == "" {
		effectiveDockerfile = svc.DockerfilePath
	}
	effectiveRegistryID := registryID
	if effectiveRegistryID == nil {
		effectiveRegistryID = svc.RegistryID
	}
	if buildContext == "" {
		buildContext = "."
	}

	if imageTag == "" {
		imageTag = fmt.Sprintf("%s:%s-%d", effectiveImageRepo, gitBranch, time.Now().Unix())
	}

	build := &model.Build{
		ServiceID:      serviceID,
		TriggerUserID:  userID,
		BuildClusterID: buildClusterID,
		GitBranch:      gitBranch,
		GitCommit:      gitCommit,
		ImageTag:       imageTag,
		Status:         model.BuildStatusPending,
		Name:           name,
		DockerfilePath: effectiveDockerfile,
		RegistryID:     effectiveRegistryID,
		ImageRepo:      effectiveImageRepo,
		BuildContext:   buildContext,
	}

	if err := s.buildRepo.Create(build); err != nil {
		return nil, fmt.Errorf("创建构建记录失败: %w", err)
	}

	return build, nil
}

// GetByID 根据 ID 获取构建详情
func (s *BuildService) GetByID(id uint) (*model.Build, error) {
	return s.buildRepo.FindByID(id)
}

// Delete 删除构建记录
func (s *BuildService) Delete(id uint) error {
	return s.buildRepo.Delete(id)
}

// List 分页列出构建记录，可按 serviceID 过滤
func (s *BuildService) List(page, pageSize int, serviceID *uint) ([]model.Build, int64, error) {
	return s.buildRepo.List(page, pageSize, serviceID)
}

// Cancel 取消构建，仅 pending/building 状态可取消
func (s *BuildService) Cancel(id uint) error {
	build, err := s.buildRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("构建记录不存在: %w", err)
	}

	if build.Status != model.BuildStatusPending && build.Status != model.BuildStatusBuilding {
		return fmt.Errorf("无法取消状态为 %s 的构建", build.Status)
	}

	return s.buildRepo.UpdateStatus(id, model.BuildStatusCancelled)
}

// GetLog 获取构建日志文本
func (s *BuildService) GetLog(id uint) (string, error) {
	build, err := s.buildRepo.FindByID(id)
	if err != nil {
		return "", fmt.Errorf("构建记录不存在: %w", err)
	}
	return build.Log, nil
}

// UpdateStatus 更新构建状态
func (s *BuildService) UpdateStatus(id uint, status string) error {
	return s.buildRepo.UpdateStatus(id, status)
}

// AppendLog 追加构建日志
func (s *BuildService) AppendLog(id uint, logChunk string) error {
	return s.buildRepo.AppendLog(id, logChunk)
}
