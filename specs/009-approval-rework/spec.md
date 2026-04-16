# 009 - 审批流程重构：预览前置 + Admin 唯一审批

## 1. Overview

将 dry-run 预览从审批之后移到审批之前。创建部署后自动触发预览，预览完成后根据发起人角色自动决定是否需要审批。Admin 自动放行，非 Admin 需要 Admin 审批（审批页面可查看预览结果）。

## 2. Current State

- `ApprovalService` 已实现但**未集成**到部署流程
- `CheckAndCreateApproval` 使用 ServiceMember Owner 作为审批人
- `NeedsApproval` 规则：admin/owner 免审，其他需审
- 部署创建后直接设为 `approved`，跳过审批
- 预览是手动触发的独立操作

## 3. New State Machine

```
创建 → previewing → previewed → [自动判断]
                                    ├── Admin 创建 → approved → deploying → success/failed
                                    └── 非 Admin → pending_approval → approved → deploying → success/failed
                                                         ↓
                                                      rejected
```

## 4. Changes

### 4.1 DeployService.CreateDeployment

创建后初始状态改为 `previewing`（不是 `approved`），自动触发预览。

### 4.2 Preview 完成回调

预览完成后（`SavePreview` 内部或新方法）：
1. 查发起人的全局角色
2. Admin → 直接设为 `approved`
3. 非 Admin → 设为 `pending_approval`，调用 `CheckAndCreateApproval` 为所有全局 Admin 创建审批记录

### 4.3 ApprovalRule 简化

```go
// NeedsApproval: 仅 Admin 免审
func NeedsApproval(globalRole string) bool {
    return globalRole != "admin"
}

// CanApprove: 必须是 Admin 且非发起人
func CanApprove(approverRole string, approverID, requesterID uint) bool {
    return approverRole == "admin" && approverID != requesterID
}
```

### 4.4 CheckAndCreateApproval 重写

不再查 ServiceMember Owner，改为查全局 Admin 用户列表：

```go
func (s *ApprovalService) CheckAndCreateApprovalForAdmins(deployment *model.Deployment, requesterID uint) error {
    // 1. 查所有 Admin 用户（排除发起人）
    admins = userRepo.ListAdmins() exclude requesterID
    // 2. 为每个 Admin 创建 Approval 记录
    // 3. 更新部署状态为 pending_approval
}
```

### 4.5 Handler 改动

- `DeployHandler.Create` → 创建后自动触发预览（异步）
- `Preview 完成回调` → 自动判断审批/放行
- `ExecuteDeploy` → 只接受 `approved` 状态
- 审批详情页 → 展示 `preview_yaml` 和 `preview_summary`

### 4.6 前端改动

- **部署列表**：去掉手动「预览」按钮（创建后自动预览）
- **部署详情**：`previewing` 显示"预览中，请稍候"；`pending_approval` 显示审批信息
- **审批详情页**：显示 deployment 的 `preview_yaml` 和 `preview_summary`
- **审批列表**：链接到部署详情
- **创建部署后**：前端显示"正在预览..."提示

## 5. API Changes

| 变更 | 说明 |
|------|------|
| POST /deployments | 创建后自动触发预览，返回 previewing 状态 |
| POST /deployments/:id/preview | **移除**（自动触发，不再手动） |
| POST /approvals/:id/approve | 增加校验：必须是 Admin 且非发起人 |
| GET /approvals/:id | 关联返回 deployment.preview_yaml |

## 6. Success Criteria

- [ ] 创建部署后自动触发预览
- [ ] Admin 部署预览完成后自动放行
- [ ] 非 Admin 部署预览完成后自动创建审批记录
- [ ] 审批页面展示预览结果（YAML + 摘要）
- [ ] 仅 Admin 可审批，发起人不能审批自己的部署
- [ ] 审批通过后状态变为 approved，可执行
- [ ] 审批拒绝后状态变为 rejected

## 7. Non-Goals

- 多人会签
- 自定义审批链
- 回滚审批
- 审批超时
- 审批通知
