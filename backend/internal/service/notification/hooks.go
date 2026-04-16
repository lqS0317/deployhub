package notification

// Notifier 通知接口，供其他服务调用
// TODO: 在 build/deploy/approval 服务中注入此接口，触发对应通知：
//   - BuildService.UpdateStatus  → NotifyBuildComplete
//   - DeployService 完成部署      → NotifyDeployResult
//   - ApprovalService 创建审批    → NotifyApprovalRequest
type Notifier interface {
	NotifyBuildComplete(userID, buildID uint, serviceName, status string) error
	NotifyDeployResult(userID, deployID uint, serviceName, status string) error
	NotifyApprovalRequest(approverID, deployID uint, serviceName, requesterName string) error
}

// 编译期校验 NotificationService 实现了 Notifier 接口
var _ Notifier = (*NotificationService)(nil)
