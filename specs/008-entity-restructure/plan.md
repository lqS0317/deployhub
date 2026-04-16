# 008 - 核心实体职责重构实现计划

## Layer 1: DB Migrations (4 migrations)

- **新建** `000029_add_service_meta_fields.up/down.sql` — Service: service_type, language, language_version
- **新建** `000030_add_build_config_fields.up/down.sql` — Build: name, dockerfile_path, registry_id FK, image_repo, build_context
- **新建** `000031_add_deployment_config_fields.up/down.sql` — Deployment: deploy_type, workload_type, port, cpu/mem ×4, health_check_path, helm_repo_id, helm_chart_path, helm_release_name, helm_chart_branch
- **新建** `000032_create_cluster_namespaces.up/down.sql` — cluster_namespaces 表

## Layer 2: Model 层

- **修改** `model/service.go` — 新增 ServiceType, Language, LanguageVersion
- **修改** `model/build.go` — 新增 Name, DockerfilePath, RegistryID, ImageRepo, BuildContext + Registry 关联
- **修改** `model/deployment.go` — 新增 DeployType, WorkloadType, Port, 资源配置 ×4, HealthCheckPath, HelmRepoID, HelmChartPath, HelmReleaseName, HelmChartBranch + HelmRepo 关联
- **新建** `model/cluster_namespace.go` — ClusterNamespace 模型

## Layer 3: 系统配置

- **修改** `config/config.go` — 新增 EnvValuesMap 字段 + 解析函数

## Layer 4: Repository 层

- **新建** `repository/cluster_namespace_repo.go` + `_impl.go` — CRUD + ListByCluster

## Layer 5: Service 层

- **修改** `handler/service.go` — createServiceRequest 瘦身
- **修改** `service/build/build.go` — CreateBuild 接受构建配置参数
- **修改** `service/deploy/deploy.go` — CreateDeployment 接受部署配置参数

## Layer 6: Namespace 服务

- **新建** `service/cluster/namespace.go` — CRUD + client-go 动态加载

## Layer 7-10: Handler 层

- **修改** `handler/service.go` — 瘦身
- **修改** `handler/build.go` — 扩展 triggerBuildRequest
- **修改** `handler/deploy.go` — 扩展 createDeploymentRequest
- **新建** `handler/namespace.go` — Namespace CRUD + sync

## Layer 11-14: Executor 改造

- **修改** `service/build/executor.go` — 从 build 读取 dockerfile/registry/image_repo
- **修改** `service/deploy/executor.go` — 从 deployment 读取 port/workload_type
- **修改** `service/deploy/helm_executor.go` — 从 deployment 读取 helm_* + ENV_VALUES_MAP
- **修改** `service/deploy/helm_job.go` — 使用 deployment 字段

## Layer 15: Service 旧字段 nullable

- **新建** `000033_nullable_service_legacy_fields.up/down.sql`

## Layer 16: Frontend

- 服务表单瘦身 + 新字段
- 构建对话框扩展
- 部署对话框扩展
- Namespace 选择器

## Layer 17-18: 注册 + 验证

- main.go 注册
- go test ./... + next build
