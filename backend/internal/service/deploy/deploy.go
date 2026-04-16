package deploy

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/repository"

	"gorm.io/gorm"
)

// DeployService 部署业务逻辑
type DeployService struct {
	deployRepo  repository.DeploymentRepository
	serviceRepo repository.ServiceRepository
	buildRepo   repository.BuildRepository
}

// NewDeployService 创建部署服务
func NewDeployService(
	deployRepo repository.DeploymentRepository,
	serviceRepo repository.ServiceRepository,
	buildRepo repository.BuildRepository,
) *DeployService {
	return &DeployService{
		deployRepo:  deployRepo,
		serviceRepo: serviceRepo,
		buildRepo:   buildRepo,
	}
}

// DeployConfig 部署配置参数
type DeployConfig struct {
	DeployType         string
	WorkloadType       string
	HealthCheckPath    string
	HelmRepoID         *uint
	HelmChartPath      string
	HelmReleaseName    string
	HelmChartBranch    string
	HelmServiceAccount string
}

// CreateDeployment 创建部署记录
func (s *DeployService) CreateDeployment(serviceID, userID, clusterID uint, namespace string, buildID *uint, imageTag string, imageSource, externalImage string, cfg DeployConfig) (*model.Deployment, error) {
	svc, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("服务不存在")
		}
		return nil, fmt.Errorf("查询服务失败: %w", err)
	}

	// 如果指定了构建，校验构建状态并使用构建的镜像标签
	if buildID != nil {
		build, err := s.buildRepo.FindByID(*buildID)
		if err != nil {
			return nil, fmt.Errorf("构建记录不存在: %w", err)
		}
		if build.Status != model.BuildStatusSuccess {
			return nil, fmt.Errorf("构建未成功，当前状态: %s", build.Status)
		}
		imageTag = build.ImageTag
	}

	// external 模式用完整镜像地址，不要求 imageTag
	if imageSource == "external" && externalImage != "" {
		if imageTag == "" {
			// 从外部镜像地址提取 tag 部分作为记录
			if idx := strings.LastIndex(externalImage, ":"); idx > 0 {
				imageTag = externalImage[idx+1:]
			} else {
				imageTag = "latest"
			}
		}
	} else if imageSource != "env_file" && imageTag == "" {
		if cfg.DeployType == "helm" {
			imageTag = "latest"
		} else {
			return nil, fmt.Errorf("镜像标签不能为空")
		}
	}

	// 并发锁检查：同一服务不允许有多个活跃部署
	active, err := s.deployRepo.FindActiveByService(serviceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询活跃部署失败: %w", err)
	}
	if active != nil {
		return nil, fmt.Errorf("该服务存在进行中的部署 (ID: %d, 状态: %s)", active.ID, active.Status)
	}

	// 获取上一次成功部署的镜像标签，用于回滚
	var previousImageTag string
	lastDeploy, err := s.deployRepo.FindLastSuccessful(serviceID)
	if err == nil && lastDeploy != nil {
		previousImageTag = lastDeploy.ImageTag
	}

	// cluster_id 和 namespace 优先从请求参数取，兼容旧数据 fallback 到 service
	deployClusterID := clusterID
	deployNamespace := namespace
	if deployClusterID == 0 && svc.ClusterID != nil {
		deployClusterID = *svc.ClusterID
	}
	if deployNamespace == "" {
		deployNamespace = svc.Namespace
	}
	if deployClusterID == 0 {
		return nil, fmt.Errorf("必须指定目标集群")
	}

	deployment := &model.Deployment{
		ServiceID:          serviceID,
		BuildID:            buildID,
		TriggerUserID:      userID,
		ClusterID:          deployClusterID,
		Namespace:          deployNamespace,
		ImageTag:           imageTag,
		Status:             model.DeployStatusPreviewing,
		PreviousImageTag:   previousImageTag,
		ImageSource:        imageSource,
		ExternalImage:      externalImage,
		DeployType:         cfg.DeployType,
		WorkloadType:       cfg.WorkloadType,
		HealthCheckPath:    cfg.HealthCheckPath,
		HelmRepoID:         cfg.HelmRepoID,
		HelmChartPath:      cfg.HelmChartPath,
		HelmReleaseName:    cfg.HelmReleaseName,
		HelmChartBranch:    cfg.HelmChartBranch,
		HelmServiceAccount: cfg.HelmServiceAccount,
	}

	if err := s.deployRepo.Create(deployment); err != nil {
		return nil, fmt.Errorf("创建部署记录失败: %w", err)
	}

	return deployment, nil
}

// GetByID 根据 ID 获取部署详情
func (s *DeployService) GetByID(id uint) (*model.Deployment, error) {
	return s.deployRepo.FindByID(id)
}

// List 分页查询部署列表
func (s *DeployService) List(page, pageSize int, serviceID *uint) ([]model.Deployment, int64, error) {
	return s.deployRepo.List(page, pageSize, serviceID)
}

