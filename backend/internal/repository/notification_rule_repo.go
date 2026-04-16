package repository

import "deployhub/internal/model"

// NotificationRuleRepository 全局通知规则数据访问
type NotificationRuleRepository interface {
	List() ([]model.NotificationRule, error)
	FindByEventType(eventType string) ([]model.NotificationRule, error)
	Upsert(rule *model.NotificationRule) error
	Delete(id uint) error
}

// ServiceNotificationRuleRepository 服务级通知规则数据访问（一个服务一条记录一个渠道）
type ServiceNotificationRuleRepository interface {
	List() ([]model.ServiceNotificationRule, error)
	FindByService(serviceID uint) (*model.ServiceNotificationRule, error)
	Upsert(rule *model.ServiceNotificationRule) error
	Delete(id uint) error
}

// NotificationLogRepository 通知发送记录数据访问
type NotificationLogRepository interface {
	Create(log *model.NotificationLog) error
	List(serviceID *uint, eventType, status string, page, pageSize int) ([]model.NotificationLog, int64, error)
}
