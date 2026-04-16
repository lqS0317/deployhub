package model

import "time"

// ConfigItem 配置项（properties 格式，归属配置条目）
type ConfigItem struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ConfigEntryID uint      `gorm:"not null;index" json:"config_entry_id"`
	Key           string    `gorm:"type:varchar(255);not null" json:"key"`
	Value         string    `gorm:"type:text;not null;default:''" json:"value"`
	Comment       string    `gorm:"type:varchar(500);default:''" json:"comment,omitempty"`
	IsDeleted     bool      `gorm:"not null;default:false" json:"is_deleted"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
