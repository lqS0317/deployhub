package model

import "time"

// RoutePlugin 路由插件（Traefik Middleware / APISIX Plugin 等 CRD 资源）
type RoutePlugin struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Description string    `gorm:"type:text;default:''" json:"description,omitempty"`
	YAMLContent string    `gorm:"type:text;default:''" json:"yaml_content"`
	CreatedByID uint      `gorm:"not null" json:"created_by_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	CreatedBy *User `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
}
