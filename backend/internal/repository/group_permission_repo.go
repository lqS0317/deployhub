package repository

import "deployhub/internal/model"

// GroupPermissionRepository 组 Service 权限数据访问接口
type GroupPermissionRepository interface {
	Create(perm *model.GroupServicePermission) error
	FindByID(id uint) (*model.GroupServicePermission, error)
	Update(id uint, role string) error
	Delete(id uint) error
	ListByGroup(groupID uint) ([]model.GroupServicePermission, error)
	FindByGroupAndService(groupID, serviceID uint) (*model.GroupServicePermission, error)
	// FindRolesByUserAndService 通过 JOIN group_members 查询用户通过所有组获得的角色
	FindRolesByUserAndService(userID, serviceID uint) ([]model.GroupServicePermission, error)
	DeleteByGroup(groupID uint) error
	// FindAllByUser 获取用户通过组获得的所有权限（权限全景用）
	FindAllByUser(userID uint) ([]model.GroupServicePermission, error)
}
