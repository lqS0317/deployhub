package repository

import "deployhub/internal/model"

// SystemSettingRepository 系统配置数据访问接口
type SystemSettingRepository interface {
	Get(key string) (*model.SystemSetting, error)
	GetAll() ([]model.SystemSetting, error)
	Upsert(setting *model.SystemSetting) error
}
