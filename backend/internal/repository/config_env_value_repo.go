package repository

import "deployhub/internal/model"

// ConfigEnvValueRepository 配置环境变量值数据访问接口
type ConfigEnvValueRepository interface {
	CreateOrUpdate(val *model.ConfigEnvValue) error
	FindByTemplateAndCluster(templateID, clusterID uint) (*model.ConfigEnvValue, error)
	ListByTemplate(templateID uint) ([]model.ConfigEnvValue, error)
	Delete(id uint) error
}
