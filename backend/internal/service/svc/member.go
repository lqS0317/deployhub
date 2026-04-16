package svc

import (
	"errors"
	"fmt"

	"deployhub/internal/model"
	"deployhub/internal/repository"

	"gorm.io/gorm"
)

// MemberService 服务成员管理
type MemberService struct {
	memberRepo repository.ServiceMemberRepository
}

func NewMemberService(memberRepo repository.ServiceMemberRepository) *MemberService {
	return &MemberService{memberRepo: memberRepo}
}

func (s *MemberService) AddMember(serviceID, userID uint, role string) error {
	if _, err := s.memberRepo.FindByServiceAndUser(serviceID, userID); err == nil {
		return fmt.Errorf("该用户已是服务成员")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询成员失败: %w", err)
	}
	member := &model.ServiceMember{ServiceID: serviceID, UserID: userID, Role: role}
	return s.memberRepo.Create(member)
}

func (s *MemberService) UpdateRole(serviceID, userID uint, newRole string) error {
	member, err := s.memberRepo.FindByServiceAndUser(serviceID, userID)
	if err != nil {
		return fmt.Errorf("成员不存在: %w", err)
	}
	member.Role = newRole
	return s.memberRepo.Update(member)
}

func (s *MemberService) RemoveMember(serviceID, userID uint) error {
	member, err := s.memberRepo.FindByServiceAndUser(serviceID, userID)
	if err != nil {
		return fmt.Errorf("成员不存在: %w", err)
	}
	return s.memberRepo.Delete(member.ID)
}

func (s *MemberService) ListMembers(serviceID uint) ([]model.ServiceMember, error) {
	return s.memberRepo.ListByService(serviceID)
}

func (s *MemberService) GetOwners(serviceID uint) ([]model.ServiceMember, error) {
	return s.memberRepo.FindOwnersByService(serviceID)
}
