package model

import "time"

// ServiceMember 服务成员关系
type ServiceMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ServiceID uint      `gorm:"not null;uniqueIndex:idx_service_member_unique" json:"service_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_service_member_unique" json:"user_id"`
	Role      string    `gorm:"type:varchar(10);not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`

	Service *Service `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
