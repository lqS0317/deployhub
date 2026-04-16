package group

import (
	"errors"
	"fmt"

	"deployhub/internal/model"
	"deployhub/internal/repository"

	"gorm.io/gorm"
)

// GroupService 组管理服务
type GroupService struct {
	groupRepo  repository.GroupRepository
	memberRepo repository.GroupMemberRepository
	permRepo   repository.GroupPermissionRepository
}

func NewGroupService(
	groupRepo repository.GroupRepository,
	memberRepo repository.GroupMemberRepository,
	permRepo repository.GroupPermissionRepository,
) *GroupService {
	return &GroupService{
		groupRepo:  groupRepo,
		memberRepo: memberRepo,
		permRepo:   permRepo,
	}
}

// Create 创建组
func (s *GroupService) Create(name, description string, createdBy uint) (*model.Group, error) {
	if _, err := s.groupRepo.FindByName(name); err == nil {
		return nil, fmt.Errorf("组名 %s 已存在", name)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询组失败: %w", err)
	}

	g := &model.Group{
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
	}
	if err := s.groupRepo.Create(g); err != nil {
		return nil, fmt.Errorf("创建组失败: %w", err)
	}
	return g, nil
}

// GetByID 获取组详情
func (s *GroupService) GetByID(id uint) (*model.Group, error) {
	return s.groupRepo.FindByID(id)
}

// Update 更新组
func (s *GroupService) Update(id uint, name, description string) (*model.Group, error) {
	g, err := s.groupRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("组不存在: %w", err)
	}
	if name != "" {
		g.Name = name
	}
	g.Description = description
	if err := s.groupRepo.Update(g); err != nil {
		return nil, fmt.Errorf("更新组失败: %w", err)
	}
	return g, nil
}

// Delete 删除组（级联删除成员和权限）
func (s *GroupService) Delete(id uint) error {
	if _, err := s.groupRepo.FindByID(id); err != nil {
		return fmt.Errorf("组不存在: %w", err)
	}
	_ = s.permRepo.DeleteByGroup(id)
	_ = s.memberRepo.DeleteByGroup(id)
	return s.groupRepo.Delete(id)
}

// List 组列表
func (s *GroupService) List(page, pageSize int) ([]model.Group, int64, error) {
	return s.groupRepo.List(page, pageSize)
}

// GetMemberCount 获取组成员数
func (s *GroupService) GetMemberCount(groupID uint) int {
	members, _ := s.memberRepo.ListByGroup(groupID)
	return len(members)
}

// GetPermissionCount 获取组权限数
func (s *GroupService) GetPermissionCount(groupID uint) int {
	perms, _ := s.permRepo.ListByGroup(groupID)
	return len(perms)
}

// --- 成员管理 ---

// AddMembers 批量添加成员
func (s *GroupService) AddMembers(groupID uint, userIDs []uint) ([]model.GroupMember, error) {
	if _, err := s.groupRepo.FindByID(groupID); err != nil {
		return nil, fmt.Errorf("组不存在: %w", err)
	}

	var added []model.GroupMember
	for _, uid := range userIDs {
		exists, _ := s.memberRepo.Exists(groupID, uid)
		if exists {
			continue
		}
		m := &model.GroupMember{GroupID: groupID, UserID: uid}
		if err := s.memberRepo.Create(m); err != nil {
			return nil, fmt.Errorf("添加成员失败: %w", err)
		}
		added = append(added, *m)
	}
	return added, nil
}

// RemoveMember 移除成员
func (s *GroupService) RemoveMember(groupID, userID uint) error {
	return s.memberRepo.Delete(groupID, userID)
}

// ListMembers 列出组成员
func (s *GroupService) ListMembers(groupID uint) ([]model.GroupMember, error) {
	return s.memberRepo.ListByGroup(groupID)
}

// --- 权限管理 ---

// AddPermission 添加 Service 权限
func (s *GroupService) AddPermission(groupID, serviceID uint, role string) (*model.GroupServicePermission, error) {
	if _, err := s.permRepo.FindByGroupAndService(groupID, serviceID); err == nil {
		return nil, fmt.Errorf("该组已有此 Service 的权限配置")
	}

	p := &model.GroupServicePermission{
		GroupID:   groupID,
		ServiceID: serviceID,
		Role:      role,
	}
	if err := s.permRepo.Create(p); err != nil {
		return nil, fmt.Errorf("添加权限失败: %w", err)
	}
	return p, nil
}

// UpdatePermission 修改权限角色
func (s *GroupService) UpdatePermission(id uint, role string) error {
	return s.permRepo.Update(id, role)
}

// RemovePermission 移除权限
func (s *GroupService) RemovePermission(id uint) error {
	return s.permRepo.Delete(id)
}

// ListPermissions 列出组权限
func (s *GroupService) ListPermissions(groupID uint) ([]model.GroupServicePermission, error) {
	return s.permRepo.ListByGroup(groupID)
}
