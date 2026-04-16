package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type routePluginRepository struct {
	db *gorm.DB
}

// NewRoutePluginRepository 创建路由插件仓储实例
func NewRoutePluginRepository(db *gorm.DB) RoutePluginRepository {
	return &routePluginRepository{db: db}
}

func (r *routePluginRepository) List() ([]model.RoutePlugin, error) {
	var list []model.RoutePlugin
	err := r.db.Order("id DESC").Find(&list).Error
	return list, err
}

func (r *routePluginRepository) FindByID(id uint) (*model.RoutePlugin, error) {
	var plugin model.RoutePlugin
	err := r.db.First(&plugin, id).Error
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

func (r *routePluginRepository) FindByName(name string) (*model.RoutePlugin, error) {
	var plugin model.RoutePlugin
	err := r.db.Where("name = ?", name).First(&plugin).Error
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

func (r *routePluginRepository) Create(plugin *model.RoutePlugin) error {
	return r.db.Create(plugin).Error
}

func (r *routePluginRepository) Update(plugin *model.RoutePlugin) error {
	return r.db.Save(plugin).Error
}

func (r *routePluginRepository) Delete(id uint) error {
	return r.db.Delete(&model.RoutePlugin{}, id).Error
}
