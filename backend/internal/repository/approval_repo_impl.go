package repository

import (
	"deployhub/internal/model"

	"gorm.io/gorm"
)

type approvalRepository struct {
	db *gorm.DB
}

// NewApprovalRepository 创建审批记录仓储实例
func NewApprovalRepository(db *gorm.DB) ApprovalRepository {
	return &approvalRepository{db: db}
}

func (r *approvalRepository) Create(approval *model.Approval) error {
	return r.db.Create(approval).Error
}

func (r *approvalRepository) FindByID(id uint) (*model.Approval, error) {
	var approval model.Approval
	err := r.db.Preload("Deployment").Preload("Requester").Preload("Approver").
		First(&approval, id).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}

func (r *approvalRepository) Update(approval *model.Approval) error {
	return r.db.Save(approval).Error
}

func (r *approvalRepository) List(page, pageSize int, status string, approverID *uint) ([]model.Approval, int64, error) {
	var approvals []model.Approval
	var total int64

	query := r.db.Model(&model.Approval{})

	// 按状态过滤（空字符串或 "all" 表示不过滤）
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	// 按审批人过滤
	if approverID != nil {
		query = query.Where("approver_id = ?", *approverID)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Deployment").Preload("Requester").Preload("Approver").
		Offset(offset).Limit(pageSize).Order("id DESC").Find(&approvals).Error
	return approvals, total, err
}

func (r *approvalRepository) FindByDeployment(deploymentID uint) ([]model.Approval, error) {
	var approvals []model.Approval
	err := r.db.Where("deployment_id = ?", deploymentID).
		Preload("Requester").Preload("Approver").
		Find(&approvals).Error
	return approvals, err
}

func (r *approvalRepository) FindPendingByDeployment(deploymentID uint) (*model.Approval, error) {
	var approval model.Approval
	err := r.db.Where("deployment_id = ? AND status = ?", deploymentID, model.ApprovalStatusPending).
		First(&approval).Error
	if err != nil {
		return nil, err
	}
	return &approval, nil
}
