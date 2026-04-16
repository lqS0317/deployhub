# Feature 019: 配置中心多条目模式 + 部署零覆盖 — 任务列表

## Phase 1: DB Migrations

- [ ] **T1.1** 000054_create_config_entries.up.sql — DROP config_items, config_releases, config_permissions, service_config_envs; CREATE config_entries 表 (service_id FK CASCADE, cluster_id, name, config_type, format, draft_content)，UNIQUE(service_id, cluster_id, name)
- [ ] **T1.2** 000054_create_config_entries.down.sql
- [ ] **T1.3** 000055_recreate_config_items.up.sql — CREATE config_items (config_entry_id FK CASCADE, key, value, comment, is_deleted)，UNIQUE(config_entry_id, key) WHERE NOT is_deleted
- [ ] **T1.4** 000055_recreate_config_items.down.sql
- [ ] **T1.5** 000056_recreate_config_releases.up.sql — CREATE config_releases (config_entry_id FK CASCADE, version, snapshot jsonb, status, comment, created_by_id)，INDEX(config_entry_id, version DESC)
- [ ] **T1.6** 000056_recreate_config_releases.down.sql
- [ ] **T1.7** 000057_recreate_config_permissions.up.sql — CREATE config_permissions (service_id FK CASCADE, cluster_id, user_id, role)，UNIQUE(service_id, cluster_id, user_id)
- [ ] **T1.8** 000057_recreate_config_permissions.down.sql
- [ ] **T1.9** 000058_create_service_config_refs.up.sql — CREATE service_config_refs (service_id FK CASCADE, config_entry_name, mount_path)，UNIQUE(service_id, config_entry_name)
- [ ] **T1.10** 000058_create_service_config_refs.down.sql
- [ ] **T1.11** 000059_clean_deployment_fields.up.sql — ALTER TABLE deployments DROP COLUMN: direct_mode, raw_yaml, env_vars, volumes, volume_claim_templates, secret_refs, config_map_refs, service_account_name, service_spec, liveness_probe, readiness_probe, command, args, port, cpu_request, mem_request, cpu_limit, mem_limit, replicas
- [ ] **T1.12** 000059_clean_deployment_fields.down.sql — ADD COLUMN 恢复所有字段
- [ ] **T1.13** 000060_clean_service_config_fields.up.sql — ALTER TABLE services DROP COLUMN: default_env_vars, default_secret_refs, default_config_map_refs
- [ ] **T1.14** 000060_clean_service_config_fields.down.sql

## Phase 2: Backend Models

- [ ] **T2.1** 创建 model/config_entry.go — ConfigEntry 结构体，config_type 常量(env/configmap/secret)，format 常量复用
- [ ] **T2.2** 创建 model/service_config_ref.go — ServiceConfigRef 结构体
- [ ] **T2.3** 更新 model/config_item.go — service_id+cluster_id 改为 config_entry_id FK
- [ ] **T2.4** 更新 model/config_release.go — service_id+cluster_id 改为 config_entry_id FK
- [ ] **T2.5** 更新 model/config_permission.go — 保持 service_id+cluster_id 不变
- [ ] **T2.6** 更新 model/deployment.go — 删除 ~20 个字段（DirectMode, RawYAML, EnvVars, DeployVolumes, VolumeClaimTpls, SecretRefs, ConfigMapRefs, ServiceAccountName, ServiceSpec, LivenessProbe, ReadinessProbe, Command, Args, Port, CPURequest, MemRequest, CPULimit, MemLimit, Replicas）
- [ ] **T2.7** 更新 model/service.go — 删除 DefaultEnvVars, DefaultSecretRefs, DefaultConfigMapRefs
- [ ] **T2.8** 删除 model/service_config_env.go

## Phase 3: Repository Layer

- [ ] **T3.1** 创建 repository/config_entry_repo.go — 接口: List(svcID, clusterID), FindByID, FindByName(svcID, clusterID, name), Create, Update, Delete
- [ ] **T3.2** 创建 repository/config_entry_repo_impl.go — GORM 实现
- [ ] **T3.3** 创建 repository/service_config_ref_repo.go — 接口: List(svcID), FindByID, Create, Update, Delete
- [ ] **T3.4** 创建 repository/service_config_ref_repo_impl.go — GORM 实现
- [ ] **T3.5** 更新 repository/config_item_repo.go — 参数从 service_id+cluster_id 改为 entryID
- [ ] **T3.6** 更新 repository/config_item_repo_impl.go — WHERE config_entry_id = ?
- [ ] **T3.7** 更新 repository/config_release_repo.go — 参数从 service_id+cluster_id 改为 entryID
- [ ] **T3.8** 更新 repository/config_release_repo_impl.go — WHERE config_entry_id = ?
- [ ] **T3.9** 删除 repository/service_config_env_repo.go + impl

