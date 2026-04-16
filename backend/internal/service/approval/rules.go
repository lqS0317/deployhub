package approval

// ApprovalRule 审批规则判定器
type ApprovalRule struct{}

// NeedsApproval 判断是否需要审批（仅检查全局角色）
// Admin 免审，其他角色需要 Admin 审批
func (r *ApprovalRule) NeedsApproval(globalRole string) bool {
	return globalRole != "admin"
}

// CanApprove 判断指定用户是否有权审批
// 必须是 Admin 且不能审批自己发起的部署
func (r *ApprovalRule) CanApprove(approverRole string, approverID, requesterID uint) bool {
	if approverRole != "admin" {
		return false
	}
	return approverID != requesterID
}
