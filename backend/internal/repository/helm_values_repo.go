package repository

import "deployhub/internal/model"

// HelmValuesRepository Helm Values 数据访问接口
type HelmValuesRepository interface {
	FindByServiceAndCluster(serviceID, clusterID uint) (*model.HelmValues, error)
	Upsert(hv *model.HelmValues) error
	ListByService(serviceID uint) ([]model.HelmValues, error)
}
