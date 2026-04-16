package handler

import (
	"net/http"
	"strconv"

	"deployhub/internal/middleware"
	"deployhub/internal/pkg"
	"deployhub/internal/service/svc"

	"github.com/gin-gonic/gin"
)

// PermissionHandler 权限总览处理器
type PermissionHandler struct {
	effectiveRoleSvc *svc.EffectiveRoleService
}

func NewPermissionHandler(effectiveRoleSvc *svc.EffectiveRoleService) *PermissionHandler {
	return &PermissionHandler{effectiveRoleSvc: effectiveRoleSvc}
}

// GetUserPermissions 查看指定用户的有效权限（Admin only）
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	perms, err := h.effectiveRoleSvc.GetAllEffectivePermissions(uint(userID))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询权限失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": perms})
}

// GetMyPermissions 查看当前用户自己的有效权限
func (h *PermissionHandler) GetMyPermissions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	perms, err := h.effectiveRoleSvc.GetAllEffectivePermissions(userID)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询权限失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": perms})
}

// RegisterPermissionRoutes 注册权限总览路由
func RegisterPermissionRoutes(r *gin.RouterGroup, h *PermissionHandler) {
	r.GET("/users/:id/permissions", middleware.AdminOnly(), h.GetUserPermissions)
}
