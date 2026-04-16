package handler

import (
	"net/http"
	"strconv"
	"time"

	"deployhub/internal/middleware"
	"deployhub/internal/pkg"
	"deployhub/internal/service/audit"

	"github.com/gin-gonic/gin"
)

// AuditHandler 审计日志处理器
type AuditHandler struct {
	auditSvc *audit.AuditService
}

// NewAuditHandler 创建审计日志处理器
func NewAuditHandler(auditSvc *audit.AuditService) *AuditHandler {
	return &AuditHandler{auditSvc: auditSvc}
}

// List 查询审计日志列表，支持多条件筛选
func (h *AuditHandler) List(c *gin.Context) {
	page, pageSize := pkg.GetPagination(c)

	var userID *uint
	if uid := c.Query("user_id"); uid != "" {
		id, err := strconv.ParseUint(uid, 10, 32)
		if err != nil {
			pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的用户 ID")
			return
		}
		u := uint(id)
		userID = &u
	}

	action := c.Query("action")
	resourceType := c.Query("resource_type")

	var from, to *time.Time
	if f := c.Query("from"); f != "" {
		t, err := time.Parse(time.RFC3339, f)
		if err != nil {
			pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的起始时间，格式须为 RFC3339")
			return
		}
		from = &t
	}
	if t := c.Query("to"); t != "" {
		parsed, err := time.Parse(time.RFC3339, t)
		if err != nil {
			pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的结束时间，格式须为 RFC3339")
			return
		}
		to = &parsed
	}

	logs, total, err := h.auditSvc.List(page, pageSize, userID, action, resourceType, from, to)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询审计日志失败")
		return
	}

	pkg.Paginated(c, logs, total, page, pageSize)
}

// RegisterAuditRoutes 注册审计日志路由（仅管理员可访问）
func RegisterAuditRoutes(r *gin.RouterGroup, h *AuditHandler) {
	r.GET("/audit-logs", middleware.AdminOnly(), h.List)
}
