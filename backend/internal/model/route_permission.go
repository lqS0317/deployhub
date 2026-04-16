package model

import "time"

// RoutePermission 路由权限（按集群维度控制）
type RoutePermission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClusterID uint      `gorm:"not null;uniqueIndex:idx_rp_cluster_user;default:0" json:"cluster_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_rp_cluster_user" json:"user_id"`
	Role      string    `gorm:"type:varchar(20);not null;default:viewer" json:"role"`
	CreatedAt time.Time `json:"created_at"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
