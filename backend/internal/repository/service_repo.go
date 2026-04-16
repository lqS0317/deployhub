package repository

import "deployhub/internal/model"

// ServiceRepository 服务数据访问接口
type ServiceRepository interface {
	Create(svc *model.Service) error
	FindByID(id uint) (*model.Service, error)
	FindByName(name string) (*model.Service, error)
	Update(svc *model.Service) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.Service, int64, error)
	BatchCreate(services []*model.Service) error
}
