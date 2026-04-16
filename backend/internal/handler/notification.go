package handler

import (
	"net/http"
	"strconv"

	"deployhub/internal/middleware"
	"deployhub/internal/pkg"
	"deployhub/internal/service/notification"

	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知处理器
type NotificationHandler struct {
	notifSvc *notification.NotificationService
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler(notifSvc *notification.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifSvc: notifSvc}
}

// List 查询当前用户通知列表
func (h *NotificationHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, pageSize := pkg.GetPagination(c)

	var isRead *bool
	if v := c.Query("is_read"); v != "" {
		switch v {
		case "true":
			b := true
			isRead = &b
		case "false":
			b := false
			isRead = &b
		default:
			pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "is_read 参数无效，应为 true 或 false")
			return
		}
	}

	items, total, err := h.notifSvc.List(userID, page, pageSize, isRead)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询通知列表失败")
		return
	}

	pkg.Paginated(c, items, total, page, pageSize)
}

// MarkRead 标记单条通知为已读
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的通知 ID")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.notifSvc.MarkRead(uint(id), userID); err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "通知不存在或无权操作")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已标记为已读"})
}

// MarkAllRead 标记当前用户所有通知为已读
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	count, err := h.notifSvc.MarkAllRead(userID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "标记已读失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已全部标记为已读", "count": count})
}

// UnreadCount 获取当前用户未读通知数量
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	count, err := h.notifSvc.UnreadCount(userID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "获取未读数量失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

// RegisterNotificationRoutes 注册通知相关路由
func RegisterNotificationRoutes(r *gin.RouterGroup, h *NotificationHandler) {
	notifications := r.Group("/notifications")
	{
		notifications.GET("", h.List)
		notifications.PUT("/read-all", h.MarkAllRead)
		notifications.GET("/unread-count", h.UnreadCount)
		notifications.PUT("/:id/read", h.MarkRead)
	}
}
