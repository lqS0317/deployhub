package repository

import "deployhub/internal/model"

// ConfigPermissionRepository 配置权限数据访问接口
type ConfigPermissionRepository interface {
	// List 列出服务下的所有权限，预加载用户
	List(serviceID uint) ([]model.ConfigPermission, error)
	// FindByUserAndCluster 查找用户在指定集群的权限（含全局权限 cluster_id=0）
	FindByUserAndCluster(serviceID, clusterID, userID uint) (*model.ConfigPermission, error)
	// Upsert 创建或更新权限（基于唯一索引冲突更新角色）
	Upsert(perm *model.ConfigPermission) error
	Delete(id uint) error
}
