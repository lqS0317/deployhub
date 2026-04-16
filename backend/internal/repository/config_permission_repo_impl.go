package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type configPermissionRepository struct {
	db *gorm.DB
}

// NewConfigPermissionRepository 创建配置权限仓库
func NewConfigPermissionRepository(db *gorm.DB) ConfigPermissionRepository {
	return &configPermissionRepository{db: db}
}

func (r *configPermissionRepository) List(serviceID uint) ([]model.ConfigPermission, error) {
	var list []model.ConfigPermission
	err := r.db.Preload("User").
		Where("service_id = ?", serviceID).
		Order("id ASC").Find(&list).Error
	return list, err
}

func (r *configPermissionRepository) FindByUserAndCluster(serviceID, clusterID, userID uint) (*model.ConfigPermission, error) {
	var perm model.ConfigPermission
	err := r.db.Where("service_id = ? AND (cluster_id = ? OR cluster_id = 0) AND user_id = ?", serviceID, clusterID, userID).
		Order("cluster_id DESC"). // 优先精确匹配（cluster_id > 0）
		First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *configPermissionRepository) Upsert(perm *model.ConfigPermission) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "service_id"}, {Name: "cluster_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role"}),
	}).Create(perm).Error
}

func (r *configPermissionRepository) Delete(id uint) error {
	return r.db.Delete(&model.ConfigPermission{}, id).Error
}
