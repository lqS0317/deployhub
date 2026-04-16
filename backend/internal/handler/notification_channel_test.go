package handler

import (
	"bytes"
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

// mockChannelRepo handler 测试用 mock 仓储
type mockChannelRepo struct {
	channels map[uint]*model.NotificationChannel
	nextID   uint
}

func newMockChannelRepo() *mockChannelRepo {
	return &mockChannelRepo{channels: make(map[uint]*model.NotificationChannel), nextID: 1}
}

func (m *mockChannelRepo) Create(ch *model.NotificationChannel) error {
	for _, existing := range m.channels {
		if existing.Name == ch.Name {
			return errors.New("名称已存在")
		}
	}
	ch.ID = m.nextID
	m.nextID++
	m.channels[ch.ID] = ch
	return nil
}
func (m *mockChannelRepo) FindByID(id uint) (*model.NotificationChannel, error) {
	if ch, ok := m.channels[id]; ok {
		return ch, nil
	}
	return nil, errors.New("不存在")
}
func (m *mockChannelRepo) Update(ch *model.NotificationChannel) error {
	m.channels[ch.ID] = ch
	return nil
}
func (m *mockChannelRepo) Delete(id uint) error {
	delete(m.channels, id)
	return nil
}
func (m *mockChannelRepo) List(page, pageSize int) ([]model.NotificationChannel, int64, error) {
	var result []model.NotificationChannel
	for _, ch := range m.channels {
		result = append(result, *ch)
	}
	return result, int64(len(result)), nil
}
func (m *mockChannelRepo) FindByName(name string) (*model.NotificationChannel, error) {
	for _, ch := range m.channels {
		if ch.Name == name {
			return ch, nil
		}
	}
	return nil, errors.New("不存在")
}

var _ repository.NotificationChannelRepository = (*mockChannelRepo)(nil)

func setupChannelTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	repo := newMockChannelRepo()
	sender := notification.NewWebhookSender()
	h := NewNotificationChannelHandler(repo, sender)

	r := gin.New()
	api := r.Group("/api/v1")
	// 注入管理员角色
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("user_role", "admin")
		c.Next()
	})
	RegisterNotificationChannelRoutes(api, h)
	return r
}

func TestNotificationChannelHandler_Create(t *testing.T) {
	r := setupChannelTestRouter()

	body := map[string]string{
		"name":        "飞书告警",
		"type":        "feishu",
		"webhook_url": "https://open.feishu.cn/hook/xxx",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/notification-channels", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp model.NotificationChannel
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "飞书告警", resp.Name)
	assert.Equal(t, "feishu", resp.Type)
}

func TestNotificationChannelHandler_CreateHTTPRejected(t *testing.T) {
	r := setupChannelTestRouter()

	body := map[string]string{
		"name":        "不安全渠道",
		"type":        "generic",
		"webhook_url": "http://example.com/hook",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/notification-channels", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNotificationChannelHandler_List(t *testing.T) {
	r := setupChannelTestRouter()

	// 先创建一个渠道
	body := map[string]string{"name": "test", "type": "slack", "webhook_url": "https://hooks.slack.com/xxx"}
	jsonBody, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/notification-channels", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/notification-channels", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNotificationChannelHandler_AdminRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newMockChannelRepo()
	sender := notification.NewWebhookSender()
	h := NewNotificationChannelHandler(repo, sender)

	r := gin.New()
	api := r.Group("/api/v1")
	// 注入普通用户角色
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint(2))
		c.Set("user_role", "member")
		c.Next()
	})
	RegisterNotificationChannelRoutes(api, h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notification-channels", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
