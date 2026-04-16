package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type clusterNamespaceRepository struct {
	db *gorm.DB
}

func NewClusterNamespaceRepository(db *gorm.DB) ClusterNamespaceRepository {
	return &clusterNamespaceRepository{db: db}
}

func (r *clusterNamespaceRepository) Create(ns *model.ClusterNamespace) error {
	return r.db.Create(ns).Error
}

func (r *clusterNamespaceRepository) Delete(id uint) error {
	return r.db.Delete(&model.ClusterNamespace{}, id).Error
}

func (r *clusterNamespaceRepository) ListByCluster(clusterID uint) ([]model.ClusterNamespace, error) {
	var nss []model.ClusterNamespace
	err := r.db.Where("cluster_id = ?", clusterID).Order("is_default DESC, namespace ASC").Find(&nss).Error
	return nss, err
}

func (r *clusterNamespaceRepository) FindByClusterAndNamespace(clusterID uint, namespace string) (*model.ClusterNamespace, error) {
	var ns model.ClusterNamespace
	err := r.db.Where("cluster_id = ? AND namespace = ?", clusterID, namespace).First(&ns).Error
	return &ns, err
}
