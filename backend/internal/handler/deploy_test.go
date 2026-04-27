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
	"deployhub/internal/service/deploy"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ==================== Mock 仓储 ====================

type mockDeployRepoForHandler struct {
	deployments map[uint]*model.Deployment
	nextID      uint
}

func newMockDeployRepoForHandler() *mockDeployRepoForHandler {
	return &mockDeployRepoForHandler{deployments: make(map[uint]*model.Deployment), nextID: 1}
}

func (m *mockDeployRepoForHandler) Create(d *model.Deployment) error {
	d.ID = m.nextID
	m.nextID++
	m.deployments[d.ID] = d
	return nil
}
func (m *mockDeployRepoForHandler) FindByID(id uint) (*model.Deployment, error) {
	if d, ok := m.deployments[id]; ok {
		return d, nil
	}
	return nil, errors.New("部署不存在")
}
func (m *mockDeployRepoForHandler) Update(d *model.Deployment) error {
	m.deployments[d.ID] = d
	return nil
}
func (m *mockDeployRepoForHandler) List(page, pageSize int, serviceID *uint) ([]model.Deployment, int64, error) {
	var result []model.Deployment
	for _, d := range m.deployments {
		if serviceID != nil && d.ServiceID != *serviceID {
			continue
		}
		result = append(result, *d)
	}
	return result, int64(len(result)), nil
}
func (m *mockDeployRepoForHandler) FindActiveByService(serviceID uint) (*model.Deployment, error) {
	for _, d := range m.deployments {
		if d.ServiceID == serviceID &&
			(d.Status == model.DeployStatusPendingApproval ||
				d.Status == model.DeployStatusApproved ||
				d.Status == model.DeployStatusDeploying) {
			return d, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockDeployRepoForHandler) FindLastSuccessful(serviceID uint) (*model.Deployment, error) {
	for _, d := range m.deployments {
		if d.ServiceID == serviceID && d.Status == model.DeployStatusSuccess {
			return d, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockDeployRepoForHandler) UpdateStatus(id uint, status string) error {
	if d, ok := m.deployments[id]; ok {
		d.Status = status
		return nil
	}
	return errors.New("部署不存在")
}

func (m *mockDeployRepoForHandler) UpdateStatusWithReason(id uint, status, reason string) error {
	if d, ok := m.deployments[id]; ok {
		d.Status = status
		d.FailReason = reason
		return nil
	}
	return errors.New("部署不存在")
}

func (m *mockDeployRepoForHandler) UpdatePodStatus(id uint, status, podStatus, podMessage string) error {
	if d, ok := m.deployments[id]; ok {
		d.Status = status
		d.PodStatus = podStatus
		d.PodMessage = podMessage
	}
	return nil
}

func (m *mockDeployRepoForHandler) UpdateField(id uint, field string, value interface{}) error {
	return nil
}

func (m *mockDeployRepoForHandler) Delete(id uint) error {
	delete(m.deployments, id)
	return nil
}

func (m *mockDeployRepoForHandler) FindByStatuses(statuses []string) ([]model.Deployment, error) {
	allow := make(map[string]bool, len(statuses))
	for _, s := range statuses {
		allow[s] = true
	}
	result := make([]model.Deployment, 0)
	for _, d := range m.deployments {
		if allow[d.Status] {
			result = append(result, *d)
		}
	}
	return result, nil
}

type mockSvcRepoForDeployHandler struct {
	services map[uint]*model.Service
}

func newMockSvcRepoForDeployHandler() *mockSvcRepoForDeployHandler {
	return &mockSvcRepoForDeployHandler{services: make(map[uint]*model.Service)}
}

func (m *mockSvcRepoForDeployHandler) Create(_ *model.Service) error { return nil }
func (m *mockSvcRepoForDeployHandler) FindByName(_ string) (*model.Service, error) {
	return nil, errors.New("不存在")
}
func (m *mockSvcRepoForDeployHandler) Update(_ *model.Service) error        { return nil }
func (m *mockSvcRepoForDeployHandler) Delete(_ uint) error                  { return nil }
func (m *mockSvcRepoForDeployHandler) BatchCreate(_ []*model.Service) error { return nil }
func (m *mockSvcRepoForDeployHandler) List(_, _ int) ([]model.Service, int64, error) {
	return nil, 0, nil
}
func (m *mockSvcRepoForDeployHandler) FindByID(id uint) (*model.Service, error) {
	if s, ok := m.services[id]; ok {
		return s, nil
	}
	return nil, errors.New("服务不存在")
}

type mockBuildRepoForDeployHandler struct{}

func (m *mockBuildRepoForDeployHandler) Create(_ *model.Build) error { return nil }
func (m *mockBuildRepoForDeployHandler) FindByID(_ uint) (*model.Build, error) {
	return nil, errors.New("不存在")
}
func (m *mockBuildRepoForDeployHandler) Update(_ *model.Build) error { return nil }
func (m *mockBuildRepoForDeployHandler) List(_, _ int, _ *uint) ([]model.Build, int64, error) {
	return nil, 0, nil
}
func (m *mockBuildRepoForDeployHandler) UpdateFields(_ uint, _ map[string]interface{}) error {
	return nil
}
func (m *mockBuildRepoForDeployHandler) UpdateStatus(_ uint, _ string) error { return nil }
func (m *mockBuildRepoForDeployHandler) AppendLog(_ uint, _ string) error    { return nil }
func (m *mockBuildRepoForDeployHandler) Delete(id uint) error                { return nil }

type mockClusterNsRepoForDeployHandler struct {
	items map[uint]map[string]*model.ClusterNamespace
}

func newMockClusterNsRepoForDeployHandler() *mockClusterNsRepoForDeployHandler {
	return &mockClusterNsRepoForDeployHandler{
		items: make(map[uint]map[string]*model.ClusterNamespace),
	}
}

func (m *mockClusterNsRepoForDeployHandler) Create(ns *model.ClusterNamespace) error {
	if m.items[ns.ClusterID] == nil {
		m.items[ns.ClusterID] = make(map[string]*model.ClusterNamespace)
	}
	m.items[ns.ClusterID][ns.Namespace] = ns
	return nil
}

func (m *mockClusterNsRepoForDeployHandler) Delete(_ uint) error { return nil }

func (m *mockClusterNsRepoForDeployHandler) ListByCluster(clusterID uint) ([]model.ClusterNamespace, error) {
	clusterItems := m.items[clusterID]
	out := make([]model.ClusterNamespace, 0, len(clusterItems))
	for _, item := range clusterItems {
		out = append(out, *item)
	}
	return out, nil
}

func (m *mockClusterNsRepoForDeployHandler) FindByClusterAndNamespace(clusterID uint, namespace string) (*model.ClusterNamespace, error) {
	clusterItems := m.items[clusterID]
	if clusterItems == nil {
		return nil, gorm.ErrRecordNotFound
	}
	item, ok := clusterItems[namespace]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return item, nil
}

var _ repository.DeploymentRepository = (*mockDeployRepoForHandler)(nil)
var _ repository.ServiceRepository = (*mockSvcRepoForDeployHandler)(nil)
var _ repository.BuildRepository = (*mockBuildRepoForDeployHandler)(nil)
var _ repository.ClusterNamespaceRepository = (*mockClusterNsRepoForDeployHandler)(nil)

func setupDeployTestRouter() (*gin.Engine, *deploy.DeployService) {
	gin.SetMode(gin.TestMode)

	deployRepo := newMockDeployRepoForHandler()
	serviceRepo := newMockSvcRepoForDeployHandler()
	clusterID := uint(10)
	serviceRepo.services[1] = &model.Service{
		ID:        1,
		Name:      "test-svc",
		ClusterID: &clusterID,
		Namespace: "default",
		Replicas:  3,
		ImageRepo: "registry.example.com/test-svc",
		Port:      8080,
	}
	buildRepo := &mockBuildRepoForDeployHandler{}
	clusterNsRepo := newMockClusterNsRepoForDeployHandler()
	_ = clusterNsRepo.Create(&model.ClusterNamespace{
		ClusterID: clusterID,
		Namespace: "default",
		IsDefault: true,
	})

	deploySvc := deploy.NewDeployService(deployRepo, serviceRepo, buildRepo, clusterNsRepo)
	handler := NewDeployHandler(deploySvc, nil, nil, nil, nil, nil, nil, nil)

	r := gin.New()
	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set("user_id", uint(10))
		c.Set("user_role", "admin")
		c.Next()
	})
	RegisterDeployRoutes(api, handler)
	return r, deploySvc
}

func TestDeployHandler_CreateDeployment(t *testing.T) {
	r, _ := setupDeployTestRouter()

	body := map[string]interface{}{
		"service_id": 1,
		"image_tag":  "v1.0.0",
		"replicas":   3,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/deployments", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp model.Deployment
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", resp.ImageTag)
	assert.Equal(t, uint(10), resp.ClusterID)
}

func TestDeployHandler_ListDeployments(t *testing.T) {
	r, deploySvc := setupDeployTestRouter()

	_, err := deploySvc.CreateDeployment(1, 10, 10, "default", nil, "v1.0.0", "build", "", deploy.DeployConfig{})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/deployments", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeployHandler_GetDeployment(t *testing.T) {
	r, deploySvc := setupDeployTestRouter()

	dep, err := deploySvc.CreateDeployment(1, 10, 10, "default", nil, "v1.0.0", "build", "", deploy.DeployConfig{})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/deployments/"+itoa(dep.ID), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeployHandler_GetDeploymentNotFound(t *testing.T) {
	r, _ := setupDeployTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/deployments/999", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeployHandler_CreateDeployment_BadRequest(t *testing.T) {
	r, _ := setupDeployTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/deployments", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
