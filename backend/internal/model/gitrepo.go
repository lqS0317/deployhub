package model

import "time"

// GitRepo Git 仓库
type GitRepo struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	Name                string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	URL                 string    `gorm:"type:varchar(500);not null" json:"url"`
	Provider            string    `gorm:"type:varchar(20);not null" json:"provider"`
	AuthType            string    `gorm:"type:varchar(10);not null" json:"auth_type"`
	CredentialEncrypted string    `gorm:"type:text;not null" json:"-"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
