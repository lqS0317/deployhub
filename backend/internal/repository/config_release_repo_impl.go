package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type configReleaseRepository struct {
	db *gorm.DB
}

// NewConfigReleaseRepository 创建配置发布记录仓库
func NewConfigReleaseRepository(db *gorm.DB) ConfigReleaseRepository {
	return &configReleaseRepository{db: db}
}

func (r *configReleaseRepository) List(entryID uint) ([]model.ConfigRelease, error) {
	var list []model.ConfigRelease
	err := r.db.Preload("CreatedBy").
		Where("config_entry_id = ?", entryID).
		Order("version DESC").Find(&list).Error
	return list, err
}

func (r *configReleaseRepository) FindByID(id uint) (*model.ConfigRelease, error) {
	var release model.ConfigRelease
	err := r.db.Preload("CreatedBy").First(&release, id).Error
	if err != nil {
		return nil, err
	}
	return &release, nil
}

func (r *configReleaseRepository) FindLatestPublished(entryID uint) (*model.ConfigRelease, error) {
	var release model.ConfigRelease
	err := r.db.Where("config_entry_id = ? AND status = ?", entryID, model.ReleaseStatusPublished).
		Order("version DESC").First(&release).Error
	if err != nil {
		return nil, err
	}
	return &release, nil
}

func (r *configReleaseRepository) Create(release *model.ConfigRelease) error {
	return r.db.Create(release).Error
}

func (r *configReleaseRepository) GetNextVersion(entryID uint) (int, error) {
	var maxVersion *int
	err := r.db.Model(&model.ConfigRelease{}).
		Where("config_entry_id = ?", entryID).
		Select("COALESCE(MAX(version), 0)").Scan(&maxVersion).Error
	if err != nil {
		return 0, err
	}
	return *maxVersion + 1, nil
}

func (r *configReleaseRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.ConfigRelease{}).Where("id = ?", id).Update("status", status).Error
}
