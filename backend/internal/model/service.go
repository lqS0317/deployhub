package model

import (
	"time"

	"gorm.io/datatypes"
)

// Service 服务（核心实体）
type Service struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	DisplayName     string         `gorm:"type:varchar(200)" json:"display_name,omitempty"`
	Description     string         `gorm:"type:text" json:"description,omitempty"`
	GitRepoID       uint           `gorm:"not null;index" json:"git_repo_id"`
	GitBranch       string         `gorm:"type:varchar(200);not null;default:main" json:"git_branch"`
	DockerfilePath  string         `gorm:"type:varchar(500);not null;default:./Dockerfile" json:"dockerfile_path"`
	RegistryID      *uint          `gorm:"index" json:"registry_id,omitempty"`
	ImageRepo       string         `gorm:"type:varchar(500);default:''" json:"image_repo,omitempty"`
	ClusterID       *uint          `gorm:"index:idx_service_cluster_namespace" json:"cluster_id,omitempty"`
	Namespace       string         `gorm:"type:varchar(100);default:'';index:idx_service_cluster_namespace" json:"namespace,omitempty"`
	Replicas        int            `gorm:"not null;default:1" json:"replicas"`
	CPURequest      string         `gorm:"type:varchar(20)" json:"cpu_request,omitempty"`
	MemRequest      string         `gorm:"type:varchar(20)" json:"mem_request,omitempty"`
	CPULimit        string         `gorm:"type:varchar(20)" json:"cpu_limit,omitempty"`
	MemLimit        string         `gorm:"type:varchar(20)" json:"mem_limit,omitempty"`
	Port            int            `gorm:"default:0" json:"port"`
	HealthCheckPath string         `gorm:"type:varchar(200)" json:"health_check_path,omitempty"`
	EnvVars         datatypes.JSON `gorm:"type:jsonb" json:"env_vars,omitempty"`
	Volumes         datatypes.JSON `gorm:"type:jsonb" json:"volumes,omitempty"`
	OwnerID         uint           `gorm:"not null;index" json:"owner_id"`
	ServiceType     string         `gorm:"type:varchar(50);default:''" json:"service_type,omitempty"`
	Language        string         `gorm:"type:varchar(50);default:''" json:"language,omitempty"`
	LanguageVersion string         `gorm:"type:varchar(50);default:''" json:"language_version,omitempty"`
	DeployType      string         `gorm:"type:varchar(10);not null;default:direct" json:"deploy_type"`
	WorkloadType    string         `gorm:"type:varchar(20);not null;default:deployment" json:"workload_type"`
	HelmRepoID      *uint          `gorm:"index" json:"helm_repo_id,omitempty"`
	HelmChartPath   string         `gorm:"type:varchar(255);default:''" json:"helm_chart_path,omitempty"`
	HelmValuesPath  string         `gorm:"type:varchar(255);default:''" json:"helm_values_path,omitempty"`
	HelmReleaseName string         `gorm:"type:varchar(100);default:''" json:"helm_release_name,omitempty"`
	HelmChartBranch string         `gorm:"type:varchar(100);default:main" json:"helm_chart_branch,omitempty"`
	HelmEnvFilePath            string         `gorm:"type:varchar(255);default:''" json:"helm_env_file_path,omitempty"`
	DefaultPort                int            `gorm:"default:0" json:"default_port"`
	DefaultReplicas            int            `gorm:"default:1" json:"default_replicas"`
	DefaultCPURequest          string         `gorm:"type:varchar(20);default:''" json:"default_cpu_request,omitempty"`
	DefaultMemRequest          string         `gorm:"type:varchar(20);default:''" json:"default_mem_request,omitempty"`
	DefaultCPULimit            string         `gorm:"type:varchar(20);default:''" json:"default_cpu_limit,omitempty"`
	DefaultMemLimit            string         `gorm:"type:varchar(20);default:''" json:"default_mem_limit,omitempty"`
	DefaultCommand             datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"default_command,omitempty"`
	DefaultArgs                datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"default_args,omitempty"`
	DefaultWorkloadType        string         `gorm:"type:varchar(20);default:deployment" json:"default_workload_type,omitempty"`
	DefaultLivenessProbe       datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"default_liveness_probe,omitempty"`
	DefaultReadinessProbe      datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"default_readiness_probe,omitempty"`
	CreatedAt                  time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	GitRepo  *GitRepo  `gorm:"foreignKey:GitRepoID" json:"git_repo,omitempty"`
	Registry *Registry `gorm:"foreignKey:RegistryID" json:"registry,omitempty"`
	Cluster  *Cluster  `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	Owner    *User     `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	HelmRepo *GitRepo  `gorm:"foreignKey:HelmRepoID" json:"helm_repo,omitempty"`
}
