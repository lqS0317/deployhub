package model

import "time"

// ConfigEnvValue 配置环境变量值（按集群加密存储）
type ConfigEnvValue struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	ConfigTemplateID uint      `gorm:"not null;uniqueIndex:idx_config_env_tpl_cluster" json:"config_template_id"`
	ClusterID        uint      `gorm:"not null;uniqueIndex:idx_config_env_tpl_cluster" json:"cluster_id"`
	ValuesEncrypted  string    `gorm:"type:text;not null" json:"-"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	ConfigTemplate *ConfigTemplate `gorm:"foreignKey:ConfigTemplateID" json:"config_template,omitempty"`
	Cluster        *Cluster        `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}
