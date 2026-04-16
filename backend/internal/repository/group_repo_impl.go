package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) Create(group *model.Group) error {
	return r.db.Create(group).Error
}

func (r *groupRepository) FindByID(id uint) (*model.Group, error) {
	var g model.Group
	err := r.db.Preload("Creator").First(&g, id).Error
	return &g, err
}

func (r *groupRepository) FindByName(name string) (*model.Group, error) {
	var g model.Group
	err := r.db.Where("name = ?", name).First(&g).Error
	return &g, err
}

func (r *groupRepository) Update(group *model.Group) error {
	return r.db.Model(group).Updates(map[string]interface{}{
		"name":        group.Name,
		"description": group.Description,
	}).Error
}

func (r *groupRepository) Delete(id uint) error {
	return r.db.Delete(&model.Group{}, id).Error
}

func (r *groupRepository) List(page, pageSize int) ([]model.Group, int64, error) {
	var groups []model.Group
	var total int64

	r.db.Model(&model.Group{}).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.Preload("Creator").Offset(offset).Limit(pageSize).Order("id DESC").Find(&groups).Error
	return groups, total, err
}
