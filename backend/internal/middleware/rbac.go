package middleware

import (
	"net/http"
	"strconv"

	"deployhub/internal/pkg"
	"deployhub/internal/service/svc"

	"github.com/gin-gonic/gin"
)

// ServiceRBAC 服务级 RBAC 中间件
func ServiceRBAC(rbacSvc *svc.RBACService, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		userRole := GetUserRole(c)

		// Admin 跳过服务级权限检查
		if userRole == "admin" {
			c.Next()
			return
		}

		serviceIDStr := c.Param("id")
		serviceID, err := strconv.ParseUint(serviceIDStr, 10, 32)
		if err != nil {
			pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "无效的服务 ID")
			c.Abort()
			return
		}

		if !rbacSvc.CheckPermission(uint(serviceID), userID, requiredRole) {
			pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, "无权限操作此服务")
			c.Abort()
			return
		}

		c.Next()
	}
}
