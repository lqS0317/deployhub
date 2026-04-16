package routing

import (
	"fmt"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/repository"
)

// PluginService 路由插件服务
type PluginService struct {
	pluginRepo repository.RoutePluginRepository
	deployRepo repository.PluginDeploymentRepository
	deployer   *K8sRouteDeployer
}

// NewPluginService 创建插件服务
func NewPluginService(
	pluginRepo repository.RoutePluginRepository,
	deployRepo repository.PluginDeploymentRepository,
	deployer *K8sRouteDeployer,
) *PluginService {
	return &PluginService{
		pluginRepo: pluginRepo,
		deployRepo: deployRepo,
		deployer:   deployer,
	}
}

// ListPlugins 列出所有插件
func (s *PluginService) ListPlugins() ([]model.RoutePlugin, error) {
	return s.pluginRepo.List()
}

// GetPlugin 获取单个插件
func (s *PluginService) GetPlugin(id uint) (*model.RoutePlugin, error) {
	return s.pluginRepo.FindByID(id)
}

// CreatePlugin 创建插件
func (s *PluginService) CreatePlugin(plugin *model.RoutePlugin) error {
	existing, _ := s.pluginRepo.FindByName(plugin.Name)
	if existing != nil {
		return fmt.Errorf("同名插件已存在")
	}
	return s.pluginRepo.Create(plugin)
}

// UpdatePlugin 更新插件
func (s *PluginService) UpdatePlugin(plugin *model.RoutePlugin) error {
	_, err := s.pluginRepo.FindByID(plugin.ID)
	if err != nil {
		return fmt.Errorf("插件不存在: %w", err)
	}
	return s.pluginRepo.Update(plugin)
}

// DeletePlugin 删除插件及其部署记录
func (s *PluginService) DeletePlugin(id uint) error {
	if err := s.deployRepo.DeleteByPlugin(id); err != nil {
		return fmt.Errorf("删除部署记录失败: %w", err)
	}
	return s.pluginRepo.Delete(id)
}

// DeployPlugin 将插件部署到指定集群和命名空间
func (s *PluginService) DeployPlugin(pluginID, clusterID uint, namespace string) error {
	plugin, err := s.pluginRepo.FindByID(pluginID)
	if err != nil {
		return fmt.Errorf("插件不存在: %w", err)
	}

	if plugin.YAMLContent == "" {
		return fmt.Errorf("插件 YAML 内容为空")
	}

	pd := &model.PluginDeployment{
		PluginID:     pluginID,
		ClusterID:    clusterID,
		Namespace:    namespace,
		YAMLSnapshot: plugin.YAMLContent,
		DeployedAt:   time.Now(),
	}

	applyErr := s.deployer.ApplyYAML(clusterID, plugin.YAMLContent)
	if applyErr != nil {
		pd.Status = "failed"
		pd.ErrorMsg = applyErr.Error()
		_ = s.deployRepo.Upsert(pd)
		return fmt.Errorf("部署失败: %w", applyErr)
	}

	pd.Status = "deployed"
	pd.ErrorMsg = ""
	return s.deployRepo.Upsert(pd)
}

// GetDeployments 获取插件的所有部署记录
func (s *PluginService) GetDeployments(pluginID uint) ([]model.PluginDeployment, error) {
	return s.deployRepo.ListByPlugin(pluginID)
}
