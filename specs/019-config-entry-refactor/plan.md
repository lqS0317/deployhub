# Feature 019: 配置中心多条目模式 + 部署零覆盖 — 实现计划

## Phase 1: DB Migrations

### 1.1 重建配置中心表
- **000054_create_config_entries.up/down.sql**
  - DROP config_items, config_releases, config_permissions, service_config_envs（清空旧数据）
  - CREATE config_entries (id, service_id FK CASCADE, cluster_id, name, config_type, format, draft_content, timestamps)
  - UNIQUE(service_id, cluster_id, name)

- **000055_recreate_config_items.up/down.sql**
  - CREATE config_items (id, config_entry_id FK CASCADE, key, value, comment, is_deleted, timestamps)
  - UNIQUE INDEX (config_entry_id, key) WHERE NOT is_deleted

- **000056_recreate_config_releases.up/down.sql**
  - CREATE config_releases (id, config_entry_id FK CASCADE, version, snapshot jsonb, status, comment, created_by_id, created_at)
  - INDEX (config_entry_id, version DESC)

- **000057_recreate_config_permissions.up/down.sql**
  - CREATE config_permissions (id, service_id FK CASCADE, cluster_id, user_id, role, created_at)
  - UNIQUE(service_id, cluster_id, user_id)

- **000058_create_service_config_refs.up/down.sql**
  - CREATE service_config_refs (id, service_id FK CASCADE, config_entry_name, mount_path, created_at)
  - UNIQUE(service_id, config_entry_name)

### 1.2 Deployment 字段清理
- **000059_clean_deployment_fields.up/down.sql**
  - DROP COLUMN: direct_mode, raw_yaml, env_vars, volumes, volume_claim_templates, secret_refs, config_map_refs, service_account_name, service_spec, liveness_probe, readiness_probe, command, args, port, cpu_request, mem_request, cpu_limit, mem_limit, replicas（改为从 Service 读取）

### 1.3 Service 字段清理
- **000060_clean_service_config_fields.up/down.sql**
  - DROP COLUMN: default_env_vars, default_secret_refs, default_config_map_refs（移到配置中心）
  - 保留: default_port, default_replicas, default_cpu/mem, default_command, default_args, default_liveness/readiness_probe, default_volumes, default_volume_claim_templates, default_service_account_name, default_workload_type

## Phase 2: Backend Models

### 2.1 新模型
- model/config_entry.go — ConfigEntry 结构体 + config_type/format 常量
- model/service_config_ref.go — ServiceConfigRef 结构体

### 2.2 更新模型
- model/config_item.go — namespace_id/service_id → config_entry_id
- model/config_release.go — service_id → config_entry_id
- model/deployment.go — 删除 ~20 个字段（保留核心跟踪字段）
- model/service.go — 删除 DefaultEnvVars, DefaultSecretRefs, DefaultConfigMapRefs

### 2.3 删除模型
- 删除 model/service_config_env.go

## Phase 3: Repository Layer

### 3.1 新 Repository
- repository/config_entry_repo.go + impl — List(serviceID, clusterID), FindByID, FindByName(svcID, cID, name), Create, Update, Delete
- repository/service_config_ref_repo.go + impl — List(serviceID), FindByID, Create, Update, Delete

### 3.2 更新 Repository
- config_item_repo.go + impl — service_id+cluster_id 参数改为 config_entry_id
- config_release_repo.go + impl — service_id+cluster_id 改为 config_entry_id

### 3.3 删除 Repository
- 删除 service_config_env_repo.go + impl

## Phase 4: Service Layer

### 4.1 重写 ConfigService
- configcenter/config_service.go — 方法改为 ConfigEntry 维度：
  - CreateEntry, UpdateEntry, DeleteEntry, ListEntries, GetEntry
  - ListItems(entryID), CreateItem, UpdateItem, DeleteItem（加密逻辑不变）
  - GetDraft(entryID), SaveDraft(entryID)
  - Publish(entryID), Rollback(entryID, version), ListReleases(entryID)
  - GetPublishedConfigForEntry(entryID) — 单条目最新发布

### 4.2 新增 ServiceConfigRefService
- 或合并到 ConfigService 中：
  - ListRefs(serviceID), CreateRef, UpdateRef, DeleteRef

### 4.3 更新 ConfigPermissionService
- 保持 service_id + cluster_id 粒度不变，但内部不再查 namespace/configEnv

## Phase 5: Handler Layer

