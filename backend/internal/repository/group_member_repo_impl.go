package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type groupMemberRepository struct {
	db *gorm.DB
}

func NewGroupMemberRepository(db *gorm.DB) GroupMemberRepository {
	return &groupMemberRepository{db: db}
}

func (r *groupMemberRepository) Create(member *model.GroupMember) error {
	return r.db.Create(member).Error
}

func (r *groupMemberRepository) Delete(groupID, userID uint) error {
	result := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&model.GroupMember{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *groupMemberRepository) ListByGroup(groupID uint) ([]model.GroupMember, error) {
	var members []model.GroupMember
	err := r.db.Preload("User").Where("group_id = ?", groupID).Order("id ASC").Find(&members).Error
	return members, err
}

func (r *groupMemberRepository) ListByUser(userID uint) ([]model.GroupMember, error) {
	var members []model.GroupMember
	err := r.db.Preload("Group").Where("user_id = ?", userID).Find(&members).Error
	return members, err
}

func (r *groupMemberRepository) Exists(groupID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count).Error
	return count > 0, err
}

func (r *groupMemberRepository) DeleteByGroup(groupID uint) error {
	return r.db.Where("group_id = ?", groupID).Delete(&model.GroupMember{}).Error
}
