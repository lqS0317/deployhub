package model

import "time"

// 通知事件类型常量
const (
	EventAll               = "all" // 全局默认渠道（匹配所有事件）
	EventBuildSuccess      = "build_success"
	EventBuildFailed       = "build_failed"
	EventBuildCancelled    = "build_cancelled"
	EventDeploySuccess     = "deploy_success"
	EventDeployFailed      = "deploy_failed"
	EventApprovalPending   = "approval_pending"
	EventPodUnhealthy      = "pod_unhealthy"
	EventRollbackTriggered = "rollback_triggered"
	EventDeployCancelled   = "deploy_cancelled"
)

// AllEventTypes 所有支持的具体事件类型（不含 all）
var AllEventTypes = []string{
	EventBuildSuccess, EventBuildFailed, EventBuildCancelled,
	EventDeploySuccess, EventDeployFailed, EventDeployCancelled,
	EventApprovalPending, EventPodUnhealthy, EventRollbackTriggered,
}

// NotificationRule 全局通知规则
// event_type = "all" 表示默认渠道，具体事件类型的规则优先级更高
type NotificationRule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChannelID uint      `gorm:"not null;uniqueIndex:idx_nr_channel_event" json:"channel_id"`
	EventType string    `gorm:"type:varchar(30);not null;uniqueIndex:idx_nr_channel_event" json:"event_type"`
	Enabled   bool      `gorm:"not null;default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Channel *NotificationChannel `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
}

// ServiceNotificationRule 服务级通知规则
// 一个服务一条记录一个渠道，该服务所有事件都走这个渠道
type ServiceNotificationRule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ServiceID uint      `gorm:"not null;uniqueIndex:idx_snr_service" json:"service_id"`
	ChannelID uint      `gorm:"not null" json:"channel_id"`
	Enabled   bool      `gorm:"not null;default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Channel *NotificationChannel `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
	Service *Service             `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
}

// NotificationLog 通知发送记录
type NotificationLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ServiceID uint      `gorm:"not null;index" json:"service_id"`
	ChannelID uint      `gorm:"not null" json:"channel_id"`
	EventType string    `gorm:"type:varchar(30);not null;index" json:"event_type"`
	Title     string    `gorm:"type:varchar(255);not null;default:''" json:"title"`
	Content   string    `gorm:"type:text;not null;default:''" json:"content"`
	Status    string    `gorm:"type:varchar(10);not null;default:sent" json:"status"`
	ErrorMsg  string    `gorm:"type:text;default:''" json:"error_msg,omitempty"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
