package repository

import "deployhub/internal/model"

// RouteDeploymentRepository 路由部署记录数据访问接口
type RouteDeploymentRepository interface {
	ListByEntry(entryID uint) ([]model.RouteDeployment, error)
	FindByEntryClusterNs(entryID, clusterID uint, ns string) (*model.RouteDeployment, error)
	Upsert(rd *model.RouteDeployment) error
	DeleteByEntry(entryID uint) error
}
