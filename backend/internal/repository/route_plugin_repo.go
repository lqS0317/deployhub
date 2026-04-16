package repository

import "deployhub/internal/model"

// RoutePluginRepository 路由插件数据访问接口
type RoutePluginRepository interface {
	List() ([]model.RoutePlugin, error)
	FindByID(id uint) (*model.RoutePlugin, error)
	FindByName(name string) (*model.RoutePlugin, error)
	Create(plugin *model.RoutePlugin) error
	Update(plugin *model.RoutePlugin) error
	Delete(id uint) error
}
