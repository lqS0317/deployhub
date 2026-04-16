# 008 - 核心实体职责重构

## 1. Overview

对 Service / Build / Deployment 三个核心实体进行职责重新划分：Service 瘦身为纯服务元信息，Build 扩展为构建配置中心，Deployment 扩展为部署配置中心。新增 Namespace 管理和 Helm Env Values 自动映射。

## 2. Field Migration Plan

### 2.1 Service — 瘦身后保留字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | PK |
| name | varchar(100) | 服务名，唯一 |
| display_name | varchar(200) | 显示名 |
| description | text | 描述 |
| **service_type** | varchar(50) | **新增** 业务类型：frontend/backend/worker/cronjob 等 |
| **language** | varchar(50) | **新增** 开发语言：go/node/java/python 等 |
| **language_version** | varchar(50) | **新增** 运行时版本：go1.22/node20 等 |
| git_repo_id | uint FK | **保留** 代码仓库 |
| git_branch | varchar(200) | **保留** 默认分支（UI 预填用） |
| owner_id | uint FK | **保留** 负责人 |
| created_at | timestamp | |
| updated_at | timestamp | |

**移除的字段**（渐进式 nullable → 后续删除）：
- dockerfile_path → Build
- registry_id → Build
- image_repo → Build
- cluster_id, namespace → 已在 007 改为 nullable
- replicas, cpu/mem request/limit, port, health_check_path → Deployment
- env_vars, volumes → Deployment
- deploy_type, workload_type → Deployment
- helm_repo_id, helm_chart_path, helm_values_path, helm_release_name, helm_chart_branch, helm_env_file_path → Deployment

### 2.2 Build — 扩展后字段

| 字段 | 状态 | 类型 | 说明 |
|------|------|------|------|
| id | 已有 | uint | PK |
| service_id | 已有 | uint FK | |
| trigger_user_id | 已有 | uint FK | |
| git_branch | 已有 | varchar(200) | 构建分支 |
| git_commit | 已有 | varchar(40) | |
| image_tag | 已有 | varchar(200) | 产物标签 |
| status | 已有 | varchar(20) | |
| build_cluster_id | 已有 | uint FK | Kaniko 集群 |
| kaniko_job_name | 已有 | varchar(200) | |
| log | 已有 | text | |
| started_at / finished_at | 已有 | timestamp | |
| **name** | **新增** | varchar(200) | 构建标识名（如 "v1.2.3 构建"） |
| **dockerfile_path** | **新增** | varchar(500) | 从 Service 迁移，默认 ./Dockerfile |
| **registry_id** | **新增** | uint FK nullable | 从 Service 迁移，目标镜像仓库 |
| **image_repo** | **新增** | varchar(500) | 从 Service 迁移，镜像路径 |
| **build_context** | **新增** | varchar(500) | 构建上下文目录，默认 '.' |

### 2.3 Deployment — 扩展后字段

| 字段 | 状态 | 说明 |
|------|------|------|
| 所有已有字段 | 保留 | id, service_id, build_id, trigger_user_id, cluster_id, namespace, image_tag, replicas, status, previous_image_tag, is_rollback, rollback_from_id, helm_revision, image_source, external_image, preview_yaml, preview_summary, started_at, finished_at, created_at |
| **deploy_type** | **新增** | varchar(10), default 'direct' (direct/helm) |
| **workload_type** | **新增** | varchar(20), default 'deployment' |
| **port** | **新增** | int, nullable |
| **cpu_request** | **新增** | varchar(20) |
| **mem_request** | **新增** | varchar(20) |
| **cpu_limit** | **新增** | varchar(20) |
| **mem_limit** | **新增** | varchar(20) |
| **health_check_path** | **新增** | varchar(200) |
| **helm_repo_id** | **新增** | uint FK nullable |
| **helm_chart_path** | **新增** | varchar(255) |
| **helm_release_name** | **新增** | varchar(100) |
| **helm_chart_branch** | **新增** | varchar(100) |

### 2.4 新增模型 — ClusterNamespace

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | PK |
| cluster_id | uint FK | 所属集群 |
| namespace | varchar(100) | 命名空间名 |
| is_default | bool | 是否默认 |
| created_at | timestamp | |

唯一索引: (cluster_id, namespace)

### 2.5 新增系统配置 — ENV_VALUES_MAP

