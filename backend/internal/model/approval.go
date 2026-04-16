package model

import "time"

// 审批状态常量
const (
	ApprovalStatusPending  = "pending"
	ApprovalStatusApproved = "approved"
	ApprovalStatusRejected = "rejected"
)

// Approval 发布审批
type Approval struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	DeploymentID uint       `gorm:"not null;index" json:"deployment_id"`
	RequesterID  uint       `gorm:"not null;index" json:"requester_id"`
	ApproverID   uint       `gorm:"not null;index" json:"approver_id"`
	Status       string     `gorm:"type:varchar(10);not null" json:"status"`
	Comment      string     `gorm:"type:text" json:"comment,omitempty"`
	DecidedAt    *time.Time `json:"decided_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`

	Deployment *Deployment `gorm:"foreignKey:DeploymentID" json:"deployment,omitempty"`
	Requester  *User       `gorm:"foreignKey:RequesterID" json:"requester,omitempty"`
	Approver   *User       `gorm:"foreignKey:ApproverID" json:"approver,omitempty"`
}