## Phase 4: Service Layer

- [ ] **T4.1** 重写 configcenter/config_service.go — 构造函数改用 configEntryRepo 替代 envRepo
- [ ] **T4.2** 实现 CreateEntry(svcID, clusterID, name, configType, format) — 校验 name 唯一 + type/format 枚举
- [ ] **T4.3** 实现 UpdateEntry(id, name, description) — 只允许改名称和描述
- [ ] **T4.4** 实现 DeleteEntry(id) — FK CASCADE 处理子表
- [ ] **T4.5** 实现 ListEntries(svcID, clusterID)
- [ ] **T4.6** 实现 GetEntry(id)
- [ ] **T4.7** 更新 ListItems/CreateItem/UpdateItem/DeleteItem — 参数改为 entryID，加密逻辑根据 entry.config_type 判断
- [ ] **T4.8** 更新 GetDraft/SaveDraft — 参数改为 entryID
- [ ] **T4.9** 更新 Publish — 参数改为 entryID，快照按 entry 粒度
- [ ] **T4.10** 更新 Rollback — 参数改为 entryID
- [ ] **T4.11** 更新 ListReleases — 参数改为 entryID
- [ ] **T4.12** 新增 GetPublishedSnapshot(entryID) — 查最新 published release 的 snapshot
- [ ] **T4.13** 新增 ListRefs(svcID) / CreateRef / UpdateRef / DeleteRef — ServiceConfigRef CRUD
- [ ] **T4.14** 更新 ConfigPermissionService — 保持 service_id+cluster_id 不变，构造函数去掉旧依赖

## Phase 5: Handler Layer

- [ ] **T5.1** 重写 handler/config_center.go — ConfigCenterHandler 注入 configSvc + permSvc + configRefRepo
- [ ] **T5.2** 实现 ListEntries — GET /services/:id/config-entries?cluster_id=X
- [ ] **T5.3** 实现 CreateEntry — POST /services/:id/config-entries
- [ ] **T5.4** 实现 GetEntry — GET /config-entries/:id
- [ ] **T5.5** 实现 UpdateEntry — PUT /config-entries/:id
- [ ] **T5.6** 实现 DeleteEntry — DELETE /config-entries/:id
- [ ] **T5.7** 实现 ListItems — GET /config-entries/:id/items
- [ ] **T5.8** 实现 CreateItem — POST /config-entries/:id/items
- [ ] **T5.9** 实现 UpdateItem — PUT /config-entries/:id/items/:item_id
- [ ] **T5.10** 实现 DeleteItem — DELETE /config-entries/:id/items/:item_id
- [ ] **T5.11** 实现 GetDraft — GET /config-entries/:id/draft
- [ ] **T5.12** 实现 SaveDraft — PUT /config-entries/:id/draft
- [ ] **T5.13** 实现 Publish — POST /config-entries/:id/release
- [ ] **T5.14** 实现 Rollback — POST /config-entries/:id/rollback
- [ ] **T5.15** 实现 ListReleases — GET /config-entries/:id/releases
- [ ] **T5.16** 实现 ListRefs — GET /services/:id/config-refs
- [ ] **T5.17** 实现 CreateRef — POST /services/:id/config-refs
- [ ] **T5.18** 实现 UpdateRef — PUT /services/:id/config-refs/:ref_id
- [ ] **T5.19** 实现 DeleteRef — DELETE /services/:id/config-refs/:ref_id
- [ ] **T5.20** 保持权限路由 /services/:id/configs/:cid/permissions
- [ ] **T5.21** RegisterConfigCenterRoutes 重写 + main.go 依赖注入更新
- [ ] **T5.22** 简化 deploy handler — createDeploymentRequest 删除配置覆盖字段
- [ ] **T5.23** 简化 deploy.go — DeployConfig 删除配置字段
- [ ] **T5.24** 编译 + 全量测试验证后端

## Phase 6: Deploy Integration

