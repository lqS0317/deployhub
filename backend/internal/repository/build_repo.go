package repository

import "deployhub/internal/model"

// BuildRepository 构建记录数据访问接口
type BuildRepository interface {
	Create(build *model.Build) error
	FindByID(id uint) (*model.Build, error)
	Update(build *model.Build) error
	UpdateFields(id uint, fields map[string]interface{}) error
	List(page, pageSize int, serviceID *uint) ([]model.Build, int64, error)
	UpdateStatus(id uint, status string) error
	AppendLog(id uint, logChunk string) error
	Delete(id uint) error
}
