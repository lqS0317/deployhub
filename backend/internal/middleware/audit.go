package middleware

import (
	"strings"

	"deployhub/internal/service/audit"

	"github.com/gin-gonic/gin"
)

// 写操作 HTTP 方法
var writeMethods = map[string]bool{
	"POST":   true,
	"PUT":    true,
	"DELETE": true,
}

// HTTP 方法到动作前缀的映射
var methodAction = map[string]string{
	"POST":   "create",
	"PUT":    "update",
	"DELETE": "delete",
}

// AuditLog 审计日志中间件，仅记录成功的写操作
func AuditLog(auditSvc *audit.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if !writeMethods[c.Request.Method] {
			return
		}

		// 仅记录成功的请求（状态码 < 400）
		if c.Writer.Status() >= 400 {
			return
		}

		userID := GetUserID(c)
		if userID == 0 {
			return
		}

		action := deriveAction(c.Request.Method, c.Request.URL.Path)
		detail := map[string]string{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		}

		_ = auditSvc.Log(userID, action, "", 0, detail, c.ClientIP())
	}
}

// deriveAction 根据 HTTP 方法和路径推导操作名称
// 例如 "POST /api/v1/services" -> "create_service"
func deriveAction(method, path string) string {
	prefix := methodAction[method]
	if prefix == "" {
		prefix = strings.ToLower(method)
	}

	segments := strings.Split(strings.Trim(path, "/"), "/")
	resource := ""
	for i := len(segments) - 1; i >= 0; i-- {
		seg := segments[i]
		if seg == "" {
			continue
		}
		// 跳过纯数字段（资源 ID）和版本号段
		if isNumeric(seg) || strings.HasPrefix(seg, "v") && len(seg) <= 3 {
			continue
		}
		resource = seg
		break
	}

	if resource == "" {
		return prefix
	}

	// 去掉复数尾部 s
	resource = strings.TrimSuffix(resource, "s")
	return prefix + "_" + resource
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
