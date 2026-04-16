package audit

import (
	"encoding/json"
	"strings"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/repository"

	"gorm.io/datatypes"
)

// 需要从审计详情中过滤的敏感关键词
var sensitiveKeys = []string{"password", "token", "secret", "key", "credential"}

// AuditService 审计日志服务
type AuditService struct {
	repo repository.AuditLogRepository
}

// NewAuditService 创建审计服务实例
func NewAuditService(repo repository.AuditLogRepository) *AuditService {
	return &AuditService{repo: repo}
}

// Log 记录审计日志，自动过滤敏感字段
func (s *AuditService) Log(userID uint, action, resourceType string, resourceID uint, detail interface{}, ipAddress string) error {
	log := &model.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    ipAddress,
	}

	if detail != nil {
		raw, err := json.Marshal(detail)
		if err != nil {
			return err
		}
		sanitized := sanitizeJSON(raw)
		log.Detail = datatypes.JSON(sanitized)
	}

	return s.repo.Create(log)
}

// List 查询审计日志，支持多条件筛选
func (s *AuditService) List(page, pageSize int, userID *uint, action, resourceType string, from, to *time.Time) ([]model.AuditLog, int64, error) {
	return s.repo.List(page, pageSize, userID, action, resourceType, from, to)
}

// sanitizeJSON 移除 JSON 中包含敏感关键词的字段
func sanitizeJSON(raw []byte) []byte {
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return raw
	}

	for k := range data {
		lower := strings.ToLower(k)
		for _, s := range sensitiveKeys {
			if strings.Contains(lower, s) {
				delete(data, k)
				break
			}
		}
	}

	result, err := json.Marshal(data)
	if err != nil {
		return raw
	}
	return result
}
