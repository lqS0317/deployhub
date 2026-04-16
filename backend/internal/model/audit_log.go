package model

import (
	"time"

	"gorm.io/datatypes"
)

// AuditLog 审计日志
type AuditLog struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;index:idx_audit_user" json:"user_id"`
	Action       string         `gorm:"type:varchar(50);not null" json:"action"`
	ResourceType string         `gorm:"type:varchar(30);index:idx_audit_resource" json:"resource_type,omitempty"`
	ResourceID   uint           `gorm:"index:idx_audit_resource" json:"resource_id,omitempty"`
	Detail       datatypes.JSON `gorm:"type:jsonb" json:"detail,omitempty"`
	IPAddress    string         `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	CreatedAt    time.Time      `gorm:"index:idx_audit_created" json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
