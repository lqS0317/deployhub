package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type configItemRepository struct {
	db *gorm.DB
}

// NewConfigItemRepository 创建配置项仓库
func NewConfigItemRepository(db *gorm.DB) ConfigItemRepository {
	return &configItemRepository{db: db}
}

func (r *configItemRepository) List(entryID uint) ([]model.ConfigItem, error) {
	var list []model.ConfigItem
	err := r.db.Where("config_entry_id = ? AND is_deleted = false", entryID).
		Order("id ASC").Find(&list).Error
	return list, err
}

func (r *configItemRepository) ListAll(entryID uint) ([]model.ConfigItem, error) {
	var list []model.ConfigItem
	err := r.db.Where("config_entry_id = ?", entryID).
		Order("id ASC").Find(&list).Error
	return list, err
}

func (r *configItemRepository) FindByID(id uint) (*model.ConfigItem, error) {
	var item model.ConfigItem
	err := r.db.First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *configItemRepository) FindByKey(entryID uint, key string) (*model.ConfigItem, error) {
	var item model.ConfigItem
	err := r.db.Where("config_entry_id = ? AND key = ? AND is_deleted = false", entryID, key).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *configItemRepository) Create(item *model.ConfigItem) error {
	return r.db.Create(item).Error
}

func (r *configItemRepository) Update(item *model.ConfigItem) error {
	return r.db.Save(item).Error
}

func (r *configItemRepository) SoftDelete(id uint) error {
	return r.db.Model(&model.ConfigItem{}).Where("id = ?", id).Update("is_deleted", true).Error
}

func (r *configItemRepository) PurgeDeleted(entryID uint) error {
	return r.db.Unscoped().
		Where("config_entry_id = ? AND is_deleted = true", entryID).
		Delete(&model.ConfigItem{}).Error
}
