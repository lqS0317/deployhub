# 010 - 发布管理增强任务列表

## Phase 1: DB Migration + Model (2 tasks)

- [ ] **T001** 创建 `000035_add_fail_reason_to_deployments.up/down.sql` — deployments: fail_reason text nullable
- [ ] **T002** 修改 `model/deployment.go` — 新增 FailReason *string 字段

## Phase 2: Repository (2 tasks)

- [ ] **T003** 修改 `repository/deployment_repo.go` — 接口新增 Delete(id) + UpdateStatusWithReason(id, status, reason)
- [ ] **T004** 修改 `repository/deployment_repo_impl.go` — GORM 实现（Delete 硬删 + 关联审批，UpdateStatusWithReason 更新 status+fail_reason）

## Phase 3: Deploy Service (3 tasks)

- [ ] **T005** 新增 `deploy.go` UpdateStatusWithReason(id, status, reason) — 调 repo 层
- [ ] **T006** 新增 `deploy.go` UpdateDeployment(id, fields) — 校验 pending_approval 状态，更新字段，清空预览，重置状态
- [ ] **T007** 新增 `deploy.go` DeleteDeployment(id) — 校验可删除状态，删除关联审批，硬删

## Phase 4: Handler 编辑/删除 (3 tasks)

- [ ] **T008** 新增 `handler/deploy.go` Update handler — PUT /deployments/:id，校验状态，更新后自动重新预览
- [ ] **T009** 新增 `handler/deploy.go` Delete handler — DELETE /deployments/:id，二次确认由前端控制
- [ ] **T010** 注册路由 — deployments.PUT("/:id") + deployments.DELETE("/:id")

## Phase 5: Fail Reason 集成 (4 tasks)

- [ ] **T011** 修改 `watcher.go` markFailed — 接受 reason 参数，调 UpdateStatusWithReason
- [ ] **T012** 修改 `handler/deploy.go` autoPreview — 预览失败时写 fail_reason
- [ ] **T013** 修改 `helm_executor.go` failDeploy — 写 fail_reason
- [ ] **T014** 修改 `handler/deploy.go` ExecuteDeploy — executor 失败时写 fail_reason

## Phase 6: Pod API (2 tasks)

- [ ] **T015** 新增 `handler/deploy.go` ListPods — GET /deployments/:id/pods，通过 client-go label selector 查 app={serviceName}
- [ ] **T016** 注册路由 — deployments.GET("/:id/pods")

## Phase 7: Pod Log WS (3 tasks)

- [ ] **T017** 新增 `handler/ws.go` HandlePodLogWS — 解析 pod/container/tail 参数
- [ ] **T018** HandlePodLogWS — 校验 Pod 状态（Pending/Waiting → 返回错误，Running → Follow 日志）
- [ ] **T019** 注册 WS 路由 — ws /deployments/:id/pod-logs

## Phase 8: Frontend hooks (2 tasks)

- [ ] **T020** 新增 `hooks/use-deployments.ts` — useUpdateDeploy, useDeleteDeploy
- [ ] **T021** 新增 `hooks/use-pod-logs.ts` — useDeployPods, WS Pod 日志连接

## Phase 9: Frontend 编辑/删除 (3 tasks)

- [ ] **T022** 修改 `deployments/page.tsx` — pending_approval 增加「编辑」按钮，所有状态增加「删除」按钮
- [ ] **T023** 新建 `components/deploy/deploy-edit-dialog.tsx` — 复用 deploy-dialog 预填当前值
- [ ] **T024** 修改 `deployments/[id]/page.tsx` — 编辑/删除按钮

## Phase 10: Frontend 失败原因 (2 tasks)

- [ ] **T025** 修改 `deployments/page.tsx` — failed 行 hover tooltip 展示 fail_reason
- [ ] **T026** 修改 `deployments/[id]/page.tsx` — 失败状态红色告警卡片展示完整 fail_reason

## Phase 11: Frontend 日志面板 (3 tasks)

- [ ] **T027** 新建 `components/deploy/deploy-log-panel.tsx` — 终端风格日志面板，接入 deployment:{id} WS room
- [ ] **T028** 新建 `components/deploy/pod-log-panel.tsx` — Pod/Container 选择 + WS 实时日志
- [ ] **T029** 修改 `deployments/[id]/page.tsx` — 新增「事件日志」和「容器日志」Tab

## Phase 12: Tests + Verification (2 tasks)

- [ ] **T030** 更新 deploy handler/service 测试 — 适配新方法
- [ ] **T031** 运行 go test ./... + next build 全量通过

---

**总计: 31 tasks, 12 phases**
