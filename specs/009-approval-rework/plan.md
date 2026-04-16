# 009 - 审批流程重构实现计划

## Layer 1: ApprovalRule 简化

### 受影响文件
- **修改** `backend/internal/service/approval/rules.go` — NeedsApproval 简化为仅检查全局角色，CanApprove 校验 Admin + 非发起人

## Layer 2: UserRepository 扩展

### 受影响文件
- **修改** `backend/internal/repository/user_repo.go` — 新增 FindByRole(role) 方法
- **修改** `backend/internal/repository/user_repo_impl.go` — GORM 实现

## Layer 3: ApprovalService 重写

### 受影响文件
- **修改** `backend/internal/service/approval/approval.go` — 注入 userRepo，重写 CheckAndCreateApproval 为全局 Admin 审批，Approve 增加 Admin 校验

## Layer 4: DeployService 状态机调整

### 受影响文件
- **修改** `backend/internal/service/deploy/deploy.go` — CreateDeployment 初始状态改为 previewing，新增 OnPreviewComplete 回调方法
- **修改** `backend/internal/handler/deploy.go` — Create 后自动触发预览，Preview 完成回调自动审批/放行

## Layer 5: Helm 异步预览集成

### 受影响文件
- **修改** `backend/internal/handler/deploy.go` — Helm 异步预览 goroutine 完成后调 OnPreviewComplete
- **修改** `backend/internal/service/deploy/deploy.go` — OnPreviewComplete 查发起人角色，Admin 放行 / 非 Admin 创建审批

## Layer 6: Handler 层整合

### 受影响文件
- **修改** `backend/internal/handler/deploy.go` — ExecuteDeploy 仅接受 approved，移除手动 preview 端点触发
- **修改** `backend/internal/handler/approval.go` — 审批详情关联返回 preview_yaml
- **修改** `backend/cmd/server/main.go` — 注入 approvalSvc + userRepo 到 DeployHandler

## Layer 7: Frontend

### 受影响文件
- **修改** `frontend/src/app/(dashboard)/deployments/page.tsx` — 去掉手动预览按钮，状态显示更新
- **修改** `frontend/src/app/(dashboard)/deployments/[id]/page.tsx` — previewing 显示等待，pending_approval 显示审批信息
- **修改** `frontend/src/app/(dashboard)/approvals/page.tsx` — 审批列表链接到部署详情

## Layer 8: Tests + Verification

- 更新 approval 测试
- 更新 deploy handler 测试
- go test ./... + next build
