package approval

import (
	"errors"
	"time"

	"deployhub/internal/model"
	"deployhub/internal/repository"
	"deployhub/internal/service/notification"
)

var (
	ErrApprovalNotFound = errors.New("审批记录不存在")
	ErrApprovalDecided  = errors.New("该审批已完成，无法重复操作")
	ErrNotAuthorized    = errors.New("无权审批此记录")
	ErrNoAdmins         = errors.New("没有可用的管理员来审批")
)

// ApprovalService 审批引擎服务
type ApprovalService struct {
	approvalRepo    repository.ApprovalRepository
	deployRepo      repository.DeploymentRepository
	userRepo        repository.UserRepository
	rule            *ApprovalRule
	notifDispatcher *notification.Dispatcher
}

// NewApprovalService 创建审批服务实例
func NewApprovalService(
	approvalRepo repository.ApprovalRepository,
	deployRepo repository.DeploymentRepository,
	userRepo repository.UserRepository,
	notifDispatcher *notification.Dispatcher,
) *ApprovalService {
	return &ApprovalService{
		approvalRepo:    approvalRepo,
		deployRepo:      deployRepo,
		userRepo:        userRepo,
		rule:            &ApprovalRule{},
		notifDispatcher: notifDispatcher,
	}
}

// CreateApprovalForAdmins 为所有 Admin 创建审批记录（排除发起人）
func (s *ApprovalService) CreateApprovalForAdmins(deploymentID, requesterID uint) error {
	admins, err := s.userRepo.FindByRole("admin")
	if err != nil {
		return err
	}

	var approvers []model.User
	for _, a := range admins {
		if a.ID != requesterID {
			approvers = append(approvers, a)
		}
	}
	if len(approvers) == 0 {
		return ErrNoAdmins
	}

	for _, admin := range approvers {
		approval := &model.Approval{
			DeploymentID: deploymentID,
			RequesterID:  requesterID,
			ApproverID:   admin.ID,
			Status:       model.ApprovalStatusPending,
		}
		if err := s.approvalRepo.Create(approval); err != nil {
			return err
		}
	}

	if err := s.deployRepo.UpdateStatus(deploymentID, model.DeployStatusPendingApproval); err != nil {
		return err
	}

	// 发送待审批通知
	if s.notifDispatcher != nil {
		if dep, err := s.deployRepo.FindByID(deploymentID); err == nil {
			payload := notification.NotificationPayload{
				Namespace: dep.Namespace,
				ImageTag:  dep.ImageTag,
				DeployID:  deploymentID,
			}
			if requester, err := s.userRepo.FindByID(requesterID); err == nil {
				payload.TriggerUser = requester.Username
			}
			s.notifDispatcher.Dispatch(dep.ServiceID, model.EventApprovalPending, payload)
		}
	}

	return nil
}

// Approve 通过审批
func (s *ApprovalService) Approve(approvalID, approverID uint, comment string) error {
	a, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return ErrApprovalNotFound
	}

	if a.Status != model.ApprovalStatusPending {
		return ErrApprovalDecided
	}

	// 校验审批人：必须是 Admin 且非发起人
	approver, err := s.userRepo.FindByID(approverID)
	if err != nil {
		return ErrNotAuthorized
	}
	if !s.rule.CanApprove(approver.Role, approverID, a.RequesterID) {
		return ErrNotAuthorized
	}

	now := time.Now()
	a.Status = model.ApprovalStatusApproved
	a.Comment = comment
	a.DecidedAt = &now
	if err := s.approvalRepo.Update(a); err != nil {
		return err
	}

	return s.deployRepo.UpdateStatus(a.DeploymentID, model.DeployStatusApproved)
}

// Reject 驳回审批
func (s *ApprovalService) Reject(approvalID, approverID uint, comment string) error {
	a, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return ErrApprovalNotFound
	}

	if a.Status != model.ApprovalStatusPending {
		return ErrApprovalDecided
	}

	approver, err := s.userRepo.FindByID(approverID)
	if err != nil {
		return ErrNotAuthorized
	}
	if !s.rule.CanApprove(approver.Role, approverID, a.RequesterID) {
		return ErrNotAuthorized
	}

	now := time.Now()
	a.Status = model.ApprovalStatusRejected
	a.Comment = comment
	a.DecidedAt = &now
	if err := s.approvalRepo.Update(a); err != nil {
		return err
	}

	return s.deployRepo.UpdateStatus(a.DeploymentID, model.DeployStatusRejected)
}

// GetByID 获取审批详情
func (s *ApprovalService) GetByID(id uint) (*model.Approval, error) {
	a, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return nil, ErrApprovalNotFound
	}
	return a, nil
}

// List 分页查询审批列表
func (s *ApprovalService) List(page, pageSize int, status string, approverID *uint) ([]model.Approval, int64, error) {
	return s.approvalRepo.List(page, pageSize, status, approverID)
}
