package model

import "time"

// ConfigVersion 配置渲染版本
type ConfigVersion struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	ConfigTemplateID uint      `gorm:"not null;uniqueIndex:idx_config_version_tpl_cluster_ver" json:"config_template_id"`
	ClusterID        uint      `gorm:"not null;uniqueIndex:idx_config_version_tpl_cluster_ver" json:"cluster_id"`
	Version          int       `gorm:"not null;uniqueIndex:idx_config_version_tpl_cluster_ver" json:"version"`
	RenderedContent  string    `gorm:"type:text;not null" json:"rendered_content"`
	CreatedByID      uint      `gorm:"not null;index" json:"created_by_id"`
	CreatedAt        time.Time `json:"created_at"`

	ConfigTemplate *ConfigTemplate `gorm:"foreignKey:ConfigTemplateID" json:"config_template,omitempty"`
	Cluster        *Cluster        `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	CreatedBy      *User           `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
}
