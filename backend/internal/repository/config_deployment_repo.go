package repository

import "deployhub/internal/model"

// ConfigDeploymentRepository 配置下发记录数据访问接口
type ConfigDeploymentRepository interface {
	Create(dep *model.ConfigDeployment) error
	FindByID(id uint) (*model.ConfigDeployment, error)
	Update(dep *model.ConfigDeployment) error
	ListByVersion(versionID uint) ([]model.ConfigDeployment, error)
}
