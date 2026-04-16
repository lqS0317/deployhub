# 004 - Helm 部署任务列表

## Phase 1: DB Migrations (3 tasks)

- [ ] **T001** 创建 `000022_add_helm_fields_to_services.up/down.sql` — services 表新增 deploy_type(default 'direct'), helm_repo_id FK, helm_chart_path, helm_values_path, helm_release_name, helm_chart_branch
- [ ] **T002** 创建 `000023_add_helm_revision_to_deployments.up/down.sql` — deployments 表新增 helm_revision int nullable
- [ ] **T003** 创建 `000024_create_helm_values.up/down.sql` — helm_values 表 (id, service_id FK, cluster_id FK, content text, version int default 1, updated_by FK, timestamps, unique(service_id,cluster_id))

## Phase 2: GORM Models (3 tasks)

- [ ] **T004** 修改 `model/service.go` — 新增 DeployType, HelmRepoID, HelmChartPath, HelmValuesPath, HelmReleaseName, HelmChartBranch 字段 + HelmRepo *GitRepo 关联
- [ ] **T005** 修改 `model/deployment.go` — 新增 HelmRevision *int 字段
- [ ] **T006** 新建 `model/helm_values.go` — HelmValues 结构体：ID, ServiceID, ClusterID, Content, Version, UpdatedBy, timestamps + Service/Cluster/User 关联

## Phase 3: HelmValues Repository (2 tasks)

- [ ] **T007** 新建 `repository/helm_values_repo.go` — 接口：FindByServiceAndCluster, Upsert, ListByService
- [ ] **T008** 新建 `repository/helm_values_repo_impl.go` — GORM 实现，Upsert 用 ON CONFLICT DO UPDATE + version 自增

## Phase 4: Executor 接口抽取 (3 tasks)

- [ ] **T009** 修改 `service/deploy/executor.go` — 定义 DeployExecutor 接口 (Execute method)
- [ ] **T010** 修改 `service/deploy/executor.go` — 将现有 Executor 重命名为 DirectExecutor，实现 DeployExecutor 接口
- [ ] **T011** 更新所有引用 Executor 的地方 — deploy handler, main.go, watcher, executor_test.go

## Phase 5: HelmExecutor - Job 生成 (4 tasks)

- [ ] **T012** 新建 `service/deploy/helm_job.go` — buildHelmRunnerJob 函数：生成 Job spec（initContainer: git clone, container: helm upgrade）
- [ ] **T013** 新建 `service/deploy/helm_executor.go` — HelmExecutor 结构体 + 构造函数（注入 clientPool, gitRepoRepo, cryptoSvc, helmValuesRepo, deployRepo, wsHub）
- [ ] **T014** HelmExecutor.Execute — 异步：获取 git 凭证 → 创建临时 Secret → 创建 values ConfigMap → 创建 Runner Job
- [ ] **T015** HelmExecutor.ensureGitSecret + ensureValuesConfigMap — 创建临时 K8s 资源

## Phase 6: HelmExecutor - Job Watch + Cleanup (3 tasks)

- [ ] **T016** HelmExecutor.watchHelmJob — 轮询 Job 状态，流式获取 Pod 日志（复用 build executor 模式）
- [ ] **T017** HelmExecutor.collectHelmRevision — Job 成功后从 Pod 日志/helm 输出解析 revision 号
- [ ] **T018** HelmExecutor.cleanup — 删除临时 Secret/ConfigMap，Job 由 TTL 自动清理

## Phase 7: Helm Rollback (2 tasks)

- [ ] **T019** 新建 `service/deploy/helm_rollback.go` — HelmExecutor.Rollback(deployment, service, revision)
- [ ] **T020** Rollback Job spec：helm rollback <release> <revision> --namespace <ns>，监听完成

## Phase 8: HelmValues Service (2 tasks)

- [ ] **T021** 新建 `service/deploy/helm_values.go` — HelmValuesService 结构体 + CRUD 方法
- [ ] **T022** Upsert 逻辑：存在则 content 更新 + version++ + updated_by 更新，不存在则创建

## Phase 9: Deploy Service 修改 (3 tasks)

- [ ] **T023** 修改 `service/deploy/deploy.go` — 注入 DirectExecutor + HelmExecutor
- [ ] **T024** 新增 ExecuteDeploy(id uint) error — 加载 deployment+service → deploy_type 路由到对应 executor
- [ ] **T025** ExecuteDeploy: direct → directExecutor.Execute + watcher.Watch; helm → helmExecutor.Execute

## Phase 10: Handler 层 (6 tasks)

- [ ] **T026** 修改 `handler/service.go` — createServiceRequest/updateServiceRequest 新增 helm 字段，Create/Update 写入 model
- [ ] **T027** 修改 `handler/deploy.go` — 新增 ExecuteDeploy handler (POST /deployments/:id/execute)
- [ ] **T028** 修改 `handler/deploy.go` — 新增 HelmRollback handler (POST /deployments/:id/helm-rollback)
- [ ] **T029** 新建 `handler/helm_values.go` — HelmValuesHandler: ListValues, UpdateValues 端点
- [ ] **T030** 修改 `handler/deploy.go` — DeployHandler 注入 DeployService（含 executor 路由）
- [ ] **T031** RegisterHelmValuesRoutes + 更新 RegisterDeployRoutes

## Phase 11: main.go 注册 (2 tasks)

- [ ] **T032** 修改 `cmd/server/main.go` — 初始化 HelmExecutor, HelmValuesService, HelmValuesHandler, HelmValuesRepo
- [ ] **T033** 注册新路由：helm-values, execute, helm-rollback

## Phase 12: Frontend Types + Hooks (3 tasks)

- [ ] **T034** 修改 `types/index.ts` — Service 新增 helm 字段类型, Deployment 新增 helm_revision, 新增 HelmValues 接口
- [ ] **T035** 新建 `hooks/use-helm-values.ts` — useHelmValues(serviceId), useUpdateHelmValues
- [ ] **T036** 修改 `hooks/use-deployments.ts` — 新增 useExecuteDeploy, useHelmRollback

## Phase 13: Frontend UI (5 tasks)

- [ ] **T037** 修改 `components/service/service-create-dialog.tsx` — 新增部署类型 Radio + Helm 配置区（Helm Repo、Chart Path、Values Path、Release Name、Branch）
- [ ] **T038** 新建 `components/deploy/helm-values-editor.tsx` — YAML 文本域编辑器，按集群 Tab 切换
- [ ] **T039** 修改 `app/(dashboard)/services/[id]/page.tsx` — Helm 类型服务显示 Chart 配置信息 + Values 编辑入口
- [ ] **T040** 修改 `app/(dashboard)/deployments/page.tsx` — helm_revision 列显示 + execute/helm-rollback 按钮
- [ ] **T041** 修改 `components/deploy/deploy-dialog.tsx` — Helm 类型部署对话框适配

## Phase 14: Tests (3 tasks)

- [ ] **T042** 更新 `service/deploy/executor_test.go` — 适配 DirectExecutor 重命名
- [ ] **T043** 更新 `handler/deploy_test.go` — 适配新构造函数签名
- [ ] **T044** 运行 `go test ./...` + `next build` 全量通过

---

**总计: 44 tasks, 14 phases**
