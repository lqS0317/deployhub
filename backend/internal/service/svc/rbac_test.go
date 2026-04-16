package svc

import (
	"testing"

	"deployhub/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// mockUserRepoForRBAC 用于 RBAC 测试的 User 仓储 mock
type mockUserRepoForRBAC struct{ mock.Mock }

func (m *mockUserRepoForRBAC) Create(u *model.User) error       { return m.Called(u).Error(0) }
func (m *mockUserRepoForRBAC) FindByUsername(n string) (*model.User, error) {
	args := m.Called(n)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepoForRBAC) FindByEmail(e string) (*model.User, error) {
	args := m.Called(e)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepoForRBAC) FindByOAuth(p, o string) (*model.User, error) {
	args := m.Called(p, o)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepoForRBAC) Update(u *model.User) error        { return m.Called(u).Error(0) }
func (m *mockUserRepoForRBAC) List(p, ps int) ([]model.User, int64, error) {
	args := m.Called(p, ps)
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}
func (m *mockUserRepoForRBAC) UpdateRole(id uint, role string) error   { return m.Called(id, role).Error(0) }
func (m *mockUserRepoForRBAC) UpdateStatus(id uint, status string) error { return m.Called(id, status).Error(0) }
func (m *mockUserRepoForRBAC) FindByID(id uint) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepoForRBAC) FindByRole(role string) ([]model.User, error) { return nil, nil }

// mockGroupPermRepoForRBAC 组权限仓储 mock
type mockGroupPermRepoForRBAC struct{ mock.Mock }

func (m *mockGroupPermRepoForRBAC) Create(p *model.GroupServicePermission) error { return m.Called(p).Error(0) }
func (m *mockGroupPermRepoForRBAC) FindByID(id uint) (*model.GroupServicePermission, error) { return nil, nil }
func (m *mockGroupPermRepoForRBAC) Update(id uint, role string) error { return nil }
func (m *mockGroupPermRepoForRBAC) Delete(id uint) error { return nil }
func (m *mockGroupPermRepoForRBAC) ListByGroup(gid uint) ([]model.GroupServicePermission, error) { return nil, nil }
func (m *mockGroupPermRepoForRBAC) FindByGroupAndService(gid, sid uint) (*model.GroupServicePermission, error) { return nil, gorm.ErrRecordNotFound }
func (m *mockGroupPermRepoForRBAC) DeleteByGroup(gid uint) error { return nil }
func (m *mockGroupPermRepoForRBAC) FindAllByUser(uid uint) ([]model.GroupServicePermission, error) { return nil, nil }
func (m *mockGroupPermRepoForRBAC) FindRolesByUserAndService(uid, sid uint) ([]model.GroupServicePermission, error) {
	args := m.Called(uid, sid)
	return args.Get(0).([]model.GroupServicePermission), args.Error(1)
}

// mockServiceRepoForRBAC mock
type mockServiceRepoForRBAC struct{ mock.Mock }
func (m *mockServiceRepoForRBAC) Create(s *model.Service) error { return nil }
func (m *mockServiceRepoForRBAC) FindByID(id uint) (*model.Service, error) { return nil, nil }
func (m *mockServiceRepoForRBAC) FindByName(n string) (*model.Service, error) { return nil, nil }
func (m *mockServiceRepoForRBAC) Update(s *model.Service) error { return nil }
func (m *mockServiceRepoForRBAC) Delete(id uint) error { return nil }
func (m *mockServiceRepoForRBAC) List(p, ps int) ([]model.Service, int64, error) { return nil, 0, nil }
func (m *mockServiceRepoForRBAC) BatchCreate(svcs []*model.Service) error { return nil }

func buildEffectiveRoleSvc(memberRepo *mockMemberRepo, userRepo *mockUserRepoForRBAC, groupPermRepo *mockGroupPermRepoForRBAC) *EffectiveRoleService {
	svcRepo := &mockServiceRepoForRBAC{}
	return NewEffectiveRoleService(memberRepo, groupPermRepo, userRepo, svcRepo)
}

