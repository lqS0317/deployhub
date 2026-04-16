# Feature 021: 路由中心 + 插件中心 — 任务列表

## Phase 1: DB Migrations

- [ ] **T1.1** 000064_create_route_entries.up.sql — route_entries (id, name, resource_type, config jsonb, created_by_id, timestamps)，UNIQUE(name, resource_type)
- [ ] **T1.2** 000064_create_route_entries.down.sql
- [ ] **T1.3** 000065_create_route_deployments.up.sql — route_deployments (id, route_entry_id FK CASCADE, cluster_id, namespace, status, config_snapshot jsonb, rendered_yaml text, error_msg text, deployed_at)，UNIQUE(route_entry_id, cluster_id, namespace)
- [ ] **T1.4** 000065_create_route_deployments.down.sql
- [ ] **T1.5** 000066_create_route_permissions.up.sql — route_permissions (id, cluster_id, user_id, role)，UNIQUE(cluster_id, user_id)
- [ ] **T1.6** 000066_create_route_permissions.down.sql
- [ ] **T1.7** 000067_create_route_plugins.up.sql — route_plugins (id, name UNIQUE, description, yaml_content, created_by_id, timestamps)
- [ ] **T1.8** 000067_create_route_plugins.down.sql
- [ ] **T1.9** 000068_create_plugin_deployments.up.sql — plugin_deployments (id, plugin_id FK CASCADE, cluster_id, namespace, status, yaml_snapshot, error_msg, deployed_at)，UNIQUE(plugin_id, cluster_id, namespace)
- [ ] **T1.10** 000068_create_plugin_deployments.down.sql

## Phase 2: Backend Models

- [ ] **T2.1** model/route_entry.go — RouteEntry 结构体 + resource_type 常量 (RouteTypeService/Ingress/IngressRoute/ApisixRoute)
- [ ] **T2.2** model/route_deployment.go — RouteDeployment 结构体 + status 常量 (deployed/failed)
- [ ] **T2.3** model/route_permission.go — RoutePermission 结构体 + role 常量
- [ ] **T2.4** model/route_plugin.go — RoutePlugin 结构体
- [ ] **T2.5** model/plugin_deployment.go — PluginDeployment 结构体

## Phase 3: Repository Layer

- [ ] **T3.1** repository/route_entry_repo.go — 接口: List(resourceType), FindByID, FindByNameAndType, Create, Update, Delete
- [ ] **T3.2** repository/route_entry_repo_impl.go — GORM 实现，List 按 resource_type 过滤
- [ ] **T3.3** repository/route_deployment_repo.go — 接口: ListByEntry(entryID), FindByEntryClusterNs, Upsert, DeleteByEntry(entryID)
- [ ] **T3.4** repository/route_deployment_repo_impl.go — Upsert 用 OnConflict
- [ ] **T3.5** repository/route_permission_repo.go — 接口: List, FindByClusterAndUser, Upsert, Delete
- [ ] **T3.6** repository/route_permission_repo_impl.go
- [ ] **T3.7** repository/route_plugin_repo.go — 接口: List, FindByID, FindByName, Create, Update, Delete
- [ ] **T3.8** repository/route_plugin_repo_impl.go
- [ ] **T3.9** repository/plugin_deployment_repo.go — 接口: ListByPlugin(pluginID), FindByPluginClusterNs, Upsert, DeleteByPlugin(pluginID)
- [ ] **T3.10** repository/plugin_deployment_repo_impl.go

## Phase 4: YAML Generators

