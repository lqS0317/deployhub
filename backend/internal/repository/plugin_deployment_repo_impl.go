package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pluginDeploymentRepository struct {
	db *gorm.DB
}

// NewPluginDeploymentRepository 创建插件部署记录仓储实例
func NewPluginDeploymentRepository(db *gorm.DB) PluginDeploymentRepository {
	return &pluginDeploymentRepository{db: db}
}

func (r *pluginDeploymentRepository) ListByPlugin(pluginID uint) ([]model.PluginDeployment, error) {
	var list []model.PluginDeployment
	err := r.db.Where("plugin_id = ?", pluginID).Preload("Cluster").Order("id DESC").Find(&list).Error
	return list, err
}

func (r *pluginDeploymentRepository) FindByPluginClusterNs(pluginID, clusterID uint, ns string) (*model.PluginDeployment, error) {
	var pd model.PluginDeployment
	err := r.db.Where("plugin_id = ? AND cluster_id = ? AND namespace = ?", pluginID, clusterID, ns).First(&pd).Error
	if err != nil {
		return nil, err
	}
	return &pd, nil
}

func (r *pluginDeploymentRepository) Upsert(pd *model.PluginDeployment) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "plugin_id"}, {Name: "cluster_id"}, {Name: "namespace"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "yaml_snapshot", "error_msg", "deployed_at"}),
	}).Create(pd).Error
}

func (r *pluginDeploymentRepository) DeleteByPlugin(pluginID uint) error {
	return r.db.Where("plugin_id = ?", pluginID).Delete(&model.PluginDeployment{}).Error
}
