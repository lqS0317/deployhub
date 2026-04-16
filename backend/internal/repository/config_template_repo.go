package repository

import "deployhub/internal/model"

// ConfigTemplateRepository 配置模板数据访问接口
type ConfigTemplateRepository interface {
	Create(tpl *model.ConfigTemplate) error
	FindByID(id uint) (*model.ConfigTemplate, error)
	Update(tpl *model.ConfigTemplate) error
	Delete(id uint) error
	ListByService(serviceID uint) ([]model.ConfigTemplate, error)
}
