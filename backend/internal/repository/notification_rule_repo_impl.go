package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// --- 全局通知规则 ---

type notificationRuleRepository struct{ db *gorm.DB }

func NewNotificationRuleRepository(db *gorm.DB) NotificationRuleRepository {
	return &notificationRuleRepository{db: db}
}

func (r *notificationRuleRepository) List() ([]model.NotificationRule, error) {
	var rules []model.NotificationRule
	err := r.db.Preload("Channel").Order("event_type, id").Find(&rules).Error
	return rules, err
}

func (r *notificationRuleRepository) FindByEventType(eventType string) ([]model.NotificationRule, error) {
	var rules []model.NotificationRule
	err := r.db.Preload("Channel").Where("event_type = ? AND enabled = true", eventType).Find(&rules).Error
	return rules, err
}

func (r *notificationRuleRepository) Upsert(rule *model.NotificationRule) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "channel_id"}, {Name: "event_type"}},
		DoUpdates: clause.AssignmentColumns([]string{"enabled", "updated_at"}),
	}).Create(rule).Error
}

func (r *notificationRuleRepository) Delete(id uint) error {
	return r.db.Delete(&model.NotificationRule{}, id).Error
}

// --- 服务级通知规则 ---

type serviceNotificationRuleRepository struct{ db *gorm.DB }

func NewServiceNotificationRuleRepository(db *gorm.DB) ServiceNotificationRuleRepository {
	return &serviceNotificationRuleRepository{db: db}
}

func (r *serviceNotificationRuleRepository) List() ([]model.ServiceNotificationRule, error) {
	var rules []model.ServiceNotificationRule
	err := r.db.Preload("Channel").Preload("Service").Order("id").Find(&rules).Error
	return rules, err
}

func (r *serviceNotificationRuleRepository) FindByService(serviceID uint) (*model.ServiceNotificationRule, error) {
	var rule model.ServiceNotificationRule
	err := r.db.Preload("Channel").Where("service_id = ? AND enabled = true", serviceID).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *serviceNotificationRuleRepository) Upsert(rule *model.ServiceNotificationRule) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "service_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"channel_id", "enabled", "updated_at"}),
	}).Create(rule).Error
}

func (r *serviceNotificationRuleRepository) Delete(id uint) error {
	return r.db.Delete(&model.ServiceNotificationRule{}, id).Error
}

// --- 通知发送记录 ---

type notificationLogRepository struct{ db *gorm.DB }

func NewNotificationLogRepository(db *gorm.DB) NotificationLogRepository {
	return &notificationLogRepository{db: db}
}

func (r *notificationLogRepository) Create(log *model.NotificationLog) error {
	return r.db.Create(log).Error
}

func (r *notificationLogRepository) List(serviceID *uint, eventType, status string, page, pageSize int) ([]model.NotificationLog, int64, error) {
	var logs []model.NotificationLog
	var total int64

	q := r.db.Model(&model.NotificationLog{})
	if serviceID != nil {
		q = q.Where("service_id = ?", *serviceID)
	}
	if eventType != "" {
		q = q.Where("event_type = ?", eventType)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)

	offset := (page - 1) * pageSize
	err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&logs).Error
	return logs, total, err
}
