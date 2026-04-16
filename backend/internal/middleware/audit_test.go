package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/service/audit"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockAuditLogRepo 模拟审计日志仓储
type mockAuditLogRepo struct {
	mock.Mock
}

func (m *mockAuditLogRepo) Create(log *model.AuditLog) error {
	return m.Called(log).Error(0)
}

func (m *mockAuditLogRepo) List(page, pageSize int, userID *uint, action, resourceType string, from, to *time.Time) ([]model.AuditLog, int64, error) {
	args := m.Called(page, pageSize, userID, action, resourceType, from, to)
	return args.Get(0).([]model.AuditLog), args.Get(1).(int64), args.Error(2)
}

func setupAuditRouter(repo *mockAuditLogRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	svc := audit.NewAuditService(repo)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("username", "alice")
		c.Set("user_role", "admin")
		c.Next()
	})
	r.Use(AuditLog(svc))

	r.POST("/api/v1/services", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": 1})
	})
	r.GET("/api/v1/services", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"items": []string{}})
	})
	r.DELETE("/api/v1/services/1", func(c *gin.Context) {
		c.JSON(http.StatusNoContent, nil)
	})
	r.PUT("/api/v1/services/1", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad"})
	})
	return r
}

func TestAuditMiddlewarePOSTGeneratesLog(t *testing.T) {
	repo := new(mockAuditLogRepo)
	repo.On("Create", mock.AnythingOfType("*model.AuditLog")).Return(nil)
	r := setupAuditRouter(repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/services", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	repo.AssertCalled(t, "Create", mock.AnythingOfType("*model.AuditLog"))

	saved := repo.Calls[0].Arguments.Get(0).(*model.AuditLog)
	assert.Equal(t, uint(1), saved.UserID)
	assert.Equal(t, "create_service", saved.Action)
}

func TestAuditMiddlewareGETDoesNotLog(t *testing.T) {
	repo := new(mockAuditLogRepo)
	r := setupAuditRouter(repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/services", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestAuditMiddlewareFailedRequestDoesNotLog(t *testing.T) {
	repo := new(mockAuditLogRepo)
	r := setupAuditRouter(repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/services/1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestAuditMiddlewareDELETEGeneratesLog(t *testing.T) {
	repo := new(mockAuditLogRepo)
	repo.On("Create", mock.AnythingOfType("*model.AuditLog")).Return(nil)
	r := setupAuditRouter(repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/v1/services/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	repo.AssertCalled(t, "Create", mock.AnythingOfType("*model.AuditLog"))

	saved := repo.Calls[0].Arguments.Get(0).(*model.AuditLog)
	assert.Equal(t, "delete_service", saved.Action)
}
