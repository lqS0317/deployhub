package deploy

import "deployhub/internal/model"

// ApprovalChecker 审批检查接口，由 approval 服务实现
// 部署服务通过此接口解耦审批逻辑，在创建部署后调用以决定是否需要审批流程
type ApprovalChecker interface {
	CheckAndCreateApproval(deployment *model.Deployment, requesterID uint) (needsApproval bool, err error)
}
