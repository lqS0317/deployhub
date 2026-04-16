package approval

import (
	"deployhub/internal/model"
	"deployhub/internal/repository"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ repository.UserRepository = (*mockUserRepo)(nil)

// --- Mock 仓储 ---

type mockApprovalRepo struct {
	approvals  map[uint]*model.Approval
	nextID     uint
	createErr  error
	findByIDFn func(id uint) (*model.Approval, error)
}

func newMockApprovalRepo() *mockApprovalRepo {
	return &mockApprovalRepo{approvals: make(map[uint]*model.Approval), nextID: 1}
}

func (m *mockApprovalRepo) Create(a *model.Approval) error {
	if m.createErr != nil {
		return m.createErr
	}
	a.ID = m.nextID
	m.nextID++
	m.approvals[a.ID] = a
	return nil
}

func (m *mockApprovalRepo) FindByID(id uint) (*model.Approval, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(id)
	}
	a, ok := m.approvals[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return a, nil
}

func (m *mockApprovalRepo) Update(a *model.Approval) error {
	m.approvals[a.ID] = a
	return nil
}

func (m *mockApprovalRepo) List(page, pageSize int, status string, approverID *uint) ([]model.Approval, int64, error) {
	var result []model.Approval
	for _, a := range m.approvals {
		if status != "" && status != "all" && a.Status != status {
			continue
		}
		if approverID != nil && a.ApproverID != *approverID {
			continue
		}
		result = append(result, *a)
	}
	return result, int64(len(result)), nil
}

func (m *mockApprovalRepo) FindByDeployment(deploymentID uint) ([]model.Approval, error) {
	var result []model.Approval
	for _, a := range m.approvals {
		if a.DeploymentID == deploymentID {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *mockApprovalRepo) FindPendingByDeployment(deploymentID uint) (*model.Approval, error) {
	for _, a := range m.approvals {
		if a.DeploymentID == deploymentID && a.Status == model.ApprovalStatusPending {
			return a, nil
		}
	}
	return nil, errors.New("not found")
}

type mockDeployRepo struct {
	deployments map[uint]*model.Deployment
	statusLog   map[uint]string
}

func newMockDeployRepo() *mockDeployRepo {
	return &mockDeployRepo{
		deployments: make(map[uint]*model.Deployment),
		statusLog:   make(map[uint]string),
	}
}

func (m *mockDeployRepo) Create(d *model.Deployment) error { return nil }
func (m *mockDeployRepo) Update(d *model.Deployment) error { return nil }
func (m *mockDeployRepo) List(page, pageSize int, serviceID *uint) ([]model.Deployment, int64, error) {
	return nil, 0, nil
}
func (m *mockDeployRepo) FindActiveByService(serviceID uint) (*model.Deployment, error) {
	return nil, nil
}

func (m *mockDeployRepo) FindLastSuccessful(serviceID uint) (*model.Deployment, error) {
	return nil, nil
}

func (m *mockDeployRepo) FindByID(id uint) (*model.Deployment, error) {
	d, ok := m.deployments[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return d, nil
}

func (m *mockDeployRepo) UpdateStatus(id uint, status string) error {
	d, ok := m.deployments[id]
	if !ok {
		return errors.New("not found")
	}
	d.Status = status
	m.statusLog[id] = status
	return nil
}

func (m *mockDeployRepo) UpdateStatusWithReason(id uint, status, reason string) error {
	d, ok := m.deployments[id]
	if !ok {
		return errors.New("not found")
	}
	d.Status = status
	d.FailReason = reason
	m.statusLog[id] = status
	return nil
}

func (m *mockDeployRepo) UpdatePodStatus(id uint, status, podStatus, podMessage string) error {
	d, ok := m.deployments[id]
	if !ok {
		return errors.New("not found")
	}
	d.Status = status
	d.PodStatus = podStatus
	d.PodMessage = podMessage
	return nil
}

func (m *mockDeployRepo) UpdateField(id uint, field string, value interface{}) error {
	return nil
}

func (m *mockDeployRepo) Delete(id uint) error {
	delete(m.deployments, id)
	delete(m.statusLog, id)
	return nil
}

type mockUserRepo struct {
	byID         map[uint]*model.User
	byRole       map[string][]model.User
	findByIDFn   func(id uint) (*model.User, error)
	findByRoleFn func(role string) ([]model.User, error)
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		byID:   make(map[uint]*model.User),
		byRole: make(map[string][]model.User),
	}
}

func (m *mockUserRepo) Create(_ *model.User) error { return nil }

func (m *mockUserRepo) FindByID(id uint) (*model.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(id)
	}
	u, ok := m.byID[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (m *mockUserRepo) FindByUsername(_ string) (*model.User, error) {
	return nil, errors.New("not found")
}
func (m *mockUserRepo) FindByEmail(_ string) (*model.User, error) {
	return nil, errors.New("not found")
}
func (m *mockUserRepo) FindByOAuth(_, _ string) (*model.User, error) {
	return nil, errors.New("not found")
}
func (m *mockUserRepo) Update(_ *model.User) error                 { return nil }
func (m *mockUserRepo) List(_, _ int) ([]model.User, int64, error) { return nil, 0, nil }
func (m *mockUserRepo) UpdateRole(_ uint, _ string) error          { return nil }
func (m *mockUserRepo) UpdateStatus(_ uint, _ string) error        { return nil }

func (m *mockUserRepo) FindByRole(role string) ([]model.User, error) {
	if m.findByRoleFn != nil {
		return m.findByRoleFn(role)
	}
	users, ok := m.byRole[role]
	if !ok {
		return nil, nil
	}
	return users, nil
}

// --- 审批规则单元测试 ---

func TestApprovalRule_NeedsApproval(t *testing.T) {
	rule := &ApprovalRule{}

	tests := []struct {
		name       string
		globalRole string
		want       bool
	}{
		{"admin 无需审批（由上层跳过 CreateApprovalForAdmins）", "admin", false},
		{"member 需要审批", "member", true},
		{"developer 需要审批", "developer", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.NeedsApproval(tt.globalRole)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestApprovalRule_CanApprove(t *testing.T) {
	rule := &ApprovalRule{}

	assert.True(t, rule.CanApprove("admin", 10, 30), "admin 可审批他人发起的部署")
	assert.False(t, rule.CanApprove("admin", 10, 10), "不能审批自己发起的部署")
	assert.False(t, rule.CanApprove("member", 99, 30), "非 admin 不可审批")
	assert.False(t, rule.CanApprove("developer", 99, 30), "非 admin 不可审批")
}

// --- 审批服务测试 ---

func setupTestService() (*ApprovalService, *mockApprovalRepo, *mockDeployRepo, *mockUserRepo) {
	approvalRepo := newMockApprovalRepo()
	deployRepo := newMockDeployRepo()
	userRepo := newMockUserRepo()
	svc := NewApprovalService(approvalRepo, deployRepo, userRepo, nil)
	return svc, approvalRepo, deployRepo, userRepo
}

func TestCreateApprovalForAdmins_NonAdminCreates_OneRecordPerOtherAdmin(t *testing.T) {
	svc, approvalRepo, deployRepo, userRepo := setupTestService()

	deployment := &model.Deployment{ID: 3, ServiceID: 100, Status: model.DeployStatusPreviewing}
	deployRepo.deployments[3] = deployment

	userRepo.byRole["admin"] = []model.User{
		{ID: 50, Role: "admin"},
		{ID: 51, Role: "admin"},
	}

	err := svc.CreateApprovalForAdmins(3, 30)
	require.NoError(t, err)

	assert.Len(t, approvalRepo.approvals, 2)
	for _, a := range approvalRepo.approvals {
		assert.Equal(t, model.ApprovalStatusPending, a.Status)
		assert.Equal(t, uint(3), a.DeploymentID)
		assert.Equal(t, uint(30), a.RequesterID)
		assert.Contains(t, []uint{50, 51}, a.ApproverID)
	}
	assert.Equal(t, model.DeployStatusPendingApproval, deployment.Status)
}

func TestCreateApprovalForAdmins_ExcludesRequesterWhenRequesterIsAdmin(t *testing.T) {
	svc, approvalRepo, deployRepo, userRepo := setupTestService()

	deployment := &model.Deployment{ID: 4, ServiceID: 100, Status: model.DeployStatusPreviewing}
	deployRepo.deployments[4] = deployment

	userRepo.byRole["admin"] = []model.User{
		{ID: 30, Role: "admin"},
		{ID: 50, Role: "admin"},
		{ID: 51, Role: "admin"},
	}

	err := svc.CreateApprovalForAdmins(4, 30)
	require.NoError(t, err)

	assert.Len(t, approvalRepo.approvals, 2)
	for _, a := range approvalRepo.approvals {
		assert.NotEqual(t, uint(30), a.ApproverID, "不应给发起人本人创建待办")
		assert.Contains(t, []uint{50, 51}, a.ApproverID)
	}
}

func TestCreateApprovalForAdmins_ErrNoAdmins(t *testing.T) {
	svc, _, deployRepo, userRepo := setupTestService()

	deployment := &model.Deployment{ID: 5, ServiceID: 100, Status: model.DeployStatusPreviewing}
	deployRepo.deployments[5] = deployment

	// 仅一名 admin 且即为发起人 → 无可审批人
	userRepo.byRole["admin"] = []model.User{{ID: 30, Role: "admin"}}

	err := svc.CreateApprovalForAdmins(5, 30)
	require.ErrorIs(t, err, ErrNoAdmins)
}

func TestApprove_Success(t *testing.T) {
	svc, approvalRepo, deployRepo, userRepo := setupTestService()

	deployment := &model.Deployment{ID: 1, ServiceID: 100, Status: model.DeployStatusPendingApproval}
	deployRepo.deployments[1] = deployment

	approvalRepo.approvals[1] = &model.Approval{
		ID:           1,
		DeploymentID: 1,
		RequesterID:  30,
		ApproverID:   50,
		Status:       model.ApprovalStatusPending,
	}

	userRepo.byID[50] = &model.User{ID: 50, Role: "admin"}

	err := svc.Approve(1, 50, "同意发布")
	require.NoError(t, err)

	a := approvalRepo.approvals[1]
	assert.Equal(t, model.ApprovalStatusApproved, a.Status)
	assert.Equal(t, "同意发布", a.Comment)
	assert.NotNil(t, a.DecidedAt)
	assert.Equal(t, model.DeployStatusApproved, deployment.Status)
}

func TestApprove_NotAdmin(t *testing.T) {
	svc, approvalRepo, deployRepo, userRepo := setupTestService()

	deployment := &model.Deployment{ID: 1, ServiceID: 100, Status: model.DeployStatusPendingApproval}
	deployRepo.deployments[1] = deployment

	approvalRepo.approvals[1] = &model.Approval{
		ID:           1,
		DeploymentID: 1,
		RequesterID:  30,
		ApproverID:   50,
		Status:       model.ApprovalStatusPending,
	}

	userRepo.byID[99] = &model.User{ID: 99, Role: "member"}

	err := svc.Approve(1, 99, "")
	assert.ErrorIs(t, err, ErrNotAuthorized)
}

func TestApprove_RequesterCannotApproveOwn(t *testing.T) {
	svc, approvalRepo, deployRepo, userRepo := setupTestService()

	deployRepo.deployments[1] = &model.Deployment{ID: 1, ServiceID: 100, Status: model.DeployStatusPendingApproval}

	approvalRepo.approvals[1] = &model.Approval{
		ID:           1,
		DeploymentID: 1,
		RequesterID:  30,
		ApproverID:   50,
		Status:       model.ApprovalStatusPending,
	}

	userRepo.byID[30] = &model.User{ID: 30, Role: "admin"}

	err := svc.Approve(1, 30, "")
	assert.ErrorIs(t, err, ErrNotAuthorized)
}

func TestApprove_AlreadyDecided(t *testing.T) {
	svc, approvalRepo, _, userRepo := setupTestService()

	now := time.Now()
	approvalRepo.approvals[1] = &model.Approval{
		ID:           1,
		DeploymentID: 1,
		ApproverID:   50,
		Status:       model.ApprovalStatusApproved,
		DecidedAt:    &now,
	}

	userRepo.byID[50] = &model.User{ID: 50, Role: "admin"}

	err := svc.Approve(1, 50, "")
	assert.ErrorIs(t, err, ErrApprovalDecided)
}

func TestReject_Success(t *testing.T) {
	svc, approvalRepo, deployRepo, userRepo := setupTestService()

	deployment := &model.Deployment{ID: 1, ServiceID: 100, Status: model.DeployStatusPendingApproval}
	deployRepo.deployments[1] = deployment

	approvalRepo.approvals[1] = &model.Approval{
		ID:           1,
		DeploymentID: 1,
		RequesterID:  30,
		ApproverID:   50,
		Status:       model.ApprovalStatusPending,
	}

	userRepo.byID[50] = &model.User{ID: 50, Role: "admin"}

	err := svc.Reject(1, 50, "镜像未经安全扫描")
	require.NoError(t, err)

	a := approvalRepo.approvals[1]
	assert.Equal(t, model.ApprovalStatusRejected, a.Status)
	assert.Equal(t, "镜像未经安全扫描", a.Comment)
	assert.NotNil(t, a.DecidedAt)
	assert.Equal(t, model.DeployStatusRejected, deployment.Status)
}

func TestReject_NotAdmin(t *testing.T) {
	svc, approvalRepo, deployRepo, userRepo := setupTestService()

	deployRepo.deployments[1] = &model.Deployment{ID: 1, ServiceID: 100, Status: model.DeployStatusPendingApproval}

	approvalRepo.approvals[1] = &model.Approval{
		ID:           1,
		DeploymentID: 1,
		RequesterID:  30,
		ApproverID:   50,
		Status:       model.ApprovalStatusPending,
	}

	userRepo.byID[99] = &model.User{ID: 99, Role: "member"}

	err := svc.Reject(1, 99, "拒绝")
	assert.ErrorIs(t, err, ErrNotAuthorized)
}

func TestReject_AlreadyDecided(t *testing.T) {
	svc, approvalRepo, _, userRepo := setupTestService()

	now := time.Now()
	approvalRepo.approvals[1] = &model.Approval{
		ID:           1,
		DeploymentID: 1,
		ApproverID:   50,
		Status:       model.ApprovalStatusRejected,
		DecidedAt:    &now,
	}

	userRepo.byID[50] = &model.User{ID: 50, Role: "admin"}

	err := svc.Reject(1, 50, "再次拒绝")
	assert.ErrorIs(t, err, ErrApprovalDecided)
}

func TestGetByID(t *testing.T) {
	svc, approvalRepo, _, _ := setupTestService()

	approvalRepo.approvals[1] = &model.Approval{
		ID:           1,
		DeploymentID: 1,
		ApproverID:   50,
		Status:       model.ApprovalStatusPending,
	}

	a, err := svc.GetByID(1)
	require.NoError(t, err)
	assert.Equal(t, uint(1), a.ID)

	_, err = svc.GetByID(999)
	assert.Error(t, err)
}

func TestList(t *testing.T) {
	svc, approvalRepo, _, _ := setupTestService()

	approvalRepo.approvals[1] = &model.Approval{ID: 1, ApproverID: 50, Status: model.ApprovalStatusPending}
	approvalRepo.approvals[2] = &model.Approval{ID: 2, ApproverID: 50, Status: model.ApprovalStatusApproved}
	approvalRepo.approvals[3] = &model.Approval{ID: 3, ApproverID: 60, Status: model.ApprovalStatusPending}

	// 不过滤
	items, total, err := svc.List(1, 20, "all", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, items, 3)

	// 按状态过滤
	items, total, err = svc.List(1, 20, model.ApprovalStatusPending, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)

	// 按审批人过滤
	approverID := uint(50)
	items, total, err = svc.List(1, 20, "", &approverID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
}
