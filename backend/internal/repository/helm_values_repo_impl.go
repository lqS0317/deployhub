package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type helmValuesRepository struct {
	db *gorm.DB
}

func NewHelmValuesRepository(db *gorm.DB) HelmValuesRepository {
	return &helmValuesRepository{db: db}
}

func (r *helmValuesRepository) FindByServiceAndCluster(serviceID, clusterID uint) (*model.HelmValues, error) {
	var hv model.HelmValues
	err := r.db.Preload("Cluster").Where("service_id = ? AND cluster_id = ?", serviceID, clusterID).First(&hv).Error
	return &hv, err
}

// Upsert 插入或更新，更新时版本号自增
func (r *helmValuesRepository) Upsert(hv *model.HelmValues) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "service_id"}, {Name: "cluster_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"content":    hv.Content,
			"version":    gorm.Expr("helm_values.version + 1"),
			"updated_by": hv.UpdatedBy,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(hv).Error
}

func (r *helmValuesRepository) ListByService(serviceID uint) ([]model.HelmValues, error) {
	var vals []model.HelmValues
	err := r.db.Preload("Cluster").Where("service_id = ?", serviceID).Order("cluster_id ASC").Find(&vals).Error
	return vals, err
}
