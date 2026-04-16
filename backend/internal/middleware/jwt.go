package middleware

import (
	"net/http"
	"strings"

	"deployhub/internal/pkg"
	"deployhub/internal/service/auth"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT 认证中间件
func JWTAuth(jwtSvc *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			pkg.Error(c, http.StatusUnauthorized, pkg.CodeUnauthorized, "缺少 Authorization 头")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			pkg.Error(c, http.StatusUnauthorized, pkg.CodeUnauthorized, "Authorization 格式错误")
			c.Abort()
			return
		}

		claims, err := jwtSvc.ValidateToken(parts[1])
		if err != nil {
			pkg.Error(c, http.StatusUnauthorized, pkg.CodeUnauthorized, "令牌无效或已过期")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// GetUserID 从上下文获取当前用户 ID
func GetUserID(c *gin.Context) uint {
	id, _ := c.Get("user_id")
	if uid, ok := id.(uint); ok {
		return uid
	}
	return 0
}

// GetUserRole 从上下文获取当前用户角色
func GetUserRole(c *gin.Context) string {
	role, _ := c.Get("user_role")
	if r, ok := role.(string); ok {
		return r
	}
	return ""
}
