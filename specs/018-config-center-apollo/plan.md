# Feature 018: 配置中心改造 — 实现计划

## Phase 1: DB Migrations（清理旧表 + 创建新表）

### 1.1 创建新配置表
- **000049_create_config_namespaces.up/down.sql** — config_namespaces 表，唯一索引 (service_id, name)，ON DELETE CASCADE 关联 services
- **000050_create_config_items.up/down.sql** — config_items 表，唯一索引 (namespace_id, cluster_id, key) WHERE NOT is_deleted，ON DELETE CASCADE 关联 config_namespaces
- **000051_create_config_releases.up/down.sql** — config_releases 表，索引 (namespace_id, cluster_id, version DESC)，ON DELETE CASCADE 关联 config_namespaces
- **000052_create_config_permissions.up/down.sql** — config_permissions 表，唯一索引 (namespace_id, cluster_id, user_id)

> 暂不删除旧表，待数据迁移完成后再清理

## Phase 2: Backend Models（GORM 结构体）

### 2.1 新模型文件
- **model/config_namespace.go** — ConfigNamespace 结构体（含 Service 关联）
- **model/config_item.go** — ConfigItem 结构体（含 Namespace、Cluster 关联）
- **model/config_release.go** — ConfigRelease 结构体（含 CreatedBy 关联）+ 状态常量
- **model/config_permission.go** — ConfigPermission 结构体 + 角色常量

## Phase 3: Repository Layer

### 3.1 ConfigNamespace Repository
- 接口: List(serviceID), FindByID, FindByServiceAndName, Create, Update, Delete
- 实现: GORM 操作，Delete 级联（items + releases + permissions 由 FK CASCADE 处理）

### 3.2 ConfigItem Repository
- 接口: List(namespaceID, clusterID), FindByID, FindByKey(nsID, cID, key), Create, Update, SoftDelete, PurgeDeleted(nsID, cID)
- 实现: List 排除 is_deleted=true，SoftDelete 标记 is_deleted=true

### 3.3 ConfigRelease Repository
- 接口: List(namespaceID, clusterID), FindByID, FindLatestPublished(nsID, cID), Create, GetNextVersion(nsID, cID), UpdateStatus
- 实现: GetNextVersion 查 MAX(version)+1

### 3.4 ConfigPermission Repository
- 接口: List(namespaceID), FindByUserAndCluster(nsID, cID, userID), Upsert, Delete
- 实现: cluster_id=0 匹配所有环境

## Phase 4: Service Layer

### 4.1 ConfigService — 核心业务逻辑
- **Namespace CRUD**: 创建时校验 name 唯一，format 和 config_type 不可变更
- **Item CRUD**: 仅 properties 格式允许，key 唯一校验，secret 类型值 AES 加密
- **Draft**: yaml/json 格式保存草稿到 namespace.draft_content
- **Publish**: 收集当前配置生成快照 → 创建 ConfigRelease → 清理 soft-deleted items
- **Rollback**: 恢复历史快照 → 创建新 release（status=rolled_back）
- **GetPublishedConfig**: 查询服务所有配置集最新发布版本（供部署流程调用）

### 4.2 ConfigPermissionService — 权限逻辑
- **CheckPermission(nsID, cID, userID, requiredRole)**: 检查权限层级
  - Admin → 全通过
  - 服务 Owner → publisher（需查 namespace → service → owner_id）
  - ConfigPermission 表查询（cluster_id=0 为通配）
  - 角色层级: publisher > editor > viewer
- **GrantPermission / RevokePermission**: CRUD 操作

## Phase 5: Handler Layer（REST API）

### 5.1 ConfigNamespaceHandler
- GET /services/:id/config-namespaces — 列出服务下所有配置集
- POST /services/:id/config-namespaces — 创建配置集（需 service owner 或 admin）
- PUT /config-namespaces/:id — 更新描述（需 editor+）
- DELETE /config-namespaces/:id — 删除（需 publisher+）

### 5.2 ConfigItemHandler
- GET /config-namespaces/:id/clusters/:cid/items — 列出配置项（viewer+，secret 值按权限脱敏）
- POST /config-namespaces/:id/clusters/:cid/items — 新增配置项（editor+）
- PUT /config-namespaces/:id/clusters/:cid/items/:item_id — 修改（editor+）
- DELETE /config-namespaces/:id/clusters/:cid/items/:item_id — 软删除（editor+）

### 5.3 ConfigDraftHandler
- GET /config-namespaces/:id/clusters/:cid/draft — 获取草稿（viewer+）
- PUT /config-namespaces/:id/clusters/:cid/draft — 保存草稿（editor+）

### 5.4 ConfigReleaseHandler
- POST /config-namespaces/:id/clusters/:cid/release — 发布（publisher+）
- POST /config-namespaces/:id/clusters/:cid/rollback — 回滚到指定版本（publisher+）
- GET /config-namespaces/:id/clusters/:cid/releases — 发布历史列表

### 5.5 ConfigPermissionHandler
- GET /config-namespaces/:id/permissions — 列出权限
- POST /config-namespaces/:id/permissions — 授权（需 publisher+ 或 admin）
- DELETE /config-namespaces/:id/permissions/:perm_id — 移除权限

### 5.6 路由注册
- main.go 中注册所有新路由，注入依赖

## Phase 6: Frontend — 配置中心页面

### 6.1 页面结构
- **/configs/[serviceId]/page.tsx** — 服务配置中心主页面
- 左侧面板: 环境 tabs（从 clusters API 获取）+ 配置集列表
- 右侧内容区: 三个 Tab（配置项、更改历史、发布历史）

### 6.2 配置项编辑
- **properties 模式**: Key-Value 表格组件（内联编辑、新增行、删除行、搜索过滤）
- **yaml/json 模式**: textarea 代码编辑器（语法高亮后续可加 Monaco）

### 6.3 发布/回滚交互
- 发布弹窗: 显示与上次发布版本的 diff（新增/修改/删除项）+ 发布说明
- 回滚弹窗: 版本列表 + 选中版本快照预览 + 确认

### 6.4 权限管理
- 授权弹窗: 用户选择 + 环境选择 + 角色选择

### 6.5 入口改造
- 侧边栏"配置中心"链接改为进入服务列表，点击服务进入配置页面
- 或保持 /configs 入口，先选服务再进入配置管理

## Phase 7: Deploy Integration

### 7.1 ConfigDeployHelper
- 新建 deploy/config_deploy.go
- `GenerateConfigResources(serviceID, clusterID)` → 返回 []ConfigMapOrSecret
- 查询所有配置集最新 published release → 解析 snapshot → 生成 K8s ConfigMap/Secret YAML

### 7.2 Executor 改造
- DirectExecutor/ConfigExecutor: Execute 前调用 ConfigDeployHelper，将生成的 ConfigMap/Secret 追加到 YAML
- HelmExecutor: 在 Helm values 中注入配置（或生成 ConfigMap 后 --set 引用）

## Phase 8: 旧模型清理

### 8.1 数据迁移脚本
- 迁移 ConfigTemplate → ConfigNamespace
- 迁移 ConfigEnvValue → ConfigItem（按环境拆分）
- 迁移 ConfigVersion → ConfigRelease
- 删除旧 handler/service/repository/model 代码
- Migration: DROP 旧表

## Constitution Check

- ✅ 安全第一: Secret 值 AES-256-GCM 加密，权限校验
- ✅ 分层架构: handler → service → repository
- ✅ 多集群隔离: 配置按 Cluster 环境隔离
- ✅ 简洁优先: 4 个新模型替换 4 个旧模型，不增加复杂度
- ✅ 代码注释: 中文
