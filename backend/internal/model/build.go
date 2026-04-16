package model

import "time"

// 构建状态常量
const (
	BuildStatusPending   = "pending"
	BuildStatusBuilding  = "building"
	BuildStatusSuccess   = "success"
	BuildStatusFailed    = "failed"
	BuildStatusCancelled = "cancelled"
)

// Build 构建记录
type Build struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	ServiceID      uint       `gorm:"not null;index" json:"service_id"`
	TriggerUserID  uint       `gorm:"not null;index" json:"trigger_user_id"`
	GitBranch      string     `gorm:"type:varchar(200);not null" json:"git_branch"`
	GitCommit      string     `gorm:"type:varchar(40)" json:"git_commit,omitempty"`
	ImageTag       string     `gorm:"type:varchar(200)" json:"image_tag,omitempty"`
	Status         string     `gorm:"type:varchar(20);not null" json:"status"`
	BuildClusterID uint       `gorm:"not null;index" json:"build_cluster_id"`
	KanikoJobName  string     `gorm:"type:varchar(200)" json:"kaniko_job_name,omitempty"`
	Name           string     `gorm:"type:varchar(200);default:''" json:"name,omitempty"`
	DockerfilePath string     `gorm:"type:varchar(500);default:'./Dockerfile'" json:"dockerfile_path,omitempty"`
	RegistryID     *uint      `gorm:"index" json:"registry_id,omitempty"`
	ImageRepo      string     `gorm:"type:varchar(500);default:''" json:"image_repo,omitempty"`
	BuildContext   string     `gorm:"type:varchar(500);default:'.'" json:"build_context,omitempty"`
	Log            string     `gorm:"type:text" json:"-"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`

	Service      *Service  `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	TriggerUser  *User     `gorm:"foreignKey:TriggerUserID" json:"trigger_user,omitempty"`
	BuildCluster *Cluster  `gorm:"foreignKey:BuildClusterID" json:"build_cluster,omitempty"`
	Registry     *Registry `gorm:"foreignKey:RegistryID" json:"registry,omitempty"`
}
