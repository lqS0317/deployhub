package repository

import "deployhub/internal/model"

// RoutePermissionRepository 路由权限数据访问接口
type RoutePermissionRepository interface {
	List() ([]model.RoutePermission, error)
	FindByClusterAndUser(clusterID, userID uint) (*model.RoutePermission, error)
	Upsert(perm *model.RoutePermission) error
	Delete(id uint) error
}
