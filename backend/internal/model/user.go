package model

import "time"

// User 系统用户
type User struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	Username      string  `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email         string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	PasswordHash  *string `gorm:"type:varchar(255)" json:"-"`
	OAuthProvider *string `gorm:"column:oauth_provider;type:varchar(20)" json:"oauth_provider,omitempty"`
	OAuthID       *string `gorm:"column:oauth_id;type:varchar(100)" json:"-"`
	Role           string  `gorm:"type:varchar(10);not null;default:member" json:"role"`
	Avatar         string  `gorm:"type:varchar(500)" json:"avatar,omitempty"`
	Nickname       string  `gorm:"type:varchar(100);default:''" json:"nickname,omitempty"`
	PhoneEncrypted string  `gorm:"type:text;default:''" json:"-"`
	Status         string  `gorm:"type:varchar(10);not null;default:active" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
