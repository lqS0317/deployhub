package routing

import (
	"fmt"

	"deployhub/internal/model"
	"deployhub/internal/repository"
)

// PermissionService 路由权限服务
type PermissionService struct {
	permRepo repository.RoutePermissionRepository
}

// NewPermissionService 创建权限服务
func NewPermissionService(permRepo repository.RoutePermissionRepository) *PermissionService {
	return &PermissionService{permRepo: permRepo}
}

// CheckPermission 检查用户在指定集群上是否拥有所需角色
func (s *PermissionService) CheckPermission(clusterID, userID uint, requiredRole string) error {
	perm, err := s.permRepo.FindByClusterAndUser(clusterID, userID)
	if err != nil {
		return fmt.Errorf("无操作权限")
	}

	roleLevel := map[string]int{"viewer": 1, "editor": 2, "admin": 3}
	required := roleLevel[requiredRole]
	actual := roleLevel[perm.Role]
	if actual < required {
		return fmt.Errorf("权限不足，需要 %s 角色", requiredRole)
	}
	return nil
}

// List 列出所有权限
func (s *PermissionService) List() ([]model.RoutePermission, error) {
	return s.permRepo.List()
}

// Grant 授权（upsert）
func (s *PermissionService) Grant(perm *model.RoutePermission) error {
	return s.permRepo.Upsert(perm)
}

// Revoke 撤销权限
func (s *PermissionService) Revoke(id uint) error {
	return s.permRepo.Delete(id)
}
