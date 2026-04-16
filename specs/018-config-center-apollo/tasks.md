# Feature 018: 配置中心改造 — 任务列表

## Phase 1: DB Migrations

- [ ] **T1.1** 创建 `000049_create_config_namespaces.up.sql` — config_namespaces 表（id, service_id FK CASCADE, name, format, config_type, description, draft_content, created_at, updated_at），唯一索引 (service_id, name)
- [ ] **T1.2** 创建 `000049_create_config_namespaces.down.sql` — DROP TABLE
- [ ] **T1.3** 创建 `000050_create_config_items.up.sql` — config_items 表（id, namespace_id FK CASCADE, cluster_id, key, value, comment, is_deleted, created_at, updated_at），唯一索引 (namespace_id, cluster_id, key) WHERE NOT is_deleted
- [ ] **T1.4** 创建 `000050_create_config_items.down.sql`
- [ ] **T1.5** 创建 `000051_create_config_releases.up.sql` — config_releases 表（id, namespace_id FK CASCADE, cluster_id, version, snapshot jsonb, status, comment, created_by_id, created_at），索引 (namespace_id, cluster_id, version DESC)
- [ ] **T1.6** 创建 `000051_create_config_releases.down.sql`
- [ ] **T1.7** 创建 `000052_create_config_permissions.up.sql` — config_permissions 表（id, namespace_id FK CASCADE, cluster_id, user_id, role, created_at），唯一索引 (namespace_id, cluster_id, user_id)
- [ ] **T1.8** 创建 `000052_create_config_permissions.down.sql`

## Phase 2: Backend Models

- [ ] **T2.1** 创建 `model/config_namespace.go` — ConfigNamespace 结构体，format/config_type 常量，Service 关联
- [ ] **T2.2** 创建 `model/config_item.go` — ConfigItem 结构体，Namespace + Cluster 关联
- [ ] **T2.3** 创建 `model/config_release.go` — ConfigRelease 结构体，状态常量（published/rolled_back），CreatedBy 关联
- [ ] **T2.4** 创建 `model/config_permission.go` — ConfigPermission 结构体，角色常量（viewer/editor/publisher），角色层级方法

## Phase 3: Repository Layer

- [ ] **T3.1** 创建 `repository/config_namespace_repo.go` — 接口定义：List, FindByID, FindByServiceAndName, Create, Update, Delete
- [ ] **T3.2** 创建 `repository/config_namespace_repo_impl.go` — GORM 实现，List 按 service_id 过滤并 Preload Service
- [ ] **T3.3** 创建 `repository/config_item_repo.go` — 接口定义：List, FindByID, FindByKey, Create, Update, SoftDelete, PurgeDeleted, ListAll(含 deleted)
- [ ] **T3.4** 创建 `repository/config_item_repo_impl.go` — GORM 实现，List 排除 is_deleted，SoftDelete 设 is_deleted=true，PurgeDeleted 硬删除 is_deleted 记录
- [ ] **T3.5** 创建 `repository/config_release_repo.go` — 接口定义：List, FindByID, FindLatestPublished, Create, GetNextVersion, UpdateStatus
- [ ] **T3.6** 创建 `repository/config_release_repo_impl.go` — GORM 实现，GetNextVersion 用 COALESCE(MAX(version),0)+1，List 按 version DESC 排序
- [ ] **T3.7** 创建 `repository/config_permission_repo.go` — 接口定义：List, FindByUserAndCluster, Upsert, Delete
- [ ] **T3.8** 创建 `repository/config_permission_repo_impl.go` — GORM 实现，FindByUserAndCluster 查 cluster_id=cid OR cluster_id=0

## Phase 4: Service Layer

