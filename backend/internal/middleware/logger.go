package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger 结构化请求日志中间件，记录请求方法、路径、状态码、耗时、客户端 IP
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查端点，减少日志噪声
		if c.Request.URL.Path == "/api/v1/health" {
			c.Next()
			return
		}

		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		msg := fmt.Sprintf(
			`{"level":"%s","method":"%s","path":"%s","status":%d,"latency":"%s","client_ip":"%s"}`,
			levelForStatus(status), method, path, status, latency, clientIP,
		)

		switch {
		case status >= 500:
			log.Printf("[ERROR] %s", msg)
		case status >= 400:
			log.Printf("[WARN]  %s", msg)
		default:
			log.Printf("[INFO]  %s", msg)
		}
	}
}

// levelForStatus 根据 HTTP 状态码返回日志级别
func levelForStatus(status int) string {
	switch {
	case status >= 500:
		return "ERROR"
	case status >= 400:
		return "WARN"
	default:
		return "INFO"
	}
}
