package handler

import (
	"net/http"
	"strconv"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/pkg"
	"deployhub/internal/repository"

	"github.com/gin-gonic/gin"
)

// NotificationRuleHandler 通知规则处理器
type NotificationRuleHandler struct {
	ruleRepo    repository.NotificationRuleRepository
	svcRuleRepo repository.ServiceNotificationRuleRepository
	logRepo     repository.NotificationLogRepository
}

func NewNotificationRuleHandler(
	ruleRepo repository.NotificationRuleRepository,
	svcRuleRepo repository.ServiceNotificationRuleRepository,
	logRepo repository.NotificationLogRepository,
) *NotificationRuleHandler {
	return &NotificationRuleHandler{ruleRepo: ruleRepo, svcRuleRepo: svcRuleRepo, logRepo: logRepo}
}

// --- 全局规则 ---

func (h *NotificationRuleHandler) ListGlobalRules(c *gin.Context) {
	rules, err := h.ruleRepo.List()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询全局规则失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rules, "event_types": model.AllEventTypes})
}

type upsertRuleRequest struct {
	ChannelID uint   `json:"channel_id" binding:"required"`
	EventType string `json:"event_type" binding:"required"`
	Enabled   *bool  `json:"enabled"`
}

func (h *NotificationRuleHandler) UpsertGlobalRule(c *gin.Context) {
	var req upsertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	rule := &model.NotificationRule{
		ChannelID: req.ChannelID,
		EventType: req.EventType,
		Enabled:   enabled,
		UpdatedAt: time.Now(),
	}
	if err := h.ruleRepo.Upsert(rule); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "保存规则失败")
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *NotificationRuleHandler) DeleteGlobalRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.ruleRepo.Delete(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "删除失败")
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 服务级规则（一个服务一个渠道） ---

// ListAllServiceRules 列出所有服务级规则
func (h *NotificationRuleHandler) ListAllServiceRules(c *gin.Context) {
	rules, err := h.svcRuleRepo.List()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询服务规则失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rules})
}

type upsertServiceRuleRequest struct {
	ServiceID uint  `json:"service_id" binding:"required"`
	ChannelID uint  `json:"channel_id" binding:"required"`
	Enabled   *bool `json:"enabled"`
}

func (h *NotificationRuleHandler) UpsertServiceRule(c *gin.Context) {
	var req upsertServiceRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	rule := &model.ServiceNotificationRule{
		ServiceID: req.ServiceID,
		ChannelID: req.ChannelID,
		Enabled:   enabled,
		UpdatedAt: time.Now(),
	}
	if err := h.svcRuleRepo.Upsert(rule); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "保存服务规则失败")
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *NotificationRuleHandler) DeleteServiceRule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.svcRuleRepo.Delete(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "删除失败")
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 发送记录 ---

func (h *NotificationRuleHandler) ListLogs(c *gin.Context) {
	page, pageSize := pkg.GetPagination(c)
	var serviceID *uint
	if sid := c.Query("service_id"); sid != "" {
		v, _ := strconv.ParseUint(sid, 10, 32)
		u := uint(v)
		serviceID = &u
	}
	eventType := c.Query("event_type")
	status := c.Query("status")

	logs, total, err := h.logRepo.List(serviceID, eventType, status, page, pageSize)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询发送记录失败")
		return
	}
	pkg.Paginated(c, logs, total, page, pageSize)
}

// RegisterNotificationRuleRoutes 注册通知规则路由
func RegisterNotificationRuleRoutes(r *gin.RouterGroup, h *NotificationRuleHandler) {
	// 全局规则
	rules := r.Group("/notification-rules")
	{
		rules.GET("", h.ListGlobalRules)
		rules.POST("", h.UpsertGlobalRule)
		rules.DELETE("/:id", h.DeleteGlobalRule)
	}
	// 服务级规则
	svcRules := r.Group("/service-notification-rules")
	{
		svcRules.GET("", h.ListAllServiceRules)
		svcRules.POST("", h.UpsertServiceRule)
		svcRules.DELETE("/:id", h.DeleteServiceRule)
	}
	// 发送记录
	r.GET("/notification-logs", h.ListLogs)
}
