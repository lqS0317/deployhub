package repository

import "deployhub/internal/model"

// GroupRepository 组数据访问接口
type GroupRepository interface {
	Create(group *model.Group) error
	FindByID(id uint) (*model.Group, error)
	FindByName(name string) (*model.Group, error)
	Update(group *model.Group) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.Group, int64, error)
}
