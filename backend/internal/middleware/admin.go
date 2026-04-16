package middleware

import (
	"net/http"

	"deployhub/internal/pkg"

	"github.com/gin-gonic/gin"
)

// AdminOnly 仅管理员可访问
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetUserRole(c)
		if role != "admin" {
			pkg.Error(c, http.StatusForbidden, pkg.CodeForbidden, "需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