### 5.1 重写 config_center.go
- 路由从 /services/:id/configs/:cid/... 改为:
  - /services/:id/config-entries (LIST, CREATE)
  - /config-entries/:id (GET, PUT, DELETE)
  - /config-entries/:id/items[/:item_id] (CRUD)
  - /config-entries/:id/draft (GET, PUT)
  - /config-entries/:id/release (POST)
  - /config-entries/:id/rollback (POST)
  - /config-entries/:id/releases (GET)
  - /services/:id/config-refs (LIST, CREATE)
  - /services/:id/config-refs/:ref_id (PUT, DELETE)
  - /services/:id/configs/:cid/permissions (保持)

### 5.2 简化 deploy handler
- createDeploymentRequest 删除所有配置覆盖字段
- DeployConfig 删除配置相关字段
- Create handler 不再传配置参数

### 5.3 更新 main.go
- 替换 configEnvRepo → configEntryRepo
- 新增 configRefRepo
- 更新 ConfigService 构造函数
- 更新路由注册

## Phase 6: Deploy Integration

### 6.1 重写 ConfigDeployHelper
- deploy/config_deploy.go — GenerateConfigYAML(serviceID, clusterID, serviceName) :
  1. 查 ServiceConfigRef 获取引用列表
  2. 按 (name + cluster_id) 查 ConfigEntry
  3. 查最新 published ConfigRelease
  4. 按 config_type + mount_path 生成:
     - env → ConfigMap(envFrom)
     - configmap → ConfigMap + volumeMount
     - secret → Secret + volumeMount

### 6.2 重写 ConfigExecutor
- config_executor.go — generateYAML 改为:
  1. 从 Service 读取运行时参数（replicas, port, cpu/mem, probe, command, args, volumes, SA）
  2. 调用 ConfigDeployHelper 获取 ConfigMap/Secret YAML
  3. 将 ConfigRef 信息注入到容器 spec（envFrom, volumeMounts）
  4. 生成 Deployment/StatefulSet + Service YAML

### 6.3 清理 Deployment 相关逻辑
- deploy/deploy.go — DeployConfig 删除配置字段, CreateDeployment 简化
- handler/deploy.go — createDeploymentRequest 简化

## Phase 7: Frontend — 配置中心改造

### 7.1 Hooks
- use-config-center.ts — 改为 config-entries 路由:
  - useConfigEntries(serviceId, clusterId), useCreateEntry, useUpdateEntry, useDeleteEntry
  - useConfigItems(entryId), useCreateItem, useUpdateItem, useDeleteItem
  - useConfigDraft(entryId), useSaveDraft
  - usePublish(entryId), useRollback(entryId), useConfigReleases(entryId)
  - useServiceConfigRefs(serviceId), useCreateRef, useUpdateRef, useDeleteRef

### 7.2 配置中心页面
- configs/[serviceId]/page.tsx — 左侧环境列表 + 右侧条目列表 → 点击条目 → 编辑面板

### 7.3 配置条目组件
- config/entry-list.tsx — 条目表格（名称/类型/最新版本/操作）
- config/create-entry-dialog.tsx — 新建条目弹窗（名称/类型/格式）
- config/entry-detail.tsx — 条目编辑（复用 KV 表格 + 代码编辑器 + 发布/回滚）

## Phase 8: Frontend — 服务编辑 + 部署对话框

### 8.1 服务编辑
- service-create-dialog.tsx — 新增"配置引用"分区:
  - 表格: 配置名称 | 挂载路径 | 删除
  - 添加引用按钮

### 8.2 部署对话框简化
- deploy-dialog.tsx — 删除 DirectConfigForm 和 YAML 编辑器
- 配置模式简化为: 服务 → 集群+命名空间 → 镜像 → 确认

## Phase 9: Cleanup

### 9.1 删除文件
- model/service_config_env.go
- repository/service_config_env_repo.go + impl
- components/config/init-config-form.tsx
- components/deploy/direct-config-form.tsx
- components/service/service-runtime-config.tsx 中的 env/secret/configmap 部分

### 9.2 类型清理
- types/index.ts — 删除 ServiceConfigEnv, 新增 ConfigEntry + ServiceConfigRef
- Deployment type 删除已移除字段

## Constitution Check

- ✅ 安全第一: Secret 值 AES-256-GCM 加密
- ✅ 分层架构: handler → service → repository
- ✅ 多集群隔离: 配置按 Cluster 环境隔离
- ✅ 简洁优先: 部署零覆盖减少配置入口
- ✅ 配置中心: 多条目模式，显式引用，按条目粒度发布
