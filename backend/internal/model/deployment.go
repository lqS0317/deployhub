package model

import (
	"time"

	"gorm.io/datatypes"
)

// 部署状态常量
const (
	DeployStatusPendingApproval = "pending_approval"
	DeployStatusApproved        = "approved"
	DeployStatusPreviewing      = "previewing"
	DeployStatusPreviewed       = "previewed"
	DeployStatusDeploying       = "deploying"
	DeployStatusSuccess         = "success"
	DeployStatusFailed          = "failed"
	DeployStatusRolledBack      = "rolled_back"
	DeployStatusRejected        = "rejected"
	DeployStatusExpired         = "expired"
	DeployStatusCancelled       = "cancelled"
	DeployStatusPodChecking     = "pod_checking"
	DeployStatusPodHealthy      = "pod_healthy"
	DeployStatusPodUnhealthy    = "pod_unhealthy"
)

// Deployment 部署记录
type Deployment struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	ServiceID        uint           `gorm:"not null;index" json:"service_id"`
	BuildID          *uint          `gorm:"index" json:"build_id,omitempty"`
	TriggerUserID    uint           `gorm:"not null;index" json:"trigger_user_id"`
	ClusterID        uint           `gorm:"not null;index" json:"cluster_id"`
	Namespace        string         `gorm:"type:varchar(100);not null" json:"namespace"`
	ImageTag         string         `gorm:"type:varchar(200);not null" json:"image_tag"`
	Status           string         `gorm:"type:varchar(20);not null" json:"status"`
	PreviousImageTag string         `gorm:"type:varchar(200)" json:"previous_image_tag,omitempty"`
	IsRollback       bool           `gorm:"not null;default:false" json:"is_rollback"`
	RollbackFromID   *uint          `gorm:"index" json:"rollback_from_id,omitempty"`
	HelmRevision     *int           `json:"helm_revision,omitempty"`
	ImageSource      string         `gorm:"type:varchar(20);not null;default:build" json:"image_source"`
	ExternalImage    string         `gorm:"type:varchar(500);default:''" json:"external_image,omitempty"`
	FailReason       string         `gorm:"type:text;default:''" json:"fail_reason,omitempty"`
	PreviewYAML      *string        `gorm:"type:text" json:"preview_yaml,omitempty"`
	PreviewSummary   datatypes.JSON `gorm:"type:jsonb" json:"preview_summary,omitempty"`
	DeployType       string         `gorm:"type:varchar(10);default:direct" json:"deploy_type,omitempty"`
	WorkloadType     string         `gorm:"type:varchar(20);default:deployment" json:"workload_type,omitempty"`
	HealthCheckPath  string         `gorm:"type:varchar(200);default:''" json:"health_check_path,omitempty"`
	HelmRepoID       *uint          `gorm:"index" json:"helm_repo_id,omitempty"`
	HelmChartPath    string         `gorm:"type:varchar(255);default:''" json:"helm_chart_path,omitempty"`
	HelmReleaseName  string         `gorm:"type:varchar(100);default:''" json:"helm_release_name,omitempty"`
	HelmChartBranch  string         `gorm:"type:varchar(100);default:main" json:"helm_chart_branch,omitempty"`
	HelmServiceAccount string       `gorm:"type:varchar(100);default:''" json:"helm_service_account,omitempty"`
	DeployCommand    string         `gorm:"type:text;default:''" json:"deploy_command,omitempty"`
	PodStatus        string         `gorm:"type:varchar(20);default:''" json:"pod_status,omitempty"`
	PodMessage       string         `gorm:"type:text;default:''" json:"pod_message,omitempty"`
	StartedAt        *time.Time     `json:"started_at,omitempty"`
	FinishedAt       *time.Time     `json:"finished_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`

	Service      *Service    `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Build        *Build      `gorm:"foreignKey:BuildID" json:"build,omitempty"`
	TriggerUser  *User       `gorm:"foreignKey:TriggerUserID" json:"trigger_user,omitempty"`
	Cluster      *Cluster    `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	RollbackFrom *Deployment `gorm:"foreignKey:RollbackFromID" json:"rollback_from,omitempty"`
	HelmRepo     *GitRepo    `gorm:"foreignKey:HelmRepoID" json:"helm_repo,omitempty"`
}
