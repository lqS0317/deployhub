package repository

import "deployhub/internal/model"

// ConfigEntryRepository 配置条目数据访问接口
type ConfigEntryRepository interface {
	List(serviceID, clusterID uint) ([]model.ConfigEntry, error)
	FindByID(id uint) (*model.ConfigEntry, error)
	FindByName(serviceID, clusterID uint, name string) (*model.ConfigEntry, error)
	Create(entry *model.ConfigEntry) error
	Update(entry *model.ConfigEntry) error
	Delete(id uint) error
}
