package model

import (
	"time"

	"gorm.io/datatypes"
)

// 发布状态常量
const (
	ReleaseStatusPublished  = "published"
	ReleaseStatusRolledBack = "rolled_back"
)

// ConfigRelease 配置发布记录（全量快照版本化，归属配置条目）
type ConfigRelease struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ConfigEntryID uint           `gorm:"not null;index:idx_cr_lookup" json:"config_entry_id"`
	Version       int            `gorm:"not null;default:1" json:"version"`
	Snapshot      datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"snapshot"`
	Status        string         `gorm:"type:varchar(20);not null;default:published" json:"status"`
	Comment       string         `gorm:"type:varchar(500);default:''" json:"comment,omitempty"`
	CreatedByID   uint           `gorm:"not null" json:"created_by_id"`
	CreatedAt     time.Time      `json:"created_at"`

	CreatedBy *User `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
}