- [ ] **T4.1** service/routing/yaml_builder.go — 定义 config 解析结构体 (ServiceConfig, IngressConfig, IngressRouteConfig, ApisixRouteConfig)
- [ ] **T4.2** 实现 BuildK8sServiceYAML(name, ns, config) — 解析 config → corev1.Service → yaml.Marshal
- [ ] **T4.3** 实现 BuildIngressYAML(name, ns, config) — 解析 config → networkingv1.Ingress → yaml.Marshal
- [ ] **T4.4** 实现 BuildIngressRouteYAML(name, ns, config) — 解析 config → unstructured.Unstructured (traefik.io/v1alpha1 IngressRoute) → yaml.Marshal
- [ ] **T4.5** 实现 BuildApisixRouteYAML(name, ns, config) — 解析 config → unstructured.Unstructured (apisix.apache.org/v2 ApisixRoute) → yaml.Marshal
- [ ] **T4.6** 实现 BuildYAML(name, ns, resourceType, config) — 路由分派函数

## Phase 5: K8s Deploy Helper

- [ ] **T5.1** service/routing/k8s_deployer.go — K8sRouteDeployer 结构体，注入 ClientsetPool
- [ ] **T5.2** 实现 ApplyResource(clusterID, resourceType, name, namespace, yamlContent) — Service/Ingress 用 typed apply，IngressRoute/ApisixRoute 用 dynamic apply
- [ ] **T5.3** 实现 DeleteResource(clusterID, resourceType, name, namespace) — 从集群删除
- [ ] **T5.4** 实现 ApplyPluginYAML(clusterID, namespace, yamlContent) — 解析 YAML → dynamic apply（通用 CRD）

## Phase 6: Service Layer

- [ ] **T6.1** service/routing/route_service.go — RouteService 结构体（注入 repos + deployer + yamlBuilder）
- [ ] **T6.2** 实现 CreateEntry, UpdateEntry, DeleteEntry, GetEntry, ListEntries
- [ ] **T6.3** 实现 DeployEntry(entryID, clusterID, namespace) — 生成 YAML → apply → upsert RouteDeployment
- [ ] **T6.4** 实现 PreviewEntry(entryID, namespace) — 只生成 YAML 返回
- [ ] **T6.5** 实现 GetDeployments(entryID) — 列出部署记录
- [ ] **T6.6** 实现 HasPendingUpdates(entryID) — 对比当前 config 与已部署 config_snapshot
- [ ] **T6.7** service/routing/plugin_service.go — PluginService 结构体
- [ ] **T6.8** 实现 CreatePlugin, UpdatePlugin, DeletePlugin, GetPlugin, ListPlugins
- [ ] **T6.9** 实现 DeployPlugin(pluginID, clusterID, namespace) — apply YAML → upsert PluginDeployment
- [ ] **T6.10** 实现 GetPluginDeployments(pluginID)
- [ ] **T6.11** service/routing/permission_service.go — RoutePermissionService
- [ ] **T6.12** 实现 CheckPermission(clusterID, userID, requiredRole) — Admin 通过 → 查表 → 角色层级
- [ ] **T6.13** 实现 List, Grant, Revoke

## Phase 7: Handler Layer

- [ ] **T7.1** handler/route_entry.go — RouteEntryHandler 结构体
- [ ] **T7.2** 实现 ListEntries — GET /route-entries
- [ ] **T7.3** 实现 CreateEntry — POST /route-entries
- [ ] **T7.4** 实现 GetEntry — GET /route-entries/:id
- [ ] **T7.5** 实现 UpdateEntry — PUT /route-entries/:id
- [ ] **T7.6** 实现 DeleteEntry — DELETE /route-entries/:id
- [ ] **T7.7** 实现 DeployEntry — POST /route-entries/:id/deploy {cluster_id, namespace}
- [ ] **T7.8** 实现 ListDeployments — GET /route-entries/:id/deployments
- [ ] **T7.9** 实现 PreviewEntry — GET /route-entries/:id/preview?namespace=X
- [ ] **T7.10** handler/route_plugin.go — RoutePluginHandler 结构体
- [ ] **T7.11** 实现 ListPlugins, CreatePlugin, GetPlugin, UpdatePlugin, DeletePlugin
- [ ] **T7.12** 实现 DeployPlugin — POST /route-plugins/:id/deploy
- [ ] **T7.13** 实现 ListPluginDeployments — GET /route-plugins/:id/deployments
- [ ] **T7.14** handler/route_permission.go — RoutePermissionHandler（List/Grant/Revoke）
- [ ] **T7.15** RegisterRouteRoutes + RegisterPluginRoutes 路由注册
- [ ] **T7.16** 更新 main.go — 初始化 repos/services/handlers，注册路由
- [ ] **T7.17** 编译 + 全量测试验证后端

