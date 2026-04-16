package repository

import "deployhub/internal/model"

// GroupMemberRepository 组成员数据访问接口
type GroupMemberRepository interface {
	Create(member *model.GroupMember) error
	Delete(groupID, userID uint) error
	ListByGroup(groupID uint) ([]model.GroupMember, error)
	ListByUser(userID uint) ([]model.GroupMember, error)
	Exists(groupID, userID uint) (bool, error)
	DeleteByGroup(groupID uint) error
}
