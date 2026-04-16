package handler

import (
	"net/http"
	"strconv"

	"deployhub/internal/middleware"
	"deployhub/internal/model"
	"deployhub/internal/pkg"
	"deployhub/internal/repository"
	"deployhub/internal/service/notification"

	"github.com/gin-gonic/gin"
)

// NotificationChannelHandler 通知渠道处理器
type NotificationChannelHandler struct {
	channelRepo   repository.NotificationChannelRepository
	webhookSender *notification.WebhookSender
}

// NewNotificationChannelHandler 创建通知渠道处理器
func NewNotificationChannelHandler(
	channelRepo repository.NotificationChannelRepository,
	webhookSender *notification.WebhookSender,
) *NotificationChannelHandler {
	return &NotificationChannelHandler{
		channelRepo:   channelRepo,
		webhookSender: webhookSender,
	}
}

type createChannelRequest struct {
	Name       string `json:"name" binding:"required"`
	Type       string `json:"type" binding:"required"`
	WebhookURL string `json:"webhook_url" binding:"required"`
}

// Create 创建通知渠道（管理员）
func (h *NotificationChannelHandler) Create(c *gin.Context) {
	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	if !isValidChannelType(req.Type) {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "不支持的渠道类型，可选: feishu, dingtalk, slack, generic")
		return
	}

	if !isHTTPS(req.WebhookURL) {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "Webhook URL 必须使用 HTTPS 协议")
		return
	}

	ch := &model.NotificationChannel{
		Name:       req.Name,
		Type:       req.Type,
		WebhookURL: req.WebhookURL,
	}
	if err := h.channelRepo.Create(ch); err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, "创建通知渠道失败，名称可能已存在")
		return
	}

	c.JSON(http.StatusCreated, ch)
}

// List 列出通知渠道（管理员）
func (h *NotificationChannelHandler) List(c *gin.Context) {
	page, pageSize := pkg.GetPagination(c)
	items, total, err := h.channelRepo.List(page, pageSize)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询通知渠道列表失败")
		return
	}
	pkg.Paginated(c, items, total, page, pageSize)
}

// Get 获取通知渠道详情（管理员）
func (h *NotificationChannelHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的渠道 ID")
		return
	}

	ch, err := h.channelRepo.FindByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "通知渠道不存在")
		return
	}

	c.JSON(http.StatusOK, ch)
}

type updateChannelRequest struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	WebhookURL string `json:"webhook_url"`
}

// Update 更新通知渠道（管理员）
func (h *NotificationChannelHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的渠道 ID")
		return
	}

	ch, err := h.channelRepo.FindByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "通知渠道不存在")
		return
	}

	var req updateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}

	if req.Name != "" {
		ch.Name = req.Name
	}
	if req.Type != "" {
		if !isValidChannelType(req.Type) {
			pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "不支持的渠道类型")
			return
		}
		ch.Type = req.Type
	}
	if req.WebhookURL != "" {
		if !isHTTPS(req.WebhookURL) {
			pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "Webhook URL 必须使用 HTTPS 协议")
			return
		}
		ch.WebhookURL = req.WebhookURL
	}

	if err := h.channelRepo.Update(ch); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "更新通知渠道失败")
		return
	}

	c.JSON(http.StatusOK, ch)
}

// Delete 删除通知渠道（管理员）
func (h *NotificationChannelHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的渠道 ID")
		return
	}

	if err := h.channelRepo.Delete(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "删除通知渠道失败")
		return
	}

	c.Status(http.StatusNoContent)
}

// TestWebhook 测试 Webhook 连通性（管理员）
func (h *NotificationChannelHandler) TestWebhook(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的渠道 ID")
		return
	}

	ch, err := h.channelRepo.FindByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "通知渠道不存在")
		return
	}

	if err := h.webhookSender.Send(ch.WebhookURL, ch.Type, "DeployHub 测试通知", "这是一条测试消息，确认 Webhook 连通性。"); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "测试消息发送成功"})
}

// RegisterNotificationChannelRoutes 注册通知渠道相关路由（管理员专用）
func RegisterNotificationChannelRoutes(r *gin.RouterGroup, h *NotificationChannelHandler) {
	channels := r.Group("/notification-channels")
	channels.Use(middleware.AdminOnly())
	{
		channels.GET("", h.List)
		channels.POST("", h.Create)
		channels.GET("/:id", h.Get)
		channels.PUT("/:id", h.Update)
		channels.DELETE("/:id", h.Delete)
		channels.POST("/:id/test", h.TestWebhook)
	}
}

func isValidChannelType(t string) bool {
	switch t {
	case "feishu", "dingtalk", "slack", "generic":
		return true
	}
	return false
}

func isHTTPS(rawURL string) bool {
	return len(rawURL) >= 8 && rawURL[:8] == "https://"
}
