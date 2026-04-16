package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"deployhub/internal/service/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestJWTAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSvc := auth.NewJWTService("test-secret-key")

	t.Run("有效令牌", func(t *testing.T) {
		token, _ := jwtSvc.GenerateToken(1, "alice", "admin")

		w := httptest.NewRecorder()
		c, r := gin.CreateTestContext(w)
		r.Use(JWTAuth(jwtSvc))
		r.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"user_id": GetUserID(c)})
		})

		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, c.Request)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("缺少令牌", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.Use(JWTAuth(jwtSvc))
		r.GET("/test", func(c *gin.Context) {})

		req := httptest.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("无效令牌", func(t *testing.T) {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.Use(JWTAuth(jwtSvc))
		r.GET("/test", func(c *gin.Context) {})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
