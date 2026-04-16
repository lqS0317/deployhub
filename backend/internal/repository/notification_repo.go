package repository

import "deployhub/internal/model"

// NotificationRepository 站内通知数据访问接口
type NotificationRepository interface {
	Create(n *model.Notification) error
	FindByID(id uint) (*model.Notification, error)
	Update(n *model.Notification) error
	List(userID uint, page, pageSize int, isRead *bool) ([]model.Notification, int64, error)
	MarkAllRead(userID uint) (int64, error)
	UnreadCount(userID uint) (int64, error)
}
