package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type notificationChannelRepository struct {
	db *gorm.DB
}

// NewNotificationChannelRepository 创建通知渠道仓储实例
func NewNotificationChannelRepository(db *gorm.DB) NotificationChannelRepository {
	return &notificationChannelRepository{db: db}
}

func (r *notificationChannelRepository) Create(ch *model.NotificationChannel) error {
	return r.db.Create(ch).Error
}

func (r *notificationChannelRepository) FindByID(id uint) (*model.NotificationChannel, error) {
	var ch model.NotificationChannel
	err := r.db.First(&ch, id).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *notificationChannelRepository) Update(ch *model.NotificationChannel) error {
	return r.db.Save(ch).Error
}

func (r *notificationChannelRepository) Delete(id uint) error {
	return r.db.Delete(&model.NotificationChannel{}, id).Error
}

func (r *notificationChannelRepository) List(page, pageSize int) ([]model.NotificationChannel, int64, error) {
	var channels []model.NotificationChannel
	var total int64

	r.db.Model(&model.NotificationChannel{}).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&channels).Error
	return channels, total, err
}

func (r *notificationChannelRepository) FindByName(name string) (*model.NotificationChannel, error) {
	var ch model.NotificationChannel
	err := r.db.Where("name = ?", name).First(&ch).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}
