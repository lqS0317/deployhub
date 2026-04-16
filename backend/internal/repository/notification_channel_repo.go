package repository

import "deployhub/internal/model"

// NotificationChannelRepository 通知渠道数据访问接口
type NotificationChannelRepository interface {
	Create(ch *model.NotificationChannel) error
	FindByID(id uint) (*model.NotificationChannel, error)
	Update(ch *model.NotificationChannel) error
	Delete(id uint) error
	List(page, pageSize int) ([]model.NotificationChannel, int64, error)
	FindByName(name string) (*model.NotificationChannel, error)
}
