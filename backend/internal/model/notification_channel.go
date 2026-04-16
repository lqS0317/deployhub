package model

import "time"

// NotificationChannel 通知渠道（Webhook 配置）
type NotificationChannel struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Type       string    `gorm:"type:varchar(20);not null" json:"type"`
	WebhookURL string    `gorm:"type:varchar(500);not null" json:"webhook_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
