package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type clusterRepository struct {
	db *gorm.DB
}

// NewClusterRepository 创建集群仓储实例
func NewClusterRepository(db *gorm.DB) ClusterRepository {
	return &clusterRepository{db: db}
}

func (r *clusterRepository) Create(cluster *model.Cluster) error {
	return r.db.Create(cluster).Error
}

func (r *clusterRepository) FindByID(id uint) (*model.Cluster, error) {
	var cluster model.Cluster
	err := r.db.First(&cluster, id).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *clusterRepository) FindByName(name string) (*model.Cluster, error) {
	var cluster model.Cluster
	err := r.db.Where("name = ?", name).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *clusterRepository) Update(cluster *model.Cluster) error {
	return r.db.Save(cluster).Error
}

func (r *clusterRepository) Delete(id uint) error {
	return r.db.Delete(&model.Cluster{}, id).Error
}

func (r *clusterRepository) List(page, pageSize int) ([]model.Cluster, int64, error) {
	var clusters []model.Cluster
	var total int64

	r.db.Model(&model.Cluster{}).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&clusters).Error
	return clusters, total, err
}
