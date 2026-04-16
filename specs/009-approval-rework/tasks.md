# 009 - 审批流程重构任务列表

## Phase 1: ApprovalRule 简化 (2 tasks)

- [ ] **T001** 修改 `approval/rules.go` — NeedsApproval(globalRole) 简化：admin 免审，其他需审
- [ ] **T002** 修改 `approval/rules.go` — CanApprove(approverRole, approverID, requesterID) 校验：必须 admin 且非发起人

## Phase 2: UserRepository 扩展 (2 tasks)

- [ ] **T003** 修改 `repository/user_repo.go` — 接口新增 FindByRole(role string) ([]User, error)
- [ ] **T004** 修改 `repository/user_repo_impl.go` — GORM 实现 FindByRole

## Phase 3: ApprovalService 重写 (4 tasks)

- [ ] **T005** 修改 `approval/approval.go` — NewApprovalService 注入 userRepo
- [ ] **T006** 重写 CheckAndCreateApprovalForAdmins — 查全局 Admin（排除发起人），为每个创建 Approval 记录
- [ ] **T007** 修改 Approve — 增加校验审批人是 Admin
- [ ] **T008** 修改 `cmd/server/main.go` — 将 userRepo 传入 NewApprovalService

## Phase 4: DeployService 状态机 (3 tasks)

- [ ] **T009** 修改 `deploy/deploy.go` CreateDeployment — 初始状态改为 previewing
- [ ] **T010** 新增 `deploy/deploy.go` OnPreviewComplete(deploymentID, previewYAML, summaryJSON, triggerUserID) — 查用户角色，admin 放行 / 非 admin 创建审批
- [ ] **T011** 修改 `deploy/deploy.go` SavePreview — 调用 OnPreviewComplete

## Phase 5: DeployHandler 集成 (5 tasks)

- [ ] **T012** 修改 `handler/deploy.go` DeployHandler — 注入 approvalSvc + userRepo
- [ ] **T013** 修改 `handler/deploy.go` Create — 创建后自动触发预览（Direct 同步，Helm 异步）
- [ ] **T014** Helm 异步预览 goroutine — 完成后调 deploySvc.SavePreview（内部触发审批回调）
- [ ] **T015** 修改 ExecuteDeploy — 仅接受 approved 状态
- [ ] **T016** 修改 `cmd/server/main.go` — 注入 approvalSvc + userRepo 到 DeployHandler

## Phase 6: 审批 Handler 增强 (2 tasks)

- [ ] **T017** 修改 `handler/approval.go` Get — 关联加载 deployment.preview_yaml/preview_summary 返回
- [ ] **T018** 修改 `handler/approval.go` Approve — 校验 Admin 角色

## Phase 7: Frontend (4 tasks)

- [ ] **T019** 修改 deployments/page.tsx — 去掉手动预览按钮，approved 状态显示「执行」按钮
- [ ] **T020** 修改 deployments/[id]/page.tsx — previewing 显示加载态，pending_approval 显示审批信息和审批状态
- [ ] **T021** 修改 approvals/page.tsx — 审批项点击跳转到部署详情（或展开显示预览）
- [ ] **T022** 前端状态标签更新 — 确保所有状态颜色和文案正确

## Phase 8: Tests + Verification (2 tasks)

- [ ] **T023** 更新 approval_test.go — 适配新签名和逻辑
- [ ] **T024** 运行 go test ./... + next build 全量通过

---

**总计: 24 tasks, 8 phases**
