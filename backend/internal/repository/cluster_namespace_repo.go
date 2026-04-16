package repository

import "deployhub/internal/model"

// ClusterNamespaceRepository 集群命名空间数据访问接口
type ClusterNamespaceRepository interface {
	Create(ns *model.ClusterNamespace) error
	Delete(id uint) error
	ListByCluster(clusterID uint) ([]model.ClusterNamespace, error)
	FindByClusterAndNamespace(clusterID uint, namespace string) (*model.ClusterNamespace, error)
}
