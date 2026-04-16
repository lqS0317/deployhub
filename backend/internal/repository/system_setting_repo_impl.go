package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type systemSettingRepository struct {
	db *gorm.DB
}

func NewSystemSettingRepository(db *gorm.DB) SystemSettingRepository {
	return &systemSettingRepository{db: db}
}

func (r *systemSettingRepository) Get(key string) (*model.SystemSetting, error) {
	var s model.SystemSetting
	err := r.db.Where("key = ?", key).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *systemSettingRepository) GetAll() ([]model.SystemSetting, error) {
	var settings []model.SystemSetting
	err := r.db.Order("key").Find(&settings).Error
	return settings, err
}

func (r *systemSettingRepository) Upsert(setting *model.SystemSetting) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "description", "updated_at"}),
	}).Create(setting).Error
}