// GetServiceByID 获取服务信息
func (s *DeployService) GetServiceByID(id uint) (*model.Service, error) {
	return s.serviceRepo.FindByID(id)
}

// UpdateStatus 更新部署状态
func (s *DeployService) UpdateStatus(id uint, status string) error {
	return s.deployRepo.UpdateStatus(id, status)
}

// UpdateStatusWithReason 更新部署状态并写入失败原因
func (s *DeployService) UpdateStatusWithReason(id uint, status, reason string) error {
	return s.deployRepo.UpdateStatusWithReason(id, status, reason)
}

// UpdatePodStatus 更新部署状态和 Pod 健康检查结果
func (s *DeployService) UpdatePodStatus(id uint, status, podStatus, podMessage string) error {
	return s.deployRepo.UpdatePodStatus(id, status, podStatus, podMessage)
}

// UpdateDeploy 更新部署记录
func (s *DeployService) UpdateDeploy(dep *model.Deployment) error {
	return s.deployRepo.Update(dep)
}

// DeleteDeployment 删除部署记录（级联删除审批记录）
func (s *DeployService) DeleteDeployment(id uint) error {
	dep, err := s.deployRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("部署记录不存在: %w", err)
	}
	if dep.Status == model.DeployStatusDeploying || dep.Status == model.DeployStatusPreviewing {
		return fmt.Errorf("部署正在进行中（%s），请先取消后再删除", dep.Status)
	}
	return s.deployRepo.Delete(id)
}

// StartDeploy 将审批通过的部署转为执行中状态
func (s *DeployService) StartDeploy(id uint) error {
	dep, err := s.deployRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("部署记录不存在: %w", err)
	}

	if dep.Status != model.DeployStatusApproved {
		return fmt.Errorf("部署状态不允许执行，当前状态: %s（需要 approved）", dep.Status)
	}

	now := time.Now()
	dep.Status = model.DeployStatusDeploying
	dep.StartedAt = &now

	return s.deployRepo.Update(dep)
}

// Rollback 基于指定部署创建回滚部署
func (s *DeployService) Rollback(deploymentID, userID uint) (*model.Deployment, error) {
	original, err := s.deployRepo.FindByID(deploymentID)
	if err != nil {
		return nil, fmt.Errorf("原部署记录不存在: %w", err)
	}

	// 确定回滚目标镜像
	rollbackImageTag := original.PreviousImageTag
	if rollbackImageTag == "" {
		return nil, fmt.Errorf("无可回滚的镜像，原部署没有记录上一版本")
	}

	// 并发锁检查
	active, err := s.deployRepo.FindActiveByService(original.ServiceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询活跃部署失败: %w", err)
	}
	if active != nil {
		return nil, fmt.Errorf("该服务存在进行中的部署 (ID: %d)，无法回滚", active.ID)
	}

	// 回滚使用原部署的集群和命名空间
	rollbackDep := &model.Deployment{
		ServiceID:        original.ServiceID,
		TriggerUserID:    userID,
		ClusterID:        original.ClusterID,
		Namespace:        original.Namespace,
		ImageTag:         rollbackImageTag,
		Status:           model.DeployStatusPendingApproval,
		PreviousImageTag: original.ImageTag,
		IsRollback:       true,
		RollbackFromID:   &original.ID,
	}

	if err := s.deployRepo.Create(rollbackDep); err != nil {
		return nil, fmt.Errorf("创建回滚部署失败: %w", err)
	}

	return rollbackDep, nil
}

// StartPreview 将部署状态转为预览中
func (s *DeployService) StartPreview(id uint) error {
	dep, err := s.deployRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("部署记录不存在: %w", err)
	}
	if dep.Status != model.DeployStatusApproved && dep.Status != model.DeployStatusPendingApproval && dep.Status != model.DeployStatusPreviewing {
		return fmt.Errorf("当前状态不允许预览: %s", dep.Status)
	}
	return s.deployRepo.UpdateStatus(id, model.DeployStatusPreviewing)
}

// SavePreview 保存预览结果
func (s *DeployService) SavePreview(id uint, previewYAML string, summary []byte) error {
	dep, err := s.deployRepo.FindByID(id)
	if err != nil {
		return err
	}
	dep.PreviewYAML = &previewYAML
	dep.PreviewSummary = summary
	dep.Status = model.DeployStatusPreviewed
	return s.deployRepo.Update(dep)
}

// CancelDeploy 取消部署
func (s *DeployService) CancelDeploy(id uint) error {
	dep, err := s.deployRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("部署记录不存在: %w", err)
	}
	cancellable := map[string]bool{
		model.DeployStatusPendingApproval: true,
		model.DeployStatusApproved:        true,
		model.DeployStatusPreviewed:       true,
	}
	if !cancellable[dep.Status] {
		return fmt.Errorf("当前状态不允许取消: %s", dep.Status)
	}
	return s.deployRepo.UpdateStatus(id, model.DeployStatusCancelled)
}
