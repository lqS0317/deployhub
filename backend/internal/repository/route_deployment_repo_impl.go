package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type routeDeploymentRepository struct {
	db *gorm.DB
}

// NewRouteDeploymentRepository 创建路由部署记录仓储实例
func NewRouteDeploymentRepository(db *gorm.DB) RouteDeploymentRepository {
	return &routeDeploymentRepository{db: db}
}

func (r *routeDeploymentRepository) ListByEntry(entryID uint) ([]model.RouteDeployment, error) {
	var list []model.RouteDeployment
	err := r.db.Where("route_entry_id = ?", entryID).Preload("Cluster").Order("id DESC").Find(&list).Error
	return list, err
}

func (r *routeDeploymentRepository) FindByEntryClusterNs(entryID, clusterID uint, ns string) (*model.RouteDeployment, error) {
	var rd model.RouteDeployment
	err := r.db.Where("route_entry_id = ? AND cluster_id = ? AND namespace = ?", entryID, clusterID, ns).First(&rd).Error
	if err != nil {
		return nil, err
	}
	return &rd, nil
}

func (r *routeDeploymentRepository) Upsert(rd *model.RouteDeployment) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "route_entry_id"}, {Name: "cluster_id"}, {Name: "namespace"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "config_snapshot", "rendered_yaml", "error_msg", "deployed_at"}),
	}).Create(rd).Error
}

func (r *routeDeploymentRepository) DeleteByEntry(entryID uint) error {
	return r.db.Where("route_entry_id = ?", entryID).Delete(&model.RouteDeployment{}).Error
}
