package model

import "time"

// Group 用户组
type Group struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description string    `gorm:"type:text;default:''" json:"description"`
	CreatedBy   uint      `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}
