package repository

import "deployhub/internal/model"

// ApprovalRepository 审批记录数据访问接口
type ApprovalRepository interface {
	Create(approval *model.Approval) error
	FindByID(id uint) (*model.Approval, error)
	Update(approval *model.Approval) error
	List(page, pageSize int, status string, approverID *uint) ([]model.Approval, int64, error)
	FindByDeployment(deploymentID uint) ([]model.Approval, error)
	FindPendingByDeployment(deploymentID uint) (*model.Approval, error)
}
