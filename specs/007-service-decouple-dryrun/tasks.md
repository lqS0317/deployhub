# 007 - Service 解耦 + Dry-Run 任务列表

## Phase 1: DB Migrations (2 tasks)

- [ ] **T001** 创建 `000027_decouple_service_cluster.up/down.sql` — services: cluster_id DROP NOT NULL, namespace DROP NOT NULL SET DEFAULT ''
- [ ] **T002** 创建 `000028_add_preview_fields_to_deployments.up/down.sql` — deployments: preview_yaml text, preview_summary jsonb

## Phase 2: Model 更新 (3 tasks)

- [ ] **T003** 修改 `model/service.go` — ClusterID 改为 `*uint`（nullable），移除 GORM not null 约束
- [ ] **T004** 修改 `model/service.go` — Namespace 改为 `string` default ''，移除 not null 约束
- [ ] **T005** 修改 `model/deployment.go` — 新增 PreviewYAML *string + PreviewSummary datatypes.JSON + 3 个新状态常量

## Phase 3: Repository 层 (2 tasks)

- [ ] **T006** 修改 `repository/service_repo.go` — List 签名移除 clusterID *uint, namespace string 参数
- [ ] **T007** 修改 `repository/service_repo_impl.go` — List 实现移除 cluster_id/namespace WHERE 条件

## Phase 4: Service 层 (5 tasks)

- [ ] **T008** 修改 `service/svc/svc.go` — ServiceService.List 移除 clusterID/namespace 参数
- [ ] **T009** 修改 `service/deploy/deploy.go` — CreateDeployment 新增 clusterID + namespace 参数（不再从 svc 取）
- [ ] **T010** 修改 `service/deploy/deploy.go` — 新增 StartPreview(id) 方法（状态转为 previewing）
- [ ] **T011** 修改 `service/deploy/deploy.go` — 新增 SavePreview(id, yaml, summary) + GetPreview(id) 方法
- [ ] **T012** 修改 `service/deploy/deploy.go` — 新增 CancelDeploy(id) 方法（previewed/pending_approval → cancelled）

## Phase 5: Handler 层 — Service (3 tasks)

- [ ] **T013** 修改 `handler/service.go` — createServiceRequest: cluster_id 改为可选，namespace 改为可选
- [ ] **T014** 修改 `handler/service.go` — Create handler: ClusterID/Namespace 用指针/默认值处理
- [ ] **T015** 修改 `handler/service.go` — List handler: 移除 cluster_id/namespace query 参数

## Phase 6: Handler 层 — Deploy (5 tasks)

- [ ] **T016** 修改 `handler/deploy.go` — createDeploymentRequest 新增 ClusterID uint + Namespace string (required)
- [ ] **T017** 修改 `handler/deploy.go` — Create handler: 传 req.ClusterID + req.Namespace 给 CreateDeployment
- [ ] **T018** 新增 `handler/deploy.go` — Preview handler (POST /:id/preview): 触发 dry-run
- [ ] **T019** 新增 `handler/deploy.go` — GetPreview handler (GET /:id/preview): 返回预览结果
- [ ] **T020** 新增 `handler/deploy.go` — Cancel handler (POST /:id/cancel): 取消部署

## Phase 7: Direct Dry-Run (3 tasks)

- [ ] **T021** 新增 `executor.go` — DirectExecutor.DryRun(deployment, service) → (yaml string, error)
- [ ] **T022** DryRun 使用 K8s server-side dry-run API: metav1.CreateOptions{DryRun: []string{"All"}}
- [ ] **T023** 将 dry-run 返回的资源序列化为 YAML

## Phase 8: Helm Dry-Run (3 tasks)

- [ ] **T024** 新增 `helm_job.go` — buildHelmTemplateJob: 生成 helm template Runner Job spec
- [ ] **T025** 新增 `helm_executor.go` — HelmExecutor.Preview(deployment, service) → 创建 helm template Job
- [ ] **T026** Preview Job Watch: 收集 Pod stdout 作为 preview_yaml

## Phase 9: Preview 解析 + 脱敏 (3 tasks)

- [ ] **T027** 新建 `service/deploy/preview.go` — ParsePreviewYAML(yaml) → PreviewSummary (资源列表 + 关键变更)
- [ ] **T028** SanitizeSecrets(yaml) — 替换 Secret data 值为 "***"
- [ ] **T029** preview_test.go — 测试解析 + 脱敏

## Phase 10: 状态机集成 (3 tasks)

- [ ] **T030** 修改 `handler/deploy.go` — Preview handler 调 StartPreview → DryRun/Preview → SavePreview
- [ ] **T031** 修改 `handler/deploy.go` — ExecuteDeploy 改为仅接受 previewed 状态（向后兼容也接受 approved）
- [ ] **T032** 修改 `cmd/server/main.go` — 注册 preview/cancel 路由

## Phase 11: Build ClusterID 解耦 (1 task)

- [ ] **T033** 修改 `service/build/build.go` — buildClusterID=0 时返回错误而非 fallback svc.ClusterID

## Phase 12: Test 修复 (4 tasks)

- [ ] **T034** 修改 `service/deploy/deploy_test.go` — CreateDeployment 签名更新，mock service 不再需要 ClusterID
- [ ] **T035** 修改 `handler/deploy_test.go` — 适配新签名
- [ ] **T036** 修改 `service/svc/svc_test.go` — List 签名更新
- [ ] **T037** 修改 `handler/service.go` 相关测试 — 移除 cluster_id 断言

## Phase 13: Frontend (7 tasks)

- [ ] **T038** 修改 `types/index.ts` — Service.cluster_id/namespace 改为 optional，Deployment 新增 preview 字段
- [ ] **T039** 修改 `service-create-dialog.tsx` — cluster_id/namespace 改为可选，非必填
- [ ] **T040** 修改 `services/page.tsx` — 移除集群/命名空间列
- [ ] **T041** 修改 `use-services.ts` — List 移除 cluster_id/namespace 查询参数
- [ ] **T042** 修改 deploy 创建对话框 — 新增 cluster_id 下拉 + namespace 输入（必填）
- [ ] **T043** 新建 deploy 预览面板组件 — YAML 展示 + 资源摘要 + 确认/取消按钮
- [ ] **T044** 新增 hooks — usePreviewDeploy, useCancelDeploy

## Phase 14: Verification (1 task)

- [ ] **T045** 运行 go test ./... + next build 全量通过

---

**总计: 45 tasks, 14 phases**

**风险提示**: 这是一个破坏性 API 变更（创建部署新增必填字段），需要前后端同步上线。建议分两步部署：
1. 先上后端（兼容模式：cluster_id 可选，优先取请求值，fallback svc 值）
2. 再上前端（新表单带 cluster_id/namespace）
