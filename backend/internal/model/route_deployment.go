package model

import (
	"time"

	"gorm.io/datatypes"
)

// 路由部署状态常量
const (
	RouteDeployStatusDeployed = "deployed"
	RouteDeployStatusFailed   = "failed"
)

// RouteDeployment 路由部署记录
type RouteDeployment struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	RouteEntryID   uint           `gorm:"not null;uniqueIndex:idx_rd_entry_cluster_ns" json:"route_entry_id"`
	ClusterID      uint           `gorm:"not null;uniqueIndex:idx_rd_entry_cluster_ns" json:"cluster_id"`
	Namespace      string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_rd_entry_cluster_ns;default:default" json:"namespace"`
	Status         string         `gorm:"type:varchar(20);not null;default:deployed" json:"status"`
	ConfigSnapshot datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"config_snapshot,omitempty"`
	RenderedYAML   string         `gorm:"type:text;default:''" json:"rendered_yaml,omitempty"`
	ErrorMsg       string         `gorm:"type:text;default:''" json:"error_msg,omitempty"`
	DeployedAt     time.Time      `json:"deployed_at"`

	Cluster *Cluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}
