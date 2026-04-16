package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type buildRepository struct {
	db *gorm.DB
}

// NewBuildRepository 创建构建记录仓储实例
func NewBuildRepository(db *gorm.DB) BuildRepository {
	return &buildRepository{db: db}
}

func (r *buildRepository) Create(build *model.Build) error {
	return r.db.Create(build).Error
}

func (r *buildRepository) FindByID(id uint) (*model.Build, error) {
	var build model.Build
	err := r.db.Preload("Service").Preload("TriggerUser").Preload("BuildCluster").First(&build, id).Error
	if err != nil {
		return nil, err
	}
	return &build, nil
}

func (r *buildRepository) Update(build *model.Build) error {
	return r.db.Save(build).Error
}

func (r *buildRepository) List(page, pageSize int, serviceID *uint) ([]model.Build, int64, error) {
	var builds []model.Build
	var total int64

	query := r.db.Model(&model.Build{})
	if serviceID != nil {
		query = query.Where("service_id = ?", *serviceID)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Service").Preload("TriggerUser").
		Offset(offset).Limit(pageSize).Order("id DESC").Find(&builds).Error
	return builds, total, err
}

func (r *buildRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&model.Build{}).Where("id = ?", id).Updates(fields).Error
}

func (r *buildRepository) Delete(id uint) error {
	return r.db.Delete(&model.Build{}, id).Error
}

func (r *buildRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Build{}).Where("id = ?", id).Update("status", status).Error
}

// AppendLog 使用 SQL 拼接高效追加日志，避免先读后写
func (r *buildRepository) AppendLog(id uint, logChunk string) error {
	return r.db.Model(&model.Build{}).Where("id = ?", id).
		Update("log", gorm.Expr("COALESCE(log, '') || ?", logChunk)).Error
}
