package model

import "time"

// ConfigTemplate 配置模板
type ConfigTemplate struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ServiceID       uint      `gorm:"not null;uniqueIndex:idx_config_template_service_name" json:"service_id"`
	Name            string    `gorm:"type:varchar(100);not null;uniqueIndex:idx_config_template_service_name" json:"name"`
	ConfigType      string    `gorm:"type:varchar(10);not null" json:"config_type"`
	TemplateContent string    `gorm:"type:text;not null" json:"template_content"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	Service *Service `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
}
