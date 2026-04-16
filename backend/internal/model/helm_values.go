package model

import "time"

// HelmValues 服务在特定集群的 Helm values 配置
type HelmValues struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ServiceID uint      `gorm:"not null;uniqueIndex:idx_helm_values_svc_cluster" json:"service_id"`
	ClusterID uint      `gorm:"not null;uniqueIndex:idx_helm_values_svc_cluster" json:"cluster_id"`
	Content   string    `gorm:"type:text;not null;default:''" json:"content"`
	Version   int       `gorm:"not null;default:1" json:"version"`
	UpdatedBy *uint     `json:"updated_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Service   *Service `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Cluster   *Cluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	UpdatedUser *User  `gorm:"foreignKey:UpdatedBy" json:"updated_user,omitempty"`
}