## Phase 8: Frontend — 插件中心

- [ ] **T8.1** hooks/use-route-plugins.ts — usePlugins, useCreatePlugin, useUpdatePlugin, useDeletePlugin, useDeployPlugin, usePluginDeployments
- [ ] **T8.2** types/index.ts — 新增 RoutePlugin, PluginDeployment 类型
- [ ] **T8.3** app/(dashboard)/plugins/page.tsx — 插件列表页（表格: 名称/描述/部署状态/操作）
- [ ] **T8.4** components/plugin/create-plugin-dialog.tsx — 创建/编辑弹窗（名称+描述+YAML 编辑器）
- [ ] **T8.5** components/plugin/deploy-plugin-dialog.tsx — 部署弹窗（选集群+命名空间 → 确认）
- [ ] **T8.6** 部署状态显示（已部署 N 个环境 / 有未部署更新标记）

## Phase 9: Frontend — 路由中心

- [ ] **T9.1** hooks/use-route-entries.ts — useRouteEntries(type), useCreateEntry, useUpdateEntry, useDeleteEntry, useDeployEntry, usePreviewEntry, useEntryDeployments
- [ ] **T9.2** types/index.ts — 新增 RouteEntry, RouteDeployment, RoutePermission 类型
- [ ] **T9.3** app/(dashboard)/routes/page.tsx — 4-Tab 布局 (Service/Ingress/IngressRoute/ApisixRoute) + 条目列表
- [ ] **T9.4** components/route/create-route-dialog.tsx — 创建/编辑弹窗框架（根据 type 切换表单）
- [ ] **T9.5** components/route/service-form.tsx — K8s Service 表单 (type/selector KV/ports 动态行)
- [ ] **T9.6** components/route/ingress-form.tsx — Ingress 表单 (className/tls/rules+paths 动态行/annotations KV)
- [ ] **T9.7** components/route/ingressroute-form.tsx — Traefik IngressRoute 表单 (entryPoints/routes+match+services/middlewares 多选)
- [ ] **T9.8** components/route/apisixroute-form.tsx — APISIX ApisixRoute 表单 (rules+host+paths+backend/plugins 多选)
- [ ] **T9.9** components/route/deploy-route-dialog.tsx — 部署弹窗 (选集群+命名空间 → 预览 YAML → 确认)
- [ ] **T9.10** components/route/yaml-preview.tsx — YAML 预览面板（黑底代码块）
- [ ] **T9.11** IngressRoute/ApisixRoute 表单中 middleware/plugin 字段 — 从 usePlugins 拉取下拉多选

## Phase 10: Sidebar + 清理

- [ ] **T10.1** 更新 layout.tsx 侧边栏 — 新增路由中心(/routes) + 插件中心(/plugins)
- [ ] **T10.2** 编译前后端 + 全量测试验证

---

**总计: 72 个任务**
- Phase 1: 10 tasks (DB Migrations)
- Phase 2: 5 tasks (Models)
- Phase 3: 10 tasks (Repository)
- Phase 4: 6 tasks (YAML Generators)
- Phase 5: 4 tasks (K8s Deploy Helper)
- Phase 6: 13 tasks (Service Layer)
- Phase 7: 17 tasks (Handler)
- Phase 8: 6 tasks (Frontend Plugin)
- Phase 9: 11 tasks (Frontend Route)
- Phase 10: 2 tasks (Sidebar)
