package model

import "time"

// 配置类型常量
const (
	ConfigTypeEnv            = "env"
	ConfigTypeConfigMap      = "configmap"
	ConfigTypeSecret         = "secret"
	ConfigTypeServiceAccount = "serviceaccount"
)

// 配置集格式常量
const (
	ConfigFormatProperties = "properties"
	ConfigFormatYAML       = "yaml"
	ConfigFormatJSON       = "json"
)

// ConfigEntry 配置条目（一个服务+一个集群可以有多份配置）
type ConfigEntry struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ServiceID    uint      `gorm:"not null;uniqueIndex:idx_ce_svc_cluster_name" json:"service_id"`
	ClusterID    uint      `gorm:"not null;uniqueIndex:idx_ce_svc_cluster_name" json:"cluster_id"`
	Name         string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_ce_svc_cluster_name" json:"name"`
	ConfigType   string    `gorm:"type:varchar(20);not null;default:configmap" json:"config_type"`
	Format       string    `gorm:"type:varchar(20);not null;default:properties" json:"format"`
	MountPath    string    `gorm:"type:varchar(255);default:''" json:"mount_path,omitempty"`
	DraftContent string    `gorm:"type:text;default:''" json:"draft_content,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Service *Service `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Cluster *Cluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}
