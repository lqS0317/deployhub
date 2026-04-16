package repository

import "deployhub/internal/model"

// RouteEntryRepository 路由条目数据访问接口
type RouteEntryRepository interface {
	List(resourceType string) ([]model.RouteEntry, error)
	FindByID(id uint) (*model.RouteEntry, error)
	FindByNameAndType(name, resourceType string) (*model.RouteEntry, error)
	Create(entry *model.RouteEntry) error
	Update(entry *model.RouteEntry) error
	Delete(id uint) error
}