func TestCheckPermission_OwnerCanDoAll(t *testing.T) {
	memberRepo := new(mockMemberRepo)
	userRepo := new(mockUserRepoForRBAC)
	groupPermRepo := new(mockGroupPermRepoForRBAC)

	userRepo.On("FindByID", uint(10)).Return(&model.User{ID: 10, Role: "member"}, nil)
	memberRepo.On("GetUserRole", uint(1), uint(10)).Return("owner", nil)
	groupPermRepo.On("FindRolesByUserAndService", uint(10), uint(1)).Return([]model.GroupServicePermission{}, nil)

	erSvc := buildEffectiveRoleSvc(memberRepo, userRepo, groupPermRepo)
	rbac := NewRBACService(erSvc)

	assert.True(t, rbac.CheckPermission(1, 10, "viewer"))
	assert.True(t, rbac.CheckPermission(1, 10, "developer"))
	assert.True(t, rbac.CheckPermission(1, 10, "owner"))
}

func TestCheckPermission_DeveloperCannotOwner(t *testing.T) {
	memberRepo := new(mockMemberRepo)
	userRepo := new(mockUserRepoForRBAC)
	groupPermRepo := new(mockGroupPermRepoForRBAC)

	userRepo.On("FindByID", uint(10)).Return(&model.User{ID: 10, Role: "member"}, nil)
	memberRepo.On("GetUserRole", uint(1), uint(10)).Return("developer", nil)
	groupPermRepo.On("FindRolesByUserAndService", uint(10), uint(1)).Return([]model.GroupServicePermission{}, nil)

	erSvc := buildEffectiveRoleSvc(memberRepo, userRepo, groupPermRepo)
	rbac := NewRBACService(erSvc)

	assert.True(t, rbac.CheckPermission(1, 10, "viewer"))
	assert.True(t, rbac.CheckPermission(1, 10, "developer"))
	assert.False(t, rbac.CheckPermission(1, 10, "owner"))
}

func TestCheckPermission_NotMember(t *testing.T) {
	memberRepo := new(mockMemberRepo)
	userRepo := new(mockUserRepoForRBAC)
	groupPermRepo := new(mockGroupPermRepoForRBAC)

	userRepo.On("FindByID", uint(99)).Return(&model.User{ID: 99, Role: "member"}, nil)
	memberRepo.On("GetUserRole", uint(1), uint(99)).Return("", gorm.ErrRecordNotFound)
	groupPermRepo.On("FindRolesByUserAndService", uint(99), uint(1)).Return([]model.GroupServicePermission{}, nil)

	erSvc := buildEffectiveRoleSvc(memberRepo, userRepo, groupPermRepo)
	rbac := NewRBACService(erSvc)

	assert.False(t, rbac.CheckPermission(1, 99, "viewer"))
}

func TestCheckPermission_AdminBypassAll(t *testing.T) {
	memberRepo := new(mockMemberRepo)
	userRepo := new(mockUserRepoForRBAC)
	groupPermRepo := new(mockGroupPermRepoForRBAC)

	userRepo.On("FindByID", uint(1)).Return(&model.User{ID: 1, Role: "admin"}, nil)

	erSvc := buildEffectiveRoleSvc(memberRepo, userRepo, groupPermRepo)
	rbac := NewRBACService(erSvc)

	assert.True(t, rbac.CheckPermission(999, 1, "owner"))
}

func TestCheckPermission_GroupPermissionUpgrade(t *testing.T) {
	memberRepo := new(mockMemberRepo)
	userRepo := new(mockUserRepoForRBAC)
	groupPermRepo := new(mockGroupPermRepoForRBAC)

	userRepo.On("FindByID", uint(10)).Return(&model.User{ID: 10, Role: "member"}, nil)
	memberRepo.On("GetUserRole", uint(1), uint(10)).Return("viewer", nil)
	groupPermRepo.On("FindRolesByUserAndService", uint(10), uint(1)).Return([]model.GroupServicePermission{
		{GroupID: 5, Role: "developer", Group: &model.Group{Name: "后端团队"}},
	}, nil)

	erSvc := buildEffectiveRoleSvc(memberRepo, userRepo, groupPermRepo)
	rbac := NewRBACService(erSvc)

	// 个人 viewer + 组 developer → 有效权限 developer
	assert.True(t, rbac.CheckPermission(1, 10, "developer"))
	assert.False(t, rbac.CheckPermission(1, 10, "owner"))
}
