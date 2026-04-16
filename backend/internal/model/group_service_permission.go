package model

import "time"

// GroupServicePermission 组的 Service 权限
type GroupServicePermission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GroupID   uint      `gorm:"not null;uniqueIndex:idx_group_service" json:"group_id"`
	ServiceID uint      `gorm:"not null;uniqueIndex:idx_group_service;index" json:"service_id"`
	Role      string    `gorm:"type:varchar(10);not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`

	Group   *Group   `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Service *Service `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
}
