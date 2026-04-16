package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type configEntryRepository struct {
	db *gorm.DB
}

// NewConfigEntryRepository 创建配置条目仓库
func NewConfigEntryRepository(db *gorm.DB) ConfigEntryRepository {
	return &configEntryRepository{db: db}
}

func (r *configEntryRepository) List(serviceID, clusterID uint) ([]model.ConfigEntry, error) {
	var list []model.ConfigEntry
	err := r.db.Where("service_id = ? AND cluster_id = ?", serviceID, clusterID).
		Order("id ASC").Find(&list).Error
	return list, err
}

func (r *configEntryRepository) FindByID(id uint) (*model.ConfigEntry, error) {
	var entry model.ConfigEntry
	err := r.db.First(&entry, id).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *configEntryRepository) FindByName(serviceID, clusterID uint, name string) (*model.ConfigEntry, error) {
	var entry model.ConfigEntry
	err := r.db.Where("service_id = ? AND cluster_id = ? AND name = ?", serviceID, clusterID, name).
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *configEntryRepository) Create(entry *model.ConfigEntry) error {
	return r.db.Create(entry).Error
}

func (r *configEntryRepository) Update(entry *model.ConfigEntry) error {
	return r.db.Save(entry).Error
}

func (r *configEntryRepository) Delete(id uint) error {
	return r.db.Delete(&model.ConfigEntry{}, id).Error
}
