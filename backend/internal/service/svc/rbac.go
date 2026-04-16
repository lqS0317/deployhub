package svc

// 角色权限等级
var roleLevel = map[string]int{
	"viewer":    1,
	"developer": 2,
	"owner":     3,
}

// RBACService 服务级别权限检查（委托 EffectiveRoleService）
type RBACService struct {
	effectiveRoleSvc *EffectiveRoleService
}

func NewRBACService(effectiveRoleSvc *EffectiveRoleService) *RBACService {
	return &RBACService{effectiveRoleSvc: effectiveRoleSvc}
}

// CheckPermission 检查用户是否拥有指定服务的最低角色权限
func (s *RBACService) CheckPermission(serviceID, userID uint, requiredRole string) bool {
	return s.effectiveRoleSvc.CheckPermission(userID, serviceID, requiredRole)
}
