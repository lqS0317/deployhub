package routing

import (
	"fmt"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/repository"
)

// RouteService 路由条目服务
type RouteService struct {
	entryRepo     repository.RouteEntryRepository
	deployRepo    repository.RouteDeploymentRepository
	clusterNsRepo repository.ClusterNamespaceRepository
	deployer      *K8sRouteDeployer
}

// NewRouteService 创建路由服务
func NewRouteService(
	entryRepo repository.RouteEntryRepository,
	deployRepo repository.RouteDeploymentRepository,
	clusterNsRepo repository.ClusterNamespaceRepository,
	deployer *K8sRouteDeployer,
) *RouteService {
	return &RouteService{
		entryRepo:     entryRepo,
		deployRepo:    deployRepo,
		clusterNsRepo: clusterNsRepo,
		deployer:      deployer,
	}
}

// ListEntries 列出路由条目，可按资源类型过滤
func (s *RouteService) ListEntries(resourceType string) ([]model.RouteEntry, error) {
	return s.entryRepo.List(resourceType)
}

// GetEntry 获取单个路由条目
func (s *RouteService) GetEntry(id uint) (*model.RouteEntry, error) {
	return s.entryRepo.FindByID(id)
}

// CreateEntry 创建路由条目
func (s *RouteService) CreateEntry(entry *model.RouteEntry) error {
	existing, _ := s.entryRepo.FindByNameAndType(entry.Name, entry.ResourceType)
	if existing != nil {
		return fmt.Errorf("同名同类型的路由条目已存在")
	}
	return s.entryRepo.Create(entry)
}

// UpdateEntry 更新路由条目
func (s *RouteService) UpdateEntry(entry *model.RouteEntry) error {
	_, err := s.entryRepo.FindByID(entry.ID)
	if err != nil {
		return fmt.Errorf("路由条目不存在: %w", err)
	}
	return s.entryRepo.Update(entry)
}

// DeleteEntry 删除路由条目及其部署记录
func (s *RouteService) DeleteEntry(id uint) error {
	if err := s.deployRepo.DeleteByEntry(id); err != nil {
		return fmt.Errorf("删除部署记录失败: %w", err)
	}
	return s.entryRepo.Delete(id)
}

// DeployEntry 将路由条目部署到指定集群和命名空间
func (s *RouteService) DeployEntry(entryID, clusterID uint, namespace string) error {
	if err := validateClusterNamespace(s.clusterNsRepo, clusterID, namespace); err != nil {
		return err
	}

	entry, err := s.entryRepo.FindByID(entryID)
	if err != nil {
		return fmt.Errorf("路由条目不存在: %w", err)
	}

	yamlContent, err := BuildYAML(entry.Name, namespace, entry.ResourceType, entry.Config)
	if err != nil {
		return fmt.Errorf("生成 YAML 失败: %w", err)
	}

	rd := &model.RouteDeployment{
		RouteEntryID:   entryID,
		ClusterID:      clusterID,
		Namespace:      namespace,
		ConfigSnapshot: entry.Config,
		RenderedYAML:   yamlContent,
		DeployedAt:     time.Now(),
	}

	applyErr := s.deployer.ApplyYAML(clusterID, yamlContent)
	if applyErr != nil {
		rd.Status = model.RouteDeployStatusFailed
		rd.ErrorMsg = applyErr.Error()
		_ = s.deployRepo.Upsert(rd)
		return fmt.Errorf("部署失败: %w", applyErr)
	}

	rd.Status = model.RouteDeployStatusDeployed
	rd.ErrorMsg = ""
	return s.deployRepo.Upsert(rd)
}

// PreviewEntry 预览路由条目生成的 YAML（不实际部署）
func (s *RouteService) PreviewEntry(entryID uint, namespace string) (string, error) {
	entry, err := s.entryRepo.FindByID(entryID)
	if err != nil {
		return "", fmt.Errorf("路由条目不存在: %w", err)
	}
	if namespace == "" {
		namespace = "default"
	}
	return BuildYAML(entry.Name, namespace, entry.ResourceType, entry.Config)
}

// GetDeployments 获取路由条目的所有部署记录
func (s *RouteService) GetDeployments(entryID uint) ([]model.RouteDeployment, error) {
	return s.deployRepo.ListByEntry(entryID)
}
