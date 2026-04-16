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
	"deployhub/internal/service/build"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBuildRepoForHandler handler 测试用 mock 仓储
type mockBuildRepoForHandler struct {
	builds  map[uint]*model.Build
	nextID  uint
}

func newMockBuildRepoForHandler() *mockBuildRepoForHandler {
	return &mockBuildRepoForHandler{builds: make(map[uint]*model.Build), nextID: 1}
}

func (m *mockBuildRepoForHandler) Create(b *model.Build) error {
	b.ID = m.nextID
	m.nextID++
	m.builds[b.ID] = b
	return nil
}
func (m *mockBuildRepoForHandler) FindByID(id uint) (*model.Build, error) {
	if b, ok := m.builds[id]; ok {
		return b, nil
	}
	return nil, errors.New("不存在")
}
func (m *mockBuildRepoForHandler) Update(b *model.Build) error {
	m.builds[b.ID] = b
	return nil
}
func (m *mockBuildRepoForHandler) List(page, pageSize int, serviceID *uint) ([]model.Build, int64, error) {
	var result []model.Build
	for _, b := range m.builds {
		if serviceID != nil && b.ServiceID != *serviceID {
			continue
		}
		result = append(result, *b)
	}
	return result, int64(len(result)), nil
}
func (m *mockBuildRepoForHandler) UpdateFields(id uint, fields map[string]interface{}) error {
	if b, ok := m.builds[id]; ok {
		if s, ok := fields["status"]; ok {
			b.Status = s.(string)
		}
		return nil
	}
	return errors.New("不存在")
}
func (m *mockBuildRepoForHandler) UpdateStatus(id uint, status string) error {
	if b, ok := m.builds[id]; ok {
		b.Status = status
		return nil
	}
	return errors.New("不存在")
}
func (m *mockBuildRepoForHandler) AppendLog(id uint, logChunk string) error {
	if b, ok := m.builds[id]; ok {
		b.Log += logChunk
		return nil
	}
	return errors.New("不存在")
}

func (m *mockBuildRepoForHandler) Delete(id uint) error { return nil }

type mockServiceRepoForHandler struct {
	services map[uint]*model.Service
}

func newMockServiceRepoForHandler() *mockServiceRepoForHandler {
	return &mockServiceRepoForHandler{services: make(map[uint]*model.Service)}
}

func (m *mockServiceRepoForHandler) Create(_ *model.Service) error              { return nil }
func (m *mockServiceRepoForHandler) FindByName(_ string) (*model.Service, error) { return nil, errors.New("不存在") }
func (m *mockServiceRepoForHandler) Update(_ *model.Service) error              { return nil }
func (m *mockServiceRepoForHandler) Delete(_ uint) error                        { return nil }
func (m *mockServiceRepoForHandler) BatchCreate(_ []*model.Service) error       { return nil }
func (m *mockServiceRepoForHandler) List(_, _ int) ([]model.Service, int64, error) {
	return nil, 0, nil
}
func (m *mockServiceRepoForHandler) FindByID(id uint) (*model.Service, error) {
	if s, ok := m.services[id]; ok {
		return s, nil
	}
	return nil, errors.New("服务不存在")
}

var _ repository.BuildRepository = (*mockBuildRepoForHandler)(nil)
var _ repository.ServiceRepository = (*mockServiceRepoForHandler)(nil)

func setupBuildTestRouter() (*gin.Engine, *build.BuildService) {
	gin.SetMode(gin.TestMode)

	buildRepo := newMockBuildRepoForHandler()
	serviceRepo := newMockServiceRepoForHandler()
	serviceRepo.services[1] = &model.Service{
		ID:        1,
		Name:      "test-svc",
		ImageRepo: "registry.example.com/test-svc",
	}

	buildSvc := build.NewBuildService(buildRepo, serviceRepo)
	handler := NewBuildHandler(buildSvc, nil, nil)

	r := gin.New()
	api := r.Group("/api/v1")
	// 注入模拟用户 ID
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint(10))
		c.Set("user_role", "admin")
		c.Next()
	})
	RegisterBuildRoutes(api, handler)
	return r, buildSvc
}

func TestBuildHandler_TriggerBuild(t *testing.T) {
	r, _ := setupBuildTestRouter()

	body := map[string]interface{}{
		"service_id":       1,
		"build_cluster_id": 2,
		"git_branch":       "main",
		"git_commit":       "abc123",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/builds", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp model.Build
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, model.BuildStatusPending, resp.Status)
	assert.Equal(t, "main", resp.GitBranch)
}

func TestBuildHandler_ListBuilds(t *testing.T) {
	r, buildSvc := setupBuildTestRouter()

	_, err := buildSvc.CreateBuild(1, 10, 2, "main", "abc", "", "", "", nil, "", "")
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/builds", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBuildHandler_GetBuild(t *testing.T) {
	r, buildSvc := setupBuildTestRouter()

	b, err := buildSvc.CreateBuild(1, 10, 2, "main", "abc", "", "", "", nil, "", "")
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/builds/"+itoa(b.ID), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBuildHandler_CancelBuild(t *testing.T) {
	r, buildSvc := setupBuildTestRouter()

	b, err := buildSvc.CreateBuild(1, 10, 2, "main", "abc", "", "", "", nil, "", "")
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/builds/"+itoa(b.ID)+"/cancel", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBuildHandler_GetNotFound(t *testing.T) {
	r, _ := setupBuildTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/builds/999", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func itoa(n uint) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
