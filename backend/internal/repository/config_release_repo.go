package repository

import "deployhub/internal/model"

// ConfigReleaseRepository 配置发布记录数据访问接口
type ConfigReleaseRepository interface {
	// List 按版本倒序列出发布记录，预加载创建者
	List(entryID uint) ([]model.ConfigRelease, error)
	FindByID(id uint) (*model.ConfigRelease, error)
	// FindLatestPublished 查找最新的已发布记录
	FindLatestPublished(entryID uint) (*model.ConfigRelease, error)
	Create(release *model.ConfigRelease) error
	// GetNextVersion 获取下一个版本号
	GetNextVersion(entryID uint) (int, error)
	UpdateStatus(id uint, status string) error
}
