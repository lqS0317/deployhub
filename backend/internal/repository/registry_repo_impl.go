package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type registryRepository struct {
	db *gorm.DB
}

// NewRegistryRepository 创建镜像仓库数据访问实例
func NewRegistryRepository(db *gorm.DB) RegistryRepository {
	return &registryRepository{db: db}
}

func (r *registryRepository) Create(reg *model.Registry) error {
	return r.db.Create(reg).Error
}

func (r *registryRepository) FindByID(id uint) (*model.Registry, error) {
	var reg model.Registry
	err := r.db.First(&reg, id).Error
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

func (r *registryRepository) FindByName(name string) (*model.Registry, error) {
	var reg model.Registry
	err := r.db.Where("name = ?", name).First(&reg).Error
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

func (r *registryRepository) Update(reg *model.Registry) error {
	return r.db.Save(reg).Error
}

func (r *registryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Registry{}, id).Error
}

func (r *registryRepository) List(page, pageSize int) ([]model.Registry, int64, error) {
	var regs []model.Registry
	var total int64
	r.db.Model(&model.Registry{}).Count(&total)
	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&regs).Error
	return regs, total, err
}
