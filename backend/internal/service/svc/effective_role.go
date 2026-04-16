package svc

import "deployhub/internal/repository"

// PermissionSource 权限来源
type PermissionSource struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	GroupID uint   `json:"group_id,omitempty"`
}

// EffectivePermission 用户有效权限
type EffectivePermission struct {
	ServiceID   uint               `json:"service_id"`
	ServiceName string             `json:"service_name"`
	Role        string             `json:"role"`
	Sources     []PermissionSource `json:"sources"`
}

// EffectiveRoleService 有效权限计算服务
type EffectiveRoleService struct {
	memberRepo    repository.ServiceMemberRepository
	groupPermRepo repository.GroupPermissionRepository
	userRepo      repository.UserRepository
	serviceRepo   repository.ServiceRepository
}

func NewEffectiveRoleService(
	memberRepo repository.ServiceMemberRepository,
	groupPermRepo repository.GroupPermissionRepository,
	userRepo repository.UserRepository,
	serviceRepo repository.ServiceRepository,
) *EffectiveRoleService {
	return &EffectiveRoleService{
		memberRepo:    memberRepo,
		groupPermRepo: groupPermRepo,
		userRepo:      userRepo,
		serviceRepo:   serviceRepo,
	}
}

// GetEffectiveRole 获取用户在指定 Service 上的有效角色（合并个人+组权限取最高值）
func (s *EffectiveRoleService) GetEffectiveRole(userID, serviceID uint) (string, []PermissionSource) {
	// admin 直接返回 owner
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", nil
	}
	if user.Role == "admin" {
		return "owner", []PermissionSource{{Type: "admin", Name: "管理员"}}
	}

	bestRole := ""
	var sources []PermissionSource

	// 个人权限
	personalRole, err := s.memberRepo.GetUserRole(serviceID, userID)
	if err == nil && personalRole != "" {
		bestRole = personalRole
		sources = append(sources, PermissionSource{Type: "personal", Name: "个人权限"})
	}

	// 组权限
	groupPerms, err := s.groupPermRepo.FindRolesByUserAndService(userID, serviceID)
	if err == nil {
		for _, gp := range groupPerms {
			if roleLevel[gp.Role] > roleLevel[bestRole] {
				bestRole = gp.Role
			}
			groupName := ""
			if gp.Group != nil {
				groupName = gp.Group.Name
			}
			sources = append(sources, PermissionSource{
				Type:    "group",
				Name:    groupName,
				GroupID: gp.GroupID,
			})
		}
	}

	return bestRole, sources
}

// CheckPermission 检查用户是否拥有指定 Service 的最低角色权限
func (s *EffectiveRoleService) CheckPermission(userID, serviceID uint, requiredRole string) bool {
	role, _ := s.GetEffectiveRole(userID, serviceID)
	return roleLevel[role] >= roleLevel[requiredRole]
}

// GetAllEffectivePermissions 获取用户的全部有效权限（权限全景）
func (s *EffectiveRoleService) GetAllEffectivePermissions(userID uint) ([]EffectivePermission, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// 收集所有相关 Service ID 及其权限
	permMap := make(map[uint]*EffectivePermission)

	// admin 获取所有 Service
	if user.Role == "admin" {
		services, _, err := s.serviceRepo.List(1, 1000)
		if err == nil {
			for _, svc := range services {
				permMap[svc.ID] = &EffectivePermission{
					ServiceID:   svc.ID,
					ServiceName: svc.Name,
					Role:        "owner",
					Sources:     []PermissionSource{{Type: "admin", Name: "管理员"}},
				}
			}
		}
		return mapToSlice(permMap), nil
	}

	// 个人权限：通过 ServiceMember 查询
	personalPerms, _ := s.memberRepo.ListByUser(userID)
	for _, pm := range personalPerms {
		serviceName := ""
		if pm.Service != nil {
			serviceName = pm.Service.Name
		}
		ep, exists := permMap[pm.ServiceID]
		if !exists {
			ep = &EffectivePermission{
				ServiceID:   pm.ServiceID,
				ServiceName: serviceName,
				Role:        pm.Role,
			}
			permMap[pm.ServiceID] = ep
		} else if roleLevel[pm.Role] > roleLevel[ep.Role] {
			ep.Role = pm.Role
		}
		ep.Sources = append(ep.Sources, PermissionSource{Type: "personal", Name: "个人权限"})
	}

	// 组权限
	groupPerms, _ := s.groupPermRepo.FindAllByUser(userID)
	for _, gp := range groupPerms {
		serviceName := ""
		if gp.Service != nil {
			serviceName = gp.Service.Name
		}
		groupName := ""
		if gp.Group != nil {
			groupName = gp.Group.Name
		}

		ep, exists := permMap[gp.ServiceID]
		if !exists {
			ep = &EffectivePermission{
				ServiceID:   gp.ServiceID,
				ServiceName: serviceName,
				Role:        gp.Role,
			}
			permMap[gp.ServiceID] = ep
		} else if roleLevel[gp.Role] > roleLevel[ep.Role] {
			ep.Role = gp.Role
		}
		ep.Sources = append(ep.Sources, PermissionSource{
			Type:    "group",
			Name:    groupName,
			GroupID: gp.GroupID,
		})
	}

	return mapToSlice(permMap), nil
}

func mapToSlice(m map[uint]*EffectivePermission) []EffectivePermission {
	result := make([]EffectivePermission, 0, len(m))
	for _, v := range m {
		result = append(result, *v)
	}
	return result
}
