package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type routeEntryRepository struct {
	db *gorm.DB
}

// NewRouteEntryRepository 创建路由条目仓储实例
func NewRouteEntryRepository(db *gorm.DB) RouteEntryRepository {
	return &routeEntryRepository{db: db}
}

func (r *routeEntryRepository) List(resourceType string) ([]model.RouteEntry, error) {
	var entries []model.RouteEntry
	q := r.db.Order("id DESC")
	if resourceType != "" {
		q = q.Where("resource_type = ?", resourceType)
	}
	err := q.Find(&entries).Error
	return entries, err
}

func (r *routeEntryRepository) FindByID(id uint) (*model.RouteEntry, error) {
	var entry model.RouteEntry
	err := r.db.First(&entry, id).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *routeEntryRepository) FindByNameAndType(name, resourceType string) (*model.RouteEntry, error) {
	var entry model.RouteEntry
	err := r.db.Where("name = ? AND resource_type = ?", name, resourceType).First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *routeEntryRepository) Create(entry *model.RouteEntry) error {
	return r.db.Create(entry).Error
}

func (r *routeEntryRepository) Update(entry *model.RouteEntry) error {
	return r.db.Save(entry).Error
}

func (r *routeEntryRepository) Delete(id uint) error {
	return r.db.Delete(&model.RouteEntry{}, id).Error
}
