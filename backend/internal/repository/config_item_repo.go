package repository

import "deployhub/internal/model"

// ConfigItemRepository 配置项数据访问接口
type ConfigItemRepository interface {
	// List 列出非删除的配置项
	List(entryID uint) ([]model.ConfigItem, error)
	// ListAll 列出所有配置项（含已删除）
	ListAll(entryID uint) ([]model.ConfigItem, error)
	FindByID(id uint) (*model.ConfigItem, error)
	// FindByKey 查找非删除的配置项
	FindByKey(entryID uint, key string) (*model.ConfigItem, error)
	Create(item *model.ConfigItem) error
	Update(item *model.ConfigItem) error
	// SoftDelete 软删除（设置 is_deleted=true）
	SoftDelete(id uint) error
	// PurgeDeleted 物理删除已软删除的记录
	PurgeDeleted(entryID uint) error
}
