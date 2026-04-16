package configcenter

import (
	"errors"
	"fmt"

	"deployhub/internal/model"
	"deployhub/internal/repository"

	"gorm.io/gorm"
)

// ConfigPermissionService 配置权限服务
type ConfigPermissionService struct {
	permRepo   repository.ConfigPermissionRepository
	userRepo   repository.UserRepository
	serviceRepo repository.ServiceRepository
}

// NewConfigPermissionService 创建配置权限服务
func NewConfigPermissionService(
	permRepo repository.ConfigPermissionRepository,
	userRepo repository.UserRepository,
	serviceRepo repository.ServiceRepository,
) *ConfigPermissionService {
	return &ConfigPermissionService{
		permRepo:   permRepo,
		userRepo:   userRepo,
		serviceRepo: serviceRepo,
	}
}

// CheckPermission 校验用户是否具有指定角色权限
func (s *ConfigPermissionService) CheckPermission(serviceID, clusterID, userID uint, requiredRole string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 系统管理员直接放行
	if user.Role == "admin" {
		return nil
	}

	// 服务 owner 直接放行
	svc, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		return fmt.Errorf("服务不存在")
	}
	if svc.OwnerID == userID {
		return nil
	}

	// 查找用户在该服务+集群下的权限记录
	perm, err := s.permRepo.FindByUserAndCluster(serviceID, clusterID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("无权限访问该配置")
		}
		return err
	}

	if model.RoleLevel(perm.Role) >= model.RoleLevel(requiredRole) {
		return nil
	}
	return fmt.Errorf("权限不足，需要 %s 角色", requiredRole)
}

// ListPermissions 列出服务的权限列表
func (s *ConfigPermissionService) ListPermissions(serviceID uint) ([]model.ConfigPermission, error) {
	return s.permRepo.List(serviceID)
}

// GrantPermission 授予权限
func (s *ConfigPermissionService) GrantPermission(serviceID, clusterID, userID uint, role string) error {
	if model.RoleLevel(role) == 0 {
		return fmt.Errorf("无效的角色: %s", role)
	}
	perm := &model.ConfigPermission{
		ServiceID: serviceID,
		ClusterID: clusterID,
		UserID:    userID,
		Role:      role,
	}
	return s.permRepo.Upsert(perm)
}

// RevokePermission 撤销权限
func (s *ConfigPermissionService) RevokePermission(id uint) error {
	return s.permRepo.Delete(id)
}
