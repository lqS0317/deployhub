package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORS 跨域资源共享中间件，支持通过环境变量 CORS_ORIGINS 配置允许的来源
func CORS() gin.HandlerFunc {
	allowOrigins := "http://localhost:3000"
	if v := os.Getenv("CORS_ORIGINS"); v != "" {
		allowOrigins = v
	}
	originSet := make(map[string]struct{})
	for _, o := range strings.Split(allowOrigins, ",") {
		originSet[strings.TrimSpace(o)] = struct{}{}
	}

	maxAge := 12 * time.Hour

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if _, ok := originSet[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length")
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", int(maxAge.Seconds())))
		}

		// 预检请求直接返回 204
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
