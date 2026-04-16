package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type serviceRepository struct{ db *gorm.DB }

func NewServiceRepository(db *gorm.DB) ServiceRepository { return &serviceRepository{db: db} }

func (r *serviceRepository) Create(svc *model.Service) error { return r.db.Create(svc).Error }

func (r *serviceRepository) FindByID(id uint) (*model.Service, error) {
	var svc model.Service
	err := r.db.Preload("GitRepo").Preload("Registry").Preload("Cluster").Preload("Owner").First(&svc, id).Error
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

func (r *serviceRepository) FindByName(name string) (*model.Service, error) {
	var svc model.Service
	err := r.db.Where("name = ?", name).First(&svc).Error
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

func (r *serviceRepository) Update(svc *model.Service) error { return r.db.Save(svc).Error }
func (r *serviceRepository) Delete(id uint) error  { return r.db.Delete(&model.Service{}, id).Error }

func (r *serviceRepository) List(page, pageSize int) ([]model.Service, int64, error) {
	var services []model.Service
	var total int64
	r.db.Model(&model.Service{}).Count(&total)
	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("id DESC").
		Preload("Cluster").Preload("GitRepo").Find(&services).Error
	return services, total, err
}

func (r *serviceRepository) BatchCreate(services []*model.Service) error {
	return r.db.CreateInBatches(services, 50).Error
}
