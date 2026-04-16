package repository

import "deployhub/internal/model"

// ConfigVersionRepository 配置版本数据访问接口
type ConfigVersionRepository interface {
	Create(ver *model.ConfigVersion) error
	FindByID(id uint) (*model.ConfigVersion, error)
	ListByTemplateAndCluster(templateID, clusterID uint) ([]model.ConfigVersion, error)
	GetMaxVersion(templateID, clusterID uint) (int, error)
}
