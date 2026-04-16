package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type configVersionRepository struct {
	db *gorm.DB
}

// NewConfigVersionRepository 创建配置版本仓储实例
func NewConfigVersionRepository(db *gorm.DB) ConfigVersionRepository {
	return &configVersionRepository{db: db}
}

func (r *configVersionRepository) Create(ver *model.ConfigVersion) error {
	return r.db.Create(ver).Error
}

func (r *configVersionRepository) FindByID(id uint) (*model.ConfigVersion, error) {
	var ver model.ConfigVersion
	err := r.db.First(&ver, id).Error
	if err != nil {
		return nil, err
	}
	return &ver, nil
}

func (r *configVersionRepository) ListByTemplateAndCluster(templateID, clusterID uint) ([]model.ConfigVersion, error) {
	var list []model.ConfigVersion
	err := r.db.Where("config_template_id = ? AND cluster_id = ?", templateID, clusterID).
		Order("version DESC").Find(&list).Error
	return list, err
}

// GetMaxVersion 获取指定模板+集群的最大版本号，无记录时返回 0
func (r *configVersionRepository) GetMaxVersion(templateID, clusterID uint) (int, error) {
	var maxVer int
	err := r.db.Model(&model.ConfigVersion{}).
		Where("config_template_id = ? AND cluster_id = ?", templateID, clusterID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVer).Error
	return maxVer, err
}
