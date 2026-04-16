package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type configTemplateRepository struct {
	db *gorm.DB
}

// NewConfigTemplateRepository 创建配置模板仓储实例
func NewConfigTemplateRepository(db *gorm.DB) ConfigTemplateRepository {
	return &configTemplateRepository{db: db}
}

func (r *configTemplateRepository) Create(tpl *model.ConfigTemplate) error {
	return r.db.Create(tpl).Error
}

func (r *configTemplateRepository) FindByID(id uint) (*model.ConfigTemplate, error) {
	var tpl model.ConfigTemplate
	err := r.db.First(&tpl, id).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

func (r *configTemplateRepository) Update(tpl *model.ConfigTemplate) error {
	return r.db.Save(tpl).Error
}

func (r *configTemplateRepository) Delete(id uint) error {
	return r.db.Delete(&model.ConfigTemplate{}, id).Error
}

func (r *configTemplateRepository) ListByService(serviceID uint) ([]model.ConfigTemplate, error) {
	var list []model.ConfigTemplate
	err := r.db.Where("service_id = ?", serviceID).Order("id DESC").Find(&list).Error
	return list, err
}
