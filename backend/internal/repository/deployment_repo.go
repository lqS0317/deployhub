package repository

import "deployhub/internal/model"

// DeploymentRepository 部署记录数据访问接口
type DeploymentRepository interface {
	Create(deployment *model.Deployment) error
	FindByID(id uint) (*model.Deployment, error)
	Update(deployment *model.Deployment) error
	List(page, pageSize int, serviceID *uint) ([]model.Deployment, int64, error)
	FindActiveByService(serviceID uint) (*model.Deployment, error)
	FindLastSuccessful(serviceID uint) (*model.Deployment, error)
	UpdateStatus(id uint, status string) error
	UpdateStatusWithReason(id uint, status, reason string) error
	UpdatePodStatus(id uint, status, podStatus, podMessage string) error
	UpdateField(id uint, field string, value interface{}) error
	Delete(id uint) error
}
