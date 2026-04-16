package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建通知仓储实例
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(n *model.Notification) error {
	return r.db.Create(n).Error
}

func (r *notificationRepository) FindByID(id uint) (*model.Notification, error) {
	var n model.Notification
	err := r.db.First(&n, id).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *notificationRepository) Update(n *model.Notification) error {
	return r.db.Save(n).Error
}

func (r *notificationRepository) List(userID uint, page, pageSize int, isRead *bool) ([]model.Notification, int64, error) {
	var notifications []model.Notification
	var total int64

	query := r.db.Model(&model.Notification{}).Where("user_id = ?", userID)

	// 按已读状态过滤
	if isRead != nil {
		query = query.Where("is_read = ?", *isRead)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&notifications).Error
	return notifications, total, err
}

func (r *notificationRepository) MarkAllRead(userID uint) (int64, error) {
	result := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true)
	return result.RowsAffected, result.Error
}

func (r *notificationRepository) UnreadCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}
