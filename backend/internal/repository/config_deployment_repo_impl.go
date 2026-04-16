package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type configDeploymentRepository struct {
	db *gorm.DB
}

// NewConfigDeploymentRepository 创建配置下发记录仓储实例
func NewConfigDeploymentRepository(db *gorm.DB) ConfigDeploymentRepository {
	return &configDeploymentRepository{db: db}
}

func (r *configDeploymentRepository) Create(dep *model.ConfigDeployment) error {
	return r.db.Create(dep).Error
}

func (r *configDeploymentRepository) FindByID(id uint) (*model.ConfigDeployment, error) {
	var dep model.ConfigDeployment
	err := r.db.First(&dep, id).Error
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

func (r *configDeploymentRepository) Update(dep *model.ConfigDeployment) error {
	return r.db.Save(dep).Error
}

func (r *configDeploymentRepository) ListByVersion(versionID uint) ([]model.ConfigDeployment, error) {
	var list []model.ConfigDeployment
	err := r.db.Where("config_version_id = ?", versionID).
		Order("id DESC").Find(&list).Error
	return list, err
}