- [ ] **T4.1** 创建 `service/configcenter/config_service.go` — ConfigService 结构体，注入 4 个 repo + cryptoSvc
- [ ] **T4.2** 实现 ConfigService.CreateNamespace — 校验 name 唯一 + format/config_type 枚举
- [ ] **T4.3** 实现 ConfigService.UpdateNamespace — 只允许改 description
- [ ] **T4.4** 实现 ConfigService.DeleteNamespace — 直接删除（FK CASCADE 处理子表）
- [ ] **T4.5** 实现 ConfigService.ListItems — 返回配置项列表（secret 值脱敏标记）
- [ ] **T4.6** 实现 ConfigService.CreateItem — 校验 properties 格式 + key 唯一 + secret 加密
- [ ] **T4.7** 实现 ConfigService.UpdateItem — 校验存在 + secret 加密
- [ ] **T4.8** 实现 ConfigService.DeleteItem — 软删除
- [ ] **T4.9** 实现 ConfigService.SaveDraft — yaml/json 格式草稿保存
- [ ] **T4.10** 实现 ConfigService.GetDraft — 读取草稿（分 cluster 存储）
- [ ] **T4.11** 实现 ConfigService.Publish — 收集配置 → 生成快照 JSON → 创建 release → 清理 deleted items
- [ ] **T4.12** 实现 ConfigService.Rollback — 恢复历史快照内容 → 创建新 release(rolled_back)
- [ ] **T4.13** 实现 ConfigService.GetPublishedConfig — 查询服务所有配置集最新发布版本（供部署用）
- [ ] **T4.14** 创建 `service/configcenter/permission_service.go` — ConfigPermissionService
- [ ] **T4.15** 实现 CheckPermission — Admin 通过 → Owner 通过 → 查 ConfigPermission 表 → 角色层级比较
- [ ] **T4.16** 实现 GrantPermission / RevokePermission — Upsert / Delete

## Phase 5: Handler Layer

- [ ] **T5.1** 创建 `handler/config_center.go` — ConfigCenterHandler 结构体，注入 ConfigService + ConfigPermissionService
- [ ] **T5.2** 实现 ListNamespaces — GET /services/:id/config-namespaces
- [ ] **T5.3** 实现 CreateNamespace — POST /services/:id/config-namespaces（需 owner/admin）
- [ ] **T5.4** 实现 UpdateNamespace — PUT /config-namespaces/:id（需 editor+）
- [ ] **T5.5** 实现 DeleteNamespace — DELETE /config-namespaces/:id（需 publisher+）
- [ ] **T5.6** 实现 ListItems — GET /config-namespaces/:id/clusters/:cid/items（viewer+，secret 脱敏）
- [ ] **T5.7** 实现 CreateItem — POST /config-namespaces/:id/clusters/:cid/items（editor+）
- [ ] **T5.8** 实现 UpdateItem — PUT /config-namespaces/:id/clusters/:cid/items/:item_id（editor+）
- [ ] **T5.9** 实现 DeleteItem — DELETE /config-namespaces/:id/clusters/:cid/items/:item_id（editor+）
- [ ] **T5.10** 实现 GetDraft — GET /config-namespaces/:id/clusters/:cid/draft（viewer+）
- [ ] **T5.11** 实现 SaveDraft — PUT /config-namespaces/:id/clusters/:cid/draft（editor+）
- [ ] **T5.12** 实现 Publish — POST /config-namespaces/:id/clusters/:cid/release（publisher+）
- [ ] **T5.13** 实现 Rollback — POST /config-namespaces/:id/clusters/:cid/rollback（publisher+）
- [ ] **T5.14** 实现 ListReleases — GET /config-namespaces/:id/clusters/:cid/releases
- [ ] **T5.15** 实现 ListPermissions — GET /config-namespaces/:id/permissions
- [ ] **T5.16** 实现 GrantPermission — POST /config-namespaces/:id/permissions
- [ ] **T5.17** 实现 RevokePermission — DELETE /config-namespaces/:id/permissions/:perm_id
- [ ] **T5.18** 注册路由到 main.go — 创建 repo/service/handler 实例，注册所有路由
- [ ] **T5.19** 编译 + 运行全量测试验证后端

## Phase 6: Frontend

