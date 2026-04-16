package repository

import (
	"deployhub/internal/model"
	"time"

	"gorm.io/gorm"
)

type auditLogRepository struct{ db *gorm.DB }

// NewAuditLogRepository 创建审计日志仓储实例
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditLogRepository) List(page, pageSize int, userID *uint, action, resourceType string, from, to *time.Time) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	q := r.db.Model(&model.AuditLog{})

	if userID != nil {
		q = q.Where("user_id = ?", *userID)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if resourceType != "" {
		q = q.Where("resource_type = ?", resourceType)
	}
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at <= ?", *to)
	}

	q.Count(&total)

	offset := (page - 1) * pageSize
	err := q.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Preload("User").
		Find(&logs).Error

	return logs, total, err
}
