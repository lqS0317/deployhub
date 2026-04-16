package model

import "time"

// PluginDeployment 插件部署记录
type PluginDeployment struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	PluginID     uint      `gorm:"not null;uniqueIndex:idx_pd_plugin_cluster_ns" json:"plugin_id"`
	ClusterID    uint      `gorm:"not null;uniqueIndex:idx_pd_plugin_cluster_ns" json:"cluster_id"`
	Namespace    string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_pd_plugin_cluster_ns;default:default" json:"namespace"`
	Status       string    `gorm:"type:varchar(20);not null;default:deployed" json:"status"`
	YAMLSnapshot string    `gorm:"type:text;default:''" json:"yaml_snapshot,omitempty"`
	ErrorMsg     string    `gorm:"type:text;default:''" json:"error_msg,omitempty"`
	DeployedAt   time.Time `json:"deployed_at"`

	Cluster *Cluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}
