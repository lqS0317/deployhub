package model

import "time"

// 配置权限角色常量
const (
	ConfigRoleViewer    = "viewer"
	ConfigRoleEditor    = "editor"
	ConfigRolePublisher = "publisher"
)

// ConfigPermission 配置权限（按服务+环境粒度控制）
type ConfigPermission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ServiceID uint      `gorm:"not null;uniqueIndex:idx_cp_svc_cluster_user" json:"service_id"`
	ClusterID uint      `gorm:"not null;uniqueIndex:idx_cp_svc_cluster_user;default:0" json:"cluster_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_cp_svc_cluster_user" json:"user_id"`
	Role      string    `gorm:"type:varchar(20);not null;default:viewer" json:"role"`
	CreatedAt time.Time `json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// RoleLevel 角色权限等级
func RoleLevel(role string) int {
	switch role {
	case ConfigRolePublisher:
		return 3
	case ConfigRoleEditor:
		return 2
	case ConfigRoleViewer:
		return 1
	default:
		return 0
	}
}
