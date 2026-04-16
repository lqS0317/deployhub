package model

import (
	"time"

	"gorm.io/datatypes"
)

// 路由资源类型常量
const (
	RouteTypeService      = "service"
	RouteTypeIngress      = "ingress"
	RouteTypeIngressRoute = "ingressroute"
	RouteTypeApisixRoute  = "apisixroute"
)

// RouteEntry 路由条目
type RouteEntry struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_re_name_type" json:"name"`
	ResourceType string         `gorm:"type:varchar(20);not null;uniqueIndex:idx_re_name_type" json:"resource_type"`
	Config       datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"config"`
	CreatedByID  uint           `gorm:"not null" json:"created_by_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`

	CreatedBy *User `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
}