- [ ] **T6.1** 重写 deploy/config_deploy.go — GenerateConfigResources(svcID, clusterID, svcName, configSvc, refRepo)
- [ ] **T6.2** 实现 ref → entry → release → snapshot 查找链路
- [ ] **T6.3** 实现按 config_type 生成 K8s 资源: env→ConfigMap(envFrom), configmap→ConfigMap+volumeMount, secret→Secret+volumeMount
- [ ] **T6.4** 重写 config_executor.go — generateYAML 从 Service 读运行时参数，从 ConfigDeployHelper 获取配置资源
- [ ] **T6.5** 容器 spec 注入: envFrom(env 类型), volumeMounts(configmap/secret 类型，mount_path 从 ServiceConfigRef 读)
- [ ] **T6.6** 更新 DirectExecutor — 传递 configDeployHelper + configRefRepo
- [ ] **T6.7** 更新 main.go — 注入新依赖到 executor
- [ ] **T6.8** 编译 + 全量测试验证

## Phase 7: Frontend — 配置中心

- [ ] **T7.1** 更新 hooks/use-config-center.ts — 路由改为 /config-entries/:id/..., 新增 useConfigEntries, useCreateEntry, useDeleteEntry
- [ ] **T7.2** 新增 useServiceConfigRefs, useCreateRef, useUpdateRef, useDeleteRef hooks
- [ ] **T7.3** 更新 types/index.ts — 新增 ConfigEntry, ServiceConfigRef 类型; 删除 ServiceConfigEnv; 更新 ConfigItem/ConfigRelease (entry_id)
- [ ] **T7.4** 重写 configs/[serviceId]/page.tsx — 左侧环境列表 + 右侧条目列表/详情切换
- [ ] **T7.5** 创建 components/config/entry-list.tsx — 条目表格(名称/类型/格式/最新版本/操作)
- [ ] **T7.6** 创建 components/config/create-entry-dialog.tsx — 新建条目弹窗(名称/类型/格式)
- [ ] **T7.7** 创建 components/config/entry-detail.tsx — 条目编辑面板(复用 kv-table-editor + code-editor + 发布/回滚/历史 tabs)
- [ ] **T7.8** 更新 kv-table-editor.tsx — props 从 serviceId+clusterId 改为 entryId
- [ ] **T7.9** 更新 publish-dialog.tsx — props 改为 entryId
- [ ] **T7.10** 更新 rollback-dialog.tsx — props 改为 entryId
- [ ] **T7.11** 更新 release-history.tsx — props 改为 entryId
- [ ] **T7.12** 更新 config-content.tsx — 适配条目模式（或由 entry-detail 替代）

## Phase 8: Frontend — 服务编辑 + 部署对话框

- [ ] **T8.1** 创建 components/service/config-ref-editor.tsx — 配置引用管理(表格: 配置名称/挂载路径/操作 + 添加按钮)
- [ ] **T8.2** 在 service-create-dialog.tsx 中集成 ConfigRefEditor — deploy_type=direct 时显示
- [ ] **T8.3** 简化 deploy-dialog.tsx — 删除 DirectConfigForm 引入和配置模式表单
- [ ] **T8.4** 简化 deploy-dialog.tsx — 配置部署模式只保留: 服务 → 集群+命名空间 → 镜像 → 确认
- [ ] **T8.5** 更新 use-deployments.ts — 删除配置覆盖字段(env_vars, volumes 等)

## Phase 9: Cleanup

- [ ] **T9.1** 删除 model/service_config_env.go
- [ ] **T9.2** 删除 repository/service_config_env_repo.go + impl
- [ ] **T9.3** 删除 components/config/init-config-form.tsx
- [ ] **T9.4** 删除 components/deploy/direct-config-form.tsx
- [ ] **T9.5** 简化 components/service/service-runtime-config.tsx — 移除 env/secret/configmap 相关字段(保留 port/replicas/cpu/mem/probe/command/volumes/SA)
- [ ] **T9.6** 删除 Deployment type 中已移除的字段
- [ ] **T9.7** 编译前后端 + 全量测试验证

---

**总计: 78 个任务**
- Phase 1: 14 tasks (DB Migrations)
- Phase 2: 8 tasks (Models)
- Phase 3: 9 tasks (Repository)
- Phase 4: 14 tasks (Service)
- Phase 5: 24 tasks (Handler)
- Phase 6: 8 tasks (Deploy Integration)
- Phase 7: 12 tasks (Frontend Config)
- Phase 8: 5 tasks (Frontend Service+Deploy)
- Phase 9: 7 tasks (Cleanup)
