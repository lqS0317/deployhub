package model

import "time"

// Notification 站内通知
type Notification struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	Type          string    `gorm:"type:varchar(30);not null" json:"type"`
	Title         string    `gorm:"type:varchar(200);not null" json:"title"`
	Content       string    `gorm:"type:text" json:"content,omitempty"`
	IsRead        bool      `gorm:"not null;default:false" json:"is_read"`
	ReferenceType string    `gorm:"type:varchar(30)" json:"reference_type,omitempty"`
	ReferenceID   uint      `gorm:"default:0" json:"reference_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
