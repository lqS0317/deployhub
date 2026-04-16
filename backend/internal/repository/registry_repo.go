package repository

import "deployhub/internal/model"

// RegistryRepository 镜像仓库数据访问接口
type RegistryRepository interface {
	Create(reg *model.Registry) error
	FindByID(id uint) (*model.Registry, error)
	FindByName(name string) (*model.Registry, error)
	Update(reg *model.Registry) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.Registry, int64, error)
}
