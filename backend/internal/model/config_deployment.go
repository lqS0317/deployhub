package model

import "time"

// ConfigDeployment 配置下发记录
type ConfigDeployment struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	ConfigVersionID uint       `gorm:"not null;index" json:"config_version_id"`
	ClusterID       uint       `gorm:"not null;index" json:"cluster_id"`
	Namespace       string     `gorm:"type:varchar(100);not null" json:"namespace"`
	ResourceName    string     `gorm:"type:varchar(200);not null" json:"resource_name"`
	Status          string     `gorm:"type:varchar(10);not null" json:"status"`
	DeployedByID    uint       `gorm:"not null;index" json:"deployed_by_id"`
	DeployedAt      *time.Time `json:"deployed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`

	ConfigVersion *ConfigVersion `gorm:"foreignKey:ConfigVersionID" json:"config_version,omitempty"`
	Cluster       *Cluster       `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	DeployedBy    *User          `gorm:"foreignKey:DeployedByID" json:"deployed_by,omitempty"`
}
