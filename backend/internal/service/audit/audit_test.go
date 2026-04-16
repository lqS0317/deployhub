package audit

import (
	"encoding/json"
	"testing"
	"time"

	"deployhub/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestLogCreatesAuditRecord(t *testing.T) {
	repo := new(mockAuditLogRepo)
	svc := NewAuditService(repo)

	detail := map[string]interface{}{"service_name": "my-svc"}
	repo.On("Create", mock.AnythingOfType("*model.AuditLog")).Return(nil)

	err := svc.Log(1, "create_service", "service", 10, detail, "192.168.1.1")
	require.NoError(t, err)

	// 验证写入的记录
	call := repo.Calls[0]
	saved := call.Arguments.Get(0).(*model.AuditLog)
	assert.Equal(t, uint(1), saved.UserID)
	assert.Equal(t, "create_service", saved.Action)
	assert.Equal(t, "service", saved.ResourceType)
	assert.Equal(t, uint(10), saved.ResourceID)
	assert.Equal(t, "192.168.1.1", saved.IPAddress)

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(saved.Detail, &parsed))
	assert.Equal(t, "my-svc", parsed["service_name"])
	repo.AssertExpectations(t)
}

func TestLogSanitizesDetail(t *testing.T) {
	repo := new(mockAuditLogRepo)
	svc := NewAuditService(repo)

	detail := map[string]interface{}{
		"username":    "alice",
		"password":    "secret123",
		"token":       "jwt-xxx",
		"api_key":     "key-123",
		"secret_data": "classified",
		"credential":  "cred-abc",
	}
	repo.On("Create", mock.AnythingOfType("*model.AuditLog")).Return(nil)

	err := svc.Log(1, "update_user", "user", 1, detail, "10.0.0.1")
	require.NoError(t, err)

	saved := repo.Calls[0].Arguments.Get(0).(*model.AuditLog)
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(saved.Detail, &parsed))

	// 安全字段应被移除
	assert.NotContains(t, parsed, "password")
	assert.NotContains(t, parsed, "token")
	assert.NotContains(t, parsed, "api_key")
	assert.NotContains(t, parsed, "secret_data")
	assert.NotContains(t, parsed, "credential")
	// 正常字段保留
	assert.Equal(t, "alice", parsed["username"])
}

func TestLogNilDetail(t *testing.T) {
	repo := new(mockAuditLogRepo)
	svc := NewAuditService(repo)
	repo.On("Create", mock.AnythingOfType("*model.AuditLog")).Return(nil)

	err := svc.Log(1, "login", "", 0, nil, "10.0.0.1")
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestListWithFilters(t *testing.T) {
	repo := new(mockAuditLogRepo)
	svc := NewAuditService(repo)

	userID := uint(1)
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)

	expected := []model.AuditLog{
		{ID: 1, UserID: 1, Action: "create_service"},
		{ID: 2, UserID: 1, Action: "delete_service"},
	}
	repo.On("List", 1, 20, &userID, "create_service", "service", &from, &to).
		Return(expected, int64(2), nil)

	logs, total, err := svc.List(1, 20, &userID, "create_service", "service", &from, &to)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, logs, 2)
	repo.AssertExpectations(t)
}

func TestListNoFilters(t *testing.T) {
	repo := new(mockAuditLogRepo)
	svc := NewAuditService(repo)

	expected := []model.AuditLog{{ID: 1}}
	repo.On("List", 1, 20, (*uint)(nil), "", "", (*time.Time)(nil), (*time.Time)(nil)).
		Return(expected, int64(1), nil)

	logs, total, err := svc.List(1, 20, nil, "", "", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, logs, 1)
}
