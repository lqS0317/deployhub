package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"deployhub/internal/model"
	"deployhub/internal/repository"
	"deployhub/internal/service/notification"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNotifRepoForHandler handler 测试用 mock 仓储
type mockNotifRepoForHandler struct {
	notifications map[uint]*model.Notification
	nextID        uint
}

func newMockNotifRepoForHandler() *mockNotifRepoForHandler {
	return &mockNotifRepoForHandler{notifications: make(map[uint]*model.Notification), nextID: 1}
}

func (m *mockNotifRepoForHandler) Create(n *model.Notification) error {
	n.ID = m.nextID
	m.nextID++
	m.notifications[n.ID] = n
	return nil
}
func (m *mockNotifRepoForHandler) FindByID(id uint) (*model.Notification, error) {
	if n, ok := m.notifications[id]; ok {
		return n, nil
	}
	return nil, errors.New("不存在")
}
func (m *mockNotifRepoForHandler) Update(n *model.Notification) error {
	m.notifications[n.ID] = n
	return nil
}
func (m *mockNotifRepoForHandler) List(userID uint, page, pageSize int, isRead *bool) ([]model.Notification, int64, error) {
	var result []model.Notification
	for _, n := range m.notifications {
		if n.UserID != userID {
			continue
		}
		if isRead != nil && n.IsRead != *isRead {
			continue
		}
		result = append(result, *n)
	}
	return result, int64(len(result)), nil
}
func (m *mockNotifRepoForHandler) MarkAllRead(userID uint) (int64, error) {
	var count int64
	for _, n := range m.notifications {
		if n.UserID == userID && !n.IsRead {
			n.IsRead = true
			count++
		}
	}
	return count, nil
}
func (m *mockNotifRepoForHandler) UnreadCount(userID uint) (int64, error) {
	var count int64
	for _, n := range m.notifications {
		if n.UserID == userID && !n.IsRead {
			count++
		}
	}
	return count, nil
}

var _ repository.NotificationRepository = (*mockNotifRepoForHandler)(nil)

func setupNotificationTestRouter() (*gin.Engine, *notification.NotificationService) {
	gin.SetMode(gin.TestMode)

	repo := newMockNotifRepoForHandler()
	svc := notification.NewNotificationService(repo)
	h := NewNotificationHandler(svc)

	r := gin.New()
	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint(10))
		c.Set("user_role", "member")
		c.Next()
	})
	RegisterNotificationRoutes(api, h)
	return r, svc
}

func TestNotificationHandler_List(t *testing.T) {
	r, svc := setupNotificationTestRouter()
	_, _ = svc.Create(10, "build_complete", "构建完成", "", "", 0)
	_, _ = svc.Create(10, "deploy_result", "部署结果", "", "", 0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(2), resp["total"])
}

func TestNotificationHandler_MarkRead(t *testing.T) {
	r, svc := setupNotificationTestRouter()
	n, _ := svc.Create(10, "build_complete", "构建完成", "", "", 0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/notifications/"+itoa(n.ID)+"/read", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotificationHandler_MarkAllRead(t *testing.T) {
	r, svc := setupNotificationTestRouter()
	_, _ = svc.Create(10, "build_complete", "通知1", "", "", 0)
	_, _ = svc.Create(10, "deploy_result", "通知2", "", "", 0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/notifications/read-all", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotificationHandler_UnreadCount(t *testing.T) {
	r, svc := setupNotificationTestRouter()
	_, _ = svc.Create(10, "build_complete", "通知1", "", "", 0)
	_, _ = svc.Create(10, "deploy_result", "通知2", "", "", 0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications/unread-count", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(2), resp["count"])
}
