package model

import "time"

// GroupMember 组成员
type GroupMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GroupID   uint      `gorm:"not null;uniqueIndex:idx_group_user" json:"group_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_group_user;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`

	Group *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
