package repository

import "deployhub/internal/model"

// ClusterRepository 集群数据访问接口
type ClusterRepository interface {
	Create(cluster *model.Cluster) error
	FindByID(id uint) (*model.Cluster, error)
	FindByName(name string) (*model.Cluster, error)
	Update(cluster *model.Cluster) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.Cluster, int64, error)
}
