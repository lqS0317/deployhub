package model

import "time"

// SystemSetting 系统配置项（key-value 形式，存储在 DB，运行时可动态修改）
type SystemSetting struct {
	Key         string    `gorm:"primaryKey;type:varchar(100)" json:"key"`
	Value       string    `gorm:"type:text;not null;default:''" json:"value"`
	Description string    `gorm:"type:varchar(255);default:''" json:"description,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 预定义的配置 key
const (
	SettingHelmJobNamespace = "helm_job_namespace"
	SettingEnvValuesMap     = "env_values_map"
)
