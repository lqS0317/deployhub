# 008 - 核心实体职责重构任务列表

## Phase 1: DB Migrations (4 tasks)

- [ ] **T001** 创建 `000029_add_service_meta_fields.up/down.sql` — services: service_type varchar(50) default '', language varchar(50) default '', language_version varchar(50) default ''
- [ ] **T002** 创建 `000030_add_build_config_fields.up/down.sql` — builds: name varchar(200) default '', dockerfile_path varchar(500) default './Dockerfile', registry_id int nullable FK, image_repo varchar(500) default '', build_context varchar(500) default '.'
- [ ] **T003** 创建 `000031_add_deployment_config_fields.up/down.sql` — deployments: deploy_type varchar(10) default 'direct', workload_type varchar(20) default 'deployment', port int, cpu_request/mem_request/cpu_limit/mem_limit varchar(20), health_check_path varchar(200), helm_repo_id int nullable, helm_chart_path varchar(255), helm_release_name varchar(100), helm_chart_branch varchar(100)
- [ ] **T004** 创建 `000032_create_cluster_namespaces.up/down.sql` — cluster_namespaces 表 (id, cluster_id FK, namespace varchar(100), is_default bool, created_at, unique(cluster_id,namespace))

## Phase 2: Model 层 (4 tasks)

- [ ] **T005** 修改 `model/service.go` — 新增 ServiceType, Language, LanguageVersion 字段
- [ ] **T006** 修改 `model/build.go` — 新增 Name, DockerfilePath, RegistryID(*uint), ImageRepo, BuildContext 字段 + Registry *Registry 关联
- [ ] **T007** 修改 `model/deployment.go` — 新增 DeployType, WorkloadType, Port(*int), 资源配置×4, HealthCheckPath, HelmRepoID(*uint), HelmChartPath, HelmReleaseName, HelmChartBranch + HelmRepo 关联
- [ ] **T008** 新建 `model/cluster_namespace.go` — ClusterNamespace 模型

## Phase 3: 系统配置 (2 tasks)

- [ ] **T009** 修改 `config/config.go` — 新增 EnvValuesMap map[string]string 字段
- [ ] **T010** 实现 parseEnvValuesMap(envStr) — 解析 "qanet:qa,testnet:testnet,mainnet:mainnet" 格式

## Phase 4: Repository 层 (2 tasks)

- [ ] **T011** 新建 `repository/cluster_namespace_repo.go` — 接口：Create, Delete, ListByCluster, FindByClusterAndNamespace
- [ ] **T012** 新建 `repository/cluster_namespace_repo_impl.go` — GORM 实现

## Phase 5: Service 层 — 瘦身 (4 tasks)

- [ ] **T013** 修改 `handler/service.go` — createServiceRequest 瘦身：只保留 name, display_name, description, service_type, language, language_version, git_repo_id, git_branch
- [ ] **T014** 修改 `handler/service.go` — Create handler 只设置元信息字段
- [ ] **T015** 修改 `handler/service.go` — Update handler 同步瘦身
- [ ] **T016** 修改前端 service-create-dialog — 移除构建/部署相关字段，新增 service_type/language/language_version

## Phase 6: Build 扩展 (4 tasks)

- [ ] **T017** 修改 `handler/build.go` — triggerBuildRequest 新增 name, dockerfile_path, registry_id, image_repo, build_context
- [ ] **T018** 修改 `service/build/build.go` — CreateBuild 接受新参数并写入 model
- [ ] **T019** 修改 `service/build/executor.go` — 从 build 读取 RegistryID/DockerfilePath/ImageRepo（优先 build 自身，fallback service 兼容）
- [ ] **T020** 修改前端 trigger-build-dialog — 新增 dockerfile_path, registry 选择, image_repo, build_context 输入

## Phase 7: Deployment 扩展 (4 tasks)

