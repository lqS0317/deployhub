package repository

import (
	"deployhub/internal/model"
	"time"
)

// AuditLogRepository 审计日志数据访问接口
type AuditLogRepository interface {
	Create(log *model.AuditLog) error
	List(page, pageSize int, userID *uint, action, resourceType string, from, to *time.Time) ([]model.AuditLog, int64, error)
}
