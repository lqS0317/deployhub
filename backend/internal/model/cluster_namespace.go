package model

import "time"

// ClusterNamespace 集群命名空间配置
type ClusterNamespace struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClusterID uint      `gorm:"not null;uniqueIndex:idx_cluster_ns" json:"cluster_id"`
	Namespace string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_cluster_ns" json:"namespace"`
	IsDefault bool      `gorm:"not null;default:false" json:"is_default"`
	CreatedAt time.Time `json:"created_at"`

	Cluster *Cluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}
