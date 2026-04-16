package model

import "time"

// Registry 容器镜像仓库
type Registry struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	Name                string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	URL                 string    `gorm:"type:varchar(500);not null" json:"url"`
	Provider            string    `gorm:"type:varchar(20);not null" json:"provider"`
	AuthConfigEncrypted string    `gorm:"type:text;not null" json:"-"`
	IsDefault           bool      `gorm:"not null;default:false" json:"is_default"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