环境变量 `ENV_VALUES_MAP`：
```
ENV_VALUES_MAP=qanet:qa,testnet:testnet,mainnet:mainnet
```

Helm 部署时根据 Cluster.Env 查找映射，自动拼接：
```
-f services/{serviceName}/app.yaml
-f services/{serviceName}/app-{envSuffix}.yaml
```

## 3. Executor 改造

### 3.1 BuildExecutor

当前从 `service` 读取 `RegistryID`、`DockerfilePath`、`ImageRepo`。改为从 `build` 自身读取：

```go
// 旧: e.registryRepo.FindByID(svc.RegistryID)
// 新: e.registryRepo.FindByID(build.RegistryID)

// 旧: svc.DockerfilePath → Kaniko --dockerfile
// 新: build.DockerfilePath → Kaniko --dockerfile
```

### 3.2 DirectExecutor

当前从 `service` 读取 `ImageRepo`、`Port`、`WorkloadType`。改为从 `deployment` 读取：

```go
// 旧: service.ImageRepo, service.Port, service.WorkloadType
// 新: deployment.Port, deployment.WorkloadType
// image 仍通过 resolveImage(deployment, service) 或纯 deployment
```

### 3.3 HelmExecutor

当前从 `service` 读取 Helm 配置。改为从 `deployment` 读取：
- `deployment.HelmRepoID` → Chart Git 仓库
- `deployment.HelmChartPath` → Chart 路径
- `deployment.HelmReleaseName` → Release 名称
- `deployment.HelmChartBranch` → Chart 分支

Env values 自动拼接：从 `ENV_VALUES_MAP` + `Cluster.Env` 生成文件路径。

## 4. API Changes

### Service API

- `POST /services` — 只接受元信息字段（name, display_name, description, service_type, language, language_version, git_repo_id）
- `PUT /services/:id` — 同上
- `GET /services` — 返回瘦身后的字段

### Build API

- `POST /builds` — 新增 name, dockerfile_path, registry_id, image_repo, build_context 字段
- 前端记住上次构建参数，预填表单

### Deployment API

- `POST /deployments` — 新增 deploy_type, workload_type, port, cpu/mem, health_check_path, helm_* 字段
- 前端根据 deploy_type 条件渲染不同配置区域

### Namespace API

| Method | Path | 说明 |
|--------|------|------|
| GET | /clusters/:id/namespaces | 获取集群 namespace 列表（静态+动态） |
| POST | /clusters/:id/namespaces | 添加 namespace |
| DELETE | /clusters/:id/namespaces/:ns_id | 删除 namespace |
| POST | /clusters/:id/namespaces/sync | 动态加载集群 namespace（client-go） |

## 5. Frontend Changes

- **服务创建/编辑**：瘦身为元信息（名称、类型、语言、仓库、负责人）
- **构建触发对话框**：新增 dockerfile_path、registry 选择、image_repo、build_context
- **部署发起对话框**：新增 deploy_type 选择 + 对应配置区域（Direct/Helm），namespace 从预配置列表选择
- **服务列表**：显示 service_type + language 标签
- **集群详情/设置**：Namespace 管理面板

## 6. Migration Strategy

```
Phase A: 新增字段（Build 5个, Deployment 10个, Service 3个, ClusterNamespace 表）
Phase B: 代码切换（executor 从新字段读取，handler 接受新字段）
Phase C: Service 旧字段 nullable（已在 007 对 cluster_id/namespace 做了）
Phase D: 数据迁移脚本（可选：从现有 Service 填充 Build/Deployment 默认值）
Phase E: 清理 Service 旧字段（后续版本）
```

## 7. Success Criteria

- [ ] Service 只包含元信息字段
- [ ] Build 可独立配置 dockerfile/registry/image_repo
- [ ] Deployment 可独立配置 deploy_type/workload_type/port/resources/helm_*
- [ ] Namespace 可预配置 + 动态加载
- [ ] Helm 部署自动拼接 env values 文件
- [ ] 所有 executor 从新字段读取配置
- [ ] 全部测试通过

## 8. Non-Goals

- 构建配置模板
- 部署配置模板
- Service 版本历史
- 修改审批/权限/通知核心逻辑
- 修改 Kaniko/Helm Runner 底层执行
- Namespace 自动创建
