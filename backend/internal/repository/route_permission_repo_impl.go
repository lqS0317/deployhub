package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type routePermissionRepository struct {
	db *gorm.DB
}

// NewRoutePermissionRepository 创建路由权限仓储实例
func NewRoutePermissionRepository(db *gorm.DB) RoutePermissionRepository {
	return &routePermissionRepository{db: db}
}

func (r *routePermissionRepository) List() ([]model.RoutePermission, error) {
	var list []model.RoutePermission
	err := r.db.Preload("User").Order("id DESC").Find(&list).Error
	return list, err
}

// FindByClusterAndUser 查找用户在指定集群（或全局 cluster_id=0）的权限
func (r *routePermissionRepository) FindByClusterAndUser(clusterID, userID uint) (*model.RoutePermission, error) {
	var perm model.RoutePermission
	err := r.db.Where("(cluster_id = ? OR cluster_id = 0) AND user_id = ?", clusterID, userID).
		Order("cluster_id DESC"). // 优先匹配具体集群
		First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *routePermissionRepository) Upsert(perm *model.RoutePermission) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "cluster_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role"}),
	}).Create(perm).Error
}

func (r *routePermissionRepository) Delete(id uint) error {
	return r.db.Delete(&model.RoutePermission{}, id).Error
}