- [ ] **T021** 修改 `handler/deploy.go` — createDeploymentRequest 新增 deploy_type, workload_type, port, resources, health_check_path, helm_* 字段
- [ ] **T022** 修改 `service/deploy/deploy.go` — CreateDeployment 接受新参数并写入 model
- [ ] **T023** 修改前端 deploy-dialog — 根据 deploy_type 条件渲染 Direct 配置区（port/replicas/resources/workload_type）和 Helm 配置区（chart_path/release_name/chart_branch）
- [ ] **T024** 修改前端 deploy-dialog — namespace 从预配置列表选择（支持手动输入 fallback）

## Phase 8: Namespace 管理 (4 tasks)

- [ ] **T025** 新建 `service/cluster/namespace.go` — ClusterNamespaceService：CRUD + SyncFromCluster（client-go 加载）
- [ ] **T026** 新建 `handler/namespace.go` — GET/POST/DELETE /clusters/:id/namespaces + POST /clusters/:id/namespaces/sync
- [ ] **T027** 修改 `cmd/server/main.go` — 注册 namespace 路由
- [ ] **T028** 新建前端 — 集群设置页增加 namespace 管理面板（列表 + 添加 + 删除 + 同步按钮）

## Phase 9: Executor 改造 (5 tasks)

- [ ] **T029** 修改 `service/build/executor.go` — buildKanikoJob 从 build.DockerfilePath/build.RegistryID/build.ImageRepo 读取（fallback service）
- [ ] **T030** 修改 `service/deploy/executor.go` — resolveImage 使用 deployment 字段；buildPodTemplate 使用 deployment.Port
- [ ] **T031** 修改 `service/deploy/executor.go` — Execute 中 WorkloadType 从 deployment 读取
- [ ] **T032** 修改 `service/deploy/helm_executor.go` — 从 deployment 读取 HelmRepoID/HelmChartPath/HelmReleaseName/HelmChartBranch
- [ ] **T033** 修改 `service/deploy/helm_job.go` — buildHelmUpgradeCmd 接受 deployment 的 Helm 字段 + ENV_VALUES_MAP 自动拼接

## Phase 10: ENV_VALUES_MAP 集成 (2 tasks)

- [ ] **T034** 修改 HelmExecutor 构造函数 — 注入 envValuesMap
- [ ] **T035** 修改 helm_job.go — buildHelmUpgradeCmd 根据 Cluster.Env 查 envValuesMap 自动拼接 -f services/{name}/app-{suffix}.yaml

## Phase 11: Service 旧字段 nullable (1 task)

- [ ] **T036** 创建 `000033_nullable_service_legacy_fields.up/down.sql` — dockerfile_path/registry_id/image_repo/replicas/port/deploy_type/workload_type/所有 helm_* 字段 DROP NOT NULL

## Phase 12: Test 修复 (4 tasks)

- [ ] **T037** 修改 `service/build/build_test.go` — CreateBuild 签名更新
- [ ] **T038** 修改 `handler/build_test.go` — 适配新 request 字段
- [ ] **T039** 修改 `service/deploy/deploy_test.go` + `handler/deploy_test.go` — 适配新字段
- [ ] **T040** 修改 `service/svc/svc_test.go` — Service 模型字段变更

## Phase 13: Frontend 完善 (4 tasks)

- [ ] **T041** 修改 `types/index.ts` — Service/Build/Deployment 类型更新，新增 ClusterNamespace 类型
- [ ] **T042** 新建 `hooks/use-namespaces.ts` — useClusterNamespaces, useSyncNamespaces
- [ ] **T043** 修改服务列表页 — 显示 service_type + language 标签
- [ ] **T044** 修改 `.env.example` — 新增 ENV_VALUES_MAP 配置项

## Phase 14: Verification (1 task)

- [ ] **T045** 运行 go test ./... + next build 全量通过

---

**总计: 45 tasks, 14 phases**

**关键风险管理**：
- Phase 9（Executor 改造）是最敏感的环节——需要 fallback 逻辑保证旧数据兼容
- Phase 11（旧字段 nullable）必须在 Phase 9 之后，确保代码已不依赖旧字段
- 前端构建/部署对话框扩展（Phase 6-7）会显著增大表单，需要按 deploy_type 条件渲染减少视觉复杂度
