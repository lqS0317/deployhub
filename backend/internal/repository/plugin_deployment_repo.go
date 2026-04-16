package repository

import "deployhub/internal/model"

// PluginDeploymentRepository 插件部署记录数据访问接口
type PluginDeploymentRepository interface {
	ListByPlugin(pluginID uint) ([]model.PluginDeployment, error)
	FindByPluginClusterNs(pluginID, clusterID uint, ns string) (*model.PluginDeployment, error)
	Upsert(pd *model.PluginDeployment) error
	DeleteByPlugin(pluginID uint) error
}
