package notification

import (
	"fmt"

	"deployhub/internal/model"
	"deployhub/internal/repository"
)

// NotificationService 通知管理服务
type NotificationService struct {
	repo repository.NotificationRepository
}

// NewNotificationService 创建通知服务
func NewNotificationService(repo repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

// Create 创建通知
func (s *NotificationService) Create(userID uint, nType, title, content, refType string, refID uint) (*model.Notification, error) {
	n := &model.Notification{
		UserID:        userID,
		Type:          nType,
		Title:         title,
		Content:       content,
		ReferenceType: refType,
		ReferenceID:   refID,
	}
	if err := s.repo.Create(n); err != nil {
		return nil, fmt.Errorf("创建通知失败: %w", err)
	}
	return n, nil
}

// List 分页列出用户通知，可按已读状态过滤
func (s *NotificationService) List(userID uint, page, pageSize int, isRead *bool) ([]model.Notification, int64, error) {
	return s.repo.List(userID, page, pageSize, isRead)
}

// MarkRead 标记单条通知为已读，需校验通知归属
func (s *NotificationService) MarkRead(id, userID uint) error {
	n, err := s.repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("通知不存在: %w", err)
	}
	if n.UserID != userID {
		return fmt.Errorf("无权操作此通知")
	}
	n.IsRead = true
	return s.repo.Update(n)
}

// MarkAllRead 标记用户所有未读通知为已读
func (s *NotificationService) MarkAllRead(userID uint) (int64, error) {
	return s.repo.MarkAllRead(userID)
}

// UnreadCount 获取用户未读通知数量
func (s *NotificationService) UnreadCount(userID uint) (int64, error) {
	return s.repo.UnreadCount(userID)
}

// NotifyBuildComplete 发送构建完成通知
func (s *NotificationService) NotifyBuildComplete(userID, buildID uint, serviceName, status string) error {
	title := fmt.Sprintf("构建完成 - %s", serviceName)
	content := fmt.Sprintf("服务 %s 的构建已完成，状态: %s", serviceName, status)
	_, err := s.Create(userID, "build_complete", title, content, "build", buildID)
	return err
}

// NotifyDeployResult 发送部署结果通知
func (s *NotificationService) NotifyDeployResult(userID, deployID uint, serviceName, status string) error {
	title := fmt.Sprintf("部署结果 - %s", serviceName)
	content := fmt.Sprintf("服务 %s 的部署已完成，状态: %s", serviceName, status)
	_, err := s.Create(userID, "deploy_result", title, content, "deployment", deployID)
	return err
}

// NotifyApprovalRequest 发送审批请求通知
func (s *NotificationService) NotifyApprovalRequest(approverID, deployID uint, serviceName, requesterName string) error {
	title := fmt.Sprintf("审批请求 - %s", serviceName)
	content := fmt.Sprintf("%s 请求部署服务 %s，请审批", requesterName, serviceName)
	_, err := s.Create(approverID, "approval_request", title, content, "deployment", deployID)
	return err
}
