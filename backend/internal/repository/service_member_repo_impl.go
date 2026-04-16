package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type serviceMemberRepository struct{ db *gorm.DB }

func NewServiceMemberRepository(db *gorm.DB) ServiceMemberRepository {
	return &serviceMemberRepository{db: db}
}

func (r *serviceMemberRepository) Create(member *model.ServiceMember) error {
	return r.db.Create(member).Error
}

func (r *serviceMemberRepository) FindByServiceAndUser(serviceID, userID uint) (*model.ServiceMember, error) {
	var m model.ServiceMember
	err := r.db.Where("service_id = ? AND user_id = ?", serviceID, userID).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *serviceMemberRepository) ListByService(serviceID uint) ([]model.ServiceMember, error) {
	var members []model.ServiceMember
	err := r.db.Where("service_id = ?", serviceID).Preload("User").Find(&members).Error
	return members, err
}

func (r *serviceMemberRepository) Update(member *model.ServiceMember) error {
	return r.db.Save(member).Error
}

func (r *serviceMemberRepository) Delete(id uint) error {
	return r.db.Delete(&model.ServiceMember{}, id).Error
}

func (r *serviceMemberRepository) FindOwnersByService(serviceID uint) ([]model.ServiceMember, error) {
	var members []model.ServiceMember
	err := r.db.Where("service_id = ? AND role = ?", serviceID, "owner").Preload("User").Find(&members).Error
	return members, err
}

func (r *serviceMemberRepository) ListByUser(userID uint) ([]model.ServiceMember, error) {
	var members []model.ServiceMember
	err := r.db.Where("user_id = ?", userID).Preload("Service").Find(&members).Error
	return members, err
}

func (r *serviceMemberRepository) GetUserRole(serviceID, userID uint) (string, error) {
	m, err := r.FindByServiceAndUser(serviceID, userID)
	if err != nil {
		return "", err
	}
	return m.Role, nil
}
