package repository

import "deployhub/internal/model"

// ServiceMemberRepository 服务成员数据访问接口
type ServiceMemberRepository interface {
	Create(member *model.ServiceMember) error
	FindByServiceAndUser(serviceID, userID uint) (*model.ServiceMember, error)
	ListByService(serviceID uint) ([]model.ServiceMember, error)
	Update(member *model.ServiceMember) error
	Delete(id uint) error
	FindOwnersByService(serviceID uint) ([]model.ServiceMember, error)
	GetUserRole(serviceID, userID uint) (string, error)
	ListByUser(userID uint) ([]model.ServiceMember, error)
}
