package svc

import (
	"testing"

	"deployhub/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// mock repos
type mockSvcRepo struct{ mock.Mock }

func (m *mockSvcRepo) Create(s *model.Service) error { return m.Called(s).Error(0) }
func (m *mockSvcRepo) FindByID(id uint) (*model.Service, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Service), args.Error(1)
}
func (m *mockSvcRepo) FindByName(n string) (*model.Service, error) {
	args := m.Called(n)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Service), args.Error(1)
}
func (m *mockSvcRepo) Update(s *model.Service) error  { return m.Called(s).Error(0) }
func (m *mockSvcRepo) Delete(id uint) error           { return m.Called(id).Error(0) }
func (m *mockSvcRepo) List(p, ps int) ([]model.Service, int64, error) {
	args := m.Called(p, ps)
	return args.Get(0).([]model.Service), args.Get(1).(int64), args.Error(2)
}
func (m *mockSvcRepo) BatchCreate(svcs []*model.Service) error { return m.Called(svcs).Error(0) }

type mockMemberRepo struct{ mock.Mock }

func (m *mockMemberRepo) Create(mb *model.ServiceMember) error { return m.Called(mb).Error(0) }
func (m *mockMemberRepo) FindByServiceAndUser(sid, uid uint) (*model.ServiceMember, error) {
	args := m.Called(sid, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ServiceMember), args.Error(1)
}
func (m *mockMemberRepo) ListByService(sid uint) ([]model.ServiceMember, error) {
	args := m.Called(sid)
	return args.Get(0).([]model.ServiceMember), args.Error(1)
}
func (m *mockMemberRepo) Update(mb *model.ServiceMember) error { return m.Called(mb).Error(0) }
func (m *mockMemberRepo) Delete(id uint) error                 { return m.Called(id).Error(0) }
func (m *mockMemberRepo) FindOwnersByService(sid uint) ([]model.ServiceMember, error) {
	args := m.Called(sid)
	return args.Get(0).([]model.ServiceMember), args.Error(1)
}
func (m *mockMemberRepo) GetUserRole(sid, uid uint) (string, error) {
	args := m.Called(sid, uid)
	return args.Get(0).(string), args.Error(1)
}
func (m *mockMemberRepo) ListByUser(uid uint) ([]model.ServiceMember, error) {
	return nil, nil
}

func TestCreateService(t *testing.T) {
	svcRepo := new(mockSvcRepo)
	memberRepo := new(mockMemberRepo)
	s := NewServiceService(svcRepo, memberRepo)

	svcRepo.On("FindByName", "my-svc").Return(nil, gorm.ErrRecordNotFound)
	svcRepo.On("Create", mock.AnythingOfType("*model.Service")).Return(nil)
	memberRepo.On("Create", mock.AnythingOfType("*model.ServiceMember")).Return(nil)

	cid := uint(1)
	rid := uint(1)
	svc := &model.Service{Name: "my-svc", GitRepoID: 1, RegistryID: &rid, ClusterID: &cid, ImageRepo: "repo/img", Port: 8080, OwnerID: 1}
	result, err := s.Create(svc)
	require.NoError(t, err)
	assert.Equal(t, "my-svc", result.Name)
	memberRepo.AssertCalled(t, "Create", mock.AnythingOfType("*model.ServiceMember"))
}

func TestCreateServiceDuplicate(t *testing.T) {
	svcRepo := new(mockSvcRepo)
	memberRepo := new(mockMemberRepo)
	s := NewServiceService(svcRepo, memberRepo)

	existing := &model.Service{Name: "my-svc"}
	svcRepo.On("FindByName", "my-svc").Return(existing, nil)

	svc := &model.Service{Name: "my-svc"}
	_, err := s.Create(svc)
	assert.Error(t, err)
}
