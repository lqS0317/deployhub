package handler

import (
	"net/http"
	"strconv"

	"deployhub/internal/model"
	"deployhub/internal/pkg"
	"deployhub/internal/service/routing"

	"github.com/gin-gonic/gin"
)

// RoutePermissionHandler 路由权限处理器
type RoutePermissionHandler struct {
	svc *routing.PermissionService
}

// NewRoutePermissionHandler 创建路由权限处理器
func NewRoutePermissionHandler(svc *routing.PermissionService) *RoutePermissionHandler {
	return &RoutePermissionHandler{svc: svc}
}

// List 列出所有路由权限
func (h *RoutePermissionHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询权限失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

type grantPermissionRequest struct {
	ClusterID uint   `json:"cluster_id"`
	UserID    uint   `json:"user_id" binding:"required"`
	Role      string `json:"role" binding:"required"`
}

// Grant 授权
func (h *RoutePermissionHandler) Grant(c *gin.Context) {
	var req grantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	perm := &model.RoutePermission{
		ClusterID: req.ClusterID,
		UserID:    req.UserID,
		Role:      req.Role,
	}
	if err := h.svc.Grant(perm); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "授权成功"})
}

// Revoke 撤销权限
func (h *RoutePermissionHandler) Revoke(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的 ID")
		return
	}
	if err := h.svc.Revoke(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "权限已撤销"})
}