- [ ] **T6.1** 创建 `hooks/use-config-center.ts` — TanStack Query hooks: useConfigNamespaces, useConfigItems, useConfigDraft, useConfigReleases, useConfigPermissions + mutations
- [ ] **T6.2** 创建 `app/(dashboard)/configs/page.tsx` — 配置中心入口：服务选择列表
- [ ] **T6.3** 创建 `app/(dashboard)/configs/[serviceId]/page.tsx` — 服务配置主页面：左右布局框架
- [ ] **T6.4** 创建 `components/config/env-sidebar.tsx` — 左侧面板：环境 tabs（Cluster 列表）+ 配置集列表（名称/格式/类型 badge）+ 新建配置集按钮
- [ ] **T6.5** 创建 `components/config/create-namespace-dialog.tsx` — 创建配置集弹窗：名称、格式(properties/yaml/json)、类型(configmap/secret)、描述
- [ ] **T6.6** 创建 `components/config/kv-table-editor.tsx` — properties 表格编辑器：key/value/comment 列，内联编辑、新增行、删除行、搜索过滤、secret 值脱敏
- [ ] **T6.7** 创建 `components/config/code-editor.tsx` — YAML/JSON 代码编辑器：textarea + 行号 + 保存按钮
- [ ] **T6.8** 创建 `components/config/config-content.tsx` — 右侧内容区：三个 Tab（配置项、更改历史、发布历史），根据 format 切换 KV 表格或代码编辑器
- [ ] **T6.9** 创建 `components/config/publish-dialog.tsx` — 发布弹窗：变更 diff 列表（新增/修改/删除）+ 发布说明输入 + 确认
- [ ] **T6.10** 创建 `components/config/rollback-dialog.tsx` — 回滚弹窗：版本列表 + 快照预览 + 确认回滚
- [ ] **T6.11** 创建 `components/config/release-history.tsx` — 发布历史 Tab：版本号/状态/发布人/时间/说明 + 查看快照 + 回滚按钮
- [ ] **T6.12** 创建 `components/config/permission-dialog.tsx` — 权限管理弹窗：当前权限列表 + 新增授权（用户/环境/角色）
- [ ] **T6.13** 更新侧边栏配置中心入口链接

## Phase 7: Deploy Integration

- [ ] **T7.1** 创建 `deploy/config_deploy.go` — GenerateConfigResources(serviceID, clusterID, cryptoSvc, configService) 函数
- [ ] **T7.2** 实现 snapshot 解析 → 生成 K8s ConfigMap/Secret YAML 字符串（secret 值解密）
- [ ] **T7.3** 改造 ConfigExecutor.generateYAML — 调用 GenerateConfigResources 追加 ConfigMap/Secret
- [ ] **T7.4** 改造 YamlExecutor — Execute 前追加 ConfigMap/Secret 到 raw_yaml
- [ ] **T7.5** 改造 HelmExecutor — 发布前创建 ConfigMap/Secret（直接 apply，不走 helm）
- [ ] **T7.6** 编译 + 全量测试验证

## Phase 8: 旧模型清理（后续单独执行）

- [ ] **T8.1** 编写数据迁移脚本：ConfigTemplate → ConfigNamespace
- [ ] **T8.2** 编写数据迁移脚本：ConfigEnvValue → ConfigItem
- [ ] **T8.3** 编写数据迁移脚本：ConfigVersion → ConfigRelease
- [ ] **T8.4** 删除旧 service/config/ 目录下的代码
- [ ] **T8.5** 删除旧 handler/config.go
- [ ] **T8.6** 删除旧 repository/config_*.go（旧的 4 个文件）
- [ ] **T8.7** 删除旧 model/config_*.go（旧的 4 个文件）
- [ ] **T8.8** 创建 migration DROP 旧表（config_templates, config_env_values, config_versions, config_deployments）
- [ ] **T8.9** 清理 main.go 中旧 config 相关的依赖注入和路由
- [ ] **T8.10** 编译 + 全量测试验证

---

**总计: 66 个任务**
- Phase 1: 8 tasks (DB)
- Phase 2: 4 tasks (Models)
- Phase 3: 8 tasks (Repository)
- Phase 4: 16 tasks (Service)
- Phase 5: 19 tasks (Handler)
- Phase 6: 13 tasks (Frontend)
- Phase 7: 6 tasks (Deploy Integration)
- Phase 8: 10 tasks (Cleanup)
