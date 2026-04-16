package model

import "time"

// Cluster Kubernetes 集群
type Cluster struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	Name                string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	DisplayName         string    `gorm:"type:varchar(100)" json:"display_name,omitempty"`
	Env                 string    `gorm:"type:varchar(20);not null" json:"env"`
	APIServer           string    `gorm:"type:varchar(500)" json:"api_server,omitempty"`
	KubeconfigEncrypted string    `gorm:"type:text;not null" json:"-"`
	Status              string    `gorm:"type:varchar(10);not null;default:active" json:"status"`
	HelmServiceAccount  string    `gorm:"type:varchar(100);default:''" json:"helm_service_account,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
