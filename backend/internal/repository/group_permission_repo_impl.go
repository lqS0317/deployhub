package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type groupPermissionRepository struct {
	db *gorm.DB
}

func NewGroupPermissionRepository(db *gorm.DB) GroupPermissionRepository {
	return &groupPermissionRepository{db: db}
}

func (r *groupPermissionRepository) Create(perm *model.GroupServicePermission) error {
	return r.db.Create(perm).Error
}

func (r *groupPermissionRepository) FindByID(id uint) (*model.GroupServicePermission, error) {
	var p model.GroupServicePermission
	err := r.db.Preload("Group").Preload("Service").First(&p, id).Error
	return &p, err
}

func (r *groupPermissionRepository) Update(id uint, role string) error {
	return r.db.Model(&model.GroupServicePermission{}).Where("id = ?", id).Update("role", role).Error
}

func (r *groupPermissionRepository) Delete(id uint) error {
	return r.db.Delete(&model.GroupServicePermission{}, id).Error
}

func (r *groupPermissionRepository) ListByGroup(groupID uint) ([]model.GroupServicePermission, error) {
	var perms []model.GroupServicePermission
	err := r.db.Preload("Service").Where("group_id = ?", groupID).Order("id ASC").Find(&perms).Error
	return perms, err
}

func (r *groupPermissionRepository) FindByGroupAndService(groupID, serviceID uint) (*model.GroupServicePermission, error) {
	var p model.GroupServicePermission
	err := r.db.Where("group_id = ? AND service_id = ?", groupID, serviceID).First(&p).Error
	return &p, err
}

// FindRolesByUserAndService 查询用户通过所有所属组在指定 Service 上获得的权限
func (r *groupPermissionRepository) FindRolesByUserAndService(userID, serviceID uint) ([]model.GroupServicePermission, error) {
	var perms []model.GroupServicePermission
	err := r.db.Preload("Group").
		Joins("JOIN group_members ON group_members.group_id = group_service_permissions.group_id").
		Where("group_members.user_id = ? AND group_service_permissions.service_id = ?", userID, serviceID).
		Find(&perms).Error
	return perms, err
}

func (r *groupPermissionRepository) DeleteByGroup(groupID uint) error {
	return r.db.Where("group_id = ?", groupID).Delete(&model.GroupServicePermission{}).Error
}

// FindAllByUser 获取用户通过所有组获得的所有 Service 权限
func (r *groupPermissionRepository) FindAllByUser(userID uint) ([]model.GroupServicePermission, error) {
	var perms []model.GroupServicePermission
	err := r.db.Preload("Group").Preload("Service").
		Joins("JOIN group_members ON group_members.group_id = group_service_permissions.group_id").
		Where("group_members.user_id = ?", userID).
		Find(&perms).Error
	return perms, err
}
