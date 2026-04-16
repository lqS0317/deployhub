package svc

import (
	"testing"

	"deployhub/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAddMember(t *testing.T) {
	memberRepo := new(mockMemberRepo)
	ms := NewMemberService(memberRepo)
	memberRepo.On("FindByServiceAndUser", uint(1), uint(2)).Return(nil, gorm.ErrRecordNotFound)
	memberRepo.On("Create", mock.AnythingOfType("*model.ServiceMember")).Return(nil)
	err := ms.AddMember(1, 2, "developer")
	require.NoError(t, err)
}

func TestAddMemberAlreadyExists(t *testing.T) {
	memberRepo := new(mockMemberRepo)
	ms := NewMemberService(memberRepo)
	existing := &model.ServiceMember{ServiceID: 1, UserID: 2, Role: "viewer"}
	memberRepo.On("FindByServiceAndUser", uint(1), uint(2)).Return(existing, nil)
	err := ms.AddMember(1, 2, "developer")
	assert.Error(t, err)
}

func TestUpdateMemberRole(t *testing.T) {
	memberRepo := new(mockMemberRepo)
	ms := NewMemberService(memberRepo)
	existing := &model.ServiceMember{ID: 1, ServiceID: 1, UserID: 2, Role: "viewer"}
	memberRepo.On("FindByServiceAndUser", uint(1), uint(2)).Return(existing, nil)
	memberRepo.On("Update", mock.AnythingOfType("*model.ServiceMember")).Return(nil)
	err := ms.UpdateRole(1, 2, "developer")
	require.NoError(t, err)
}

func TestRemoveMember(t *testing.T) {
	memberRepo := new(mockMemberRepo)
	ms := NewMemberService(memberRepo)
	existing := &model.ServiceMember{ID: 5, ServiceID: 1, UserID: 2, Role: "developer"}
	memberRepo.On("FindByServiceAndUser", uint(1), uint(2)).Return(existing, nil)
	memberRepo.On("Delete", uint(5)).Return(nil)
	err := ms.RemoveMember(1, 2)
	require.NoError(t, err)
}
