package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type configEnvValueRepository struct {
	db *gorm.DB
}

// NewConfigEnvValueRepository 创建配置环境变量值仓储实例
func NewConfigEnvValueRepository(db *gorm.DB) ConfigEnvValueRepository {
	return &configEnvValueRepository{db: db}
}

// CreateOrUpdate 创建或更新环境变量值（按模板+集群唯一约束）
func (r *configEnvValueRepository) CreateOrUpdate(val *model.ConfigEnvValue) error {
	var existing model.ConfigEnvValue
	err := r.db.Where("config_template_id = ? AND cluster_id = ?",
		val.ConfigTemplateID, val.ClusterID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.Create(val).Error
	}
	if err != nil {
		return err
	}

	existing.ValuesEncrypted = val.ValuesEncrypted
	return r.db.Save(&existing).Error
}

func (r *configEnvValueRepository) FindByTemplateAndCluster(templateID, clusterID uint) (*model.ConfigEnvValue, error) {
	var val model.ConfigEnvValue
	err := r.db.Where("config_template_id = ? AND cluster_id = ?", templateID, clusterID).First(&val).Error
	if err != nil {
		return nil, err
	}
	return &val, nil
}

func (r *configEnvValueRepository) ListByTemplate(templateID uint) ([]model.ConfigEnvValue, error) {
	var list []model.ConfigEnvValue
	err := r.db.Where("config_template_id = ?", templateID).
		Preload("Cluster").
		Order("cluster_id ASC").Find(&list).Error
	return list, err
}

func (r *configEnvValueRepository) Delete(id uint) error {
	return r.db.Delete(&model.ConfigEnvValue{}, id).Error
}
