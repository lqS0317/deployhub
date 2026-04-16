# Feature 021: 路由中心 + 插件中心 — 实现计划

## Phase 1: DB Migrations

- **000064_create_route_entries** — route_entries 表 (name, resource_type, config jsonb, created_by_id)，UNIQUE(name, resource_type)
- **000065_create_route_deployments** — route_deployments 表 (route_entry_id FK CASCADE, cluster_id, namespace, status, config_snapshot jsonb, rendered_yaml, error_msg, deployed_at)，UNIQUE(route_entry_id, cluster_id, namespace)
- **000066_create_route_permissions** — route_permissions 表 (cluster_id, user_id, role)，UNIQUE(cluster_id, user_id)
- **000067_create_route_plugins** — route_plugins 表 (name UNIQUE, description, yaml_content, created_by_id)
- **000068_create_plugin_deployments** — plugin_deployments 表 (plugin_id FK CASCADE, cluster_id, namespace, status, yaml_snapshot, error_msg, deployed_at)，UNIQUE(plugin_id, cluster_id, namespace)

## Phase 2: Backend Models

- model/route_entry.go — RouteEntry 结构体 + resource_type 常量
- model/route_deployment.go — RouteDeployment 结构体 + status 常量
- model/route_permission.go — RoutePermission 结构体
- model/route_plugin.go — RoutePlugin 结构体
- model/plugin_deployment.go — PluginDeployment 结构体

## Phase 3: Repository Layer

- repository/route_entry_repo.go + impl — List(resourceType), FindByID, FindByNameAndType, Create, Update, Delete
- repository/route_deployment_repo.go + impl — ListByEntry(entryID), FindByEntryClusterNs, Upsert, Delete
- repository/route_permission_repo.go + impl — List, FindByClusterAndUser, Upsert, Delete
- repository/route_plugin_repo.go + impl — List, FindByID, FindByName, Create, Update, Delete
- repository/plugin_deployment_repo.go + impl — ListByPlugin(pluginID), FindByPluginClusterNs, Upsert, Delete

## Phase 4: YAML Generators

新建 service/routing/yaml_builder.go，4 个生成函数：

### 4.1 buildK8sServiceYAML
- 输入: name, namespace, config(type/selector/ports)
- 输出: core/v1 Service YAML
- 使用 corev1.Service 结构体 → yaml.Marshal

### 4.2 buildIngressYAML
- 输入: name, namespace, config(ingressClassName/tls/rules/annotations)
- 输出: networking.k8s.io/v1 Ingress YAML
- 使用 networkingv1.Ingress 结构体

### 4.3 buildIngressRouteYAML
- 输入: name, namespace, config(entryPoints/routes/tls)
- 输出: traefik.io/v1alpha1 IngressRoute YAML（Unstructured map）
- 因为没有 traefik 的 Go types，用 unstructured.Unstructured 手动构建

### 4.4 buildApisixRouteYAML
- 输入: name, namespace, config(rules)
- 输出: apisix.apache.org/v2 ApisixRoute YAML（Unstructured map）
- 同样用 Unstructured 构建

## Phase 5: K8s Deploy Helper

新建 service/routing/k8s_deployer.go：

- DeployResource(clusterID uint, resourceType, yamlContent string) error
  - Service → clientset.CoreV1().Services(ns).Apply
  - Ingress → clientset.NetworkingV1().Ingresses(ns).Apply
  - IngressRoute/ApisixRoute → dynamic client server-side apply
- DeleteResource(clusterID uint, resourceType, name, namespace string) error
  - 从集群删除对应资源

复用现有 ClientsetPool.GetClientset 和 GetRestConfig。

## Phase 6: Service Layer

### 6.1 RouteService
- CRUD: CreateEntry, UpdateEntry, DeleteEntry, GetEntry, ListEntries
- Deploy: DeployEntry(entryID, clusterID, namespace) — 生成 YAML → apply → 记录 RouteDeployment
- Preview: PreviewEntry(entryID, namespace) — 只生成 YAML 不 apply
- GetDeployments: 列出 entry 的部署记录
- DeleteEntry 时如有已部署环境，可选删除 K8s 资源

### 6.2 PluginService
- CRUD: CreatePlugin, UpdatePlugin, DeletePlugin, GetPlugin, ListPlugins
- Deploy: DeployPlugin(pluginID, clusterID, namespace) — apply YAML → 记录 PluginDeployment
- GetDeployments

### 6.3 RoutePermissionService
- CheckPermission(clusterID, userID, requiredRole) — Admin 通过 → 查表 → 角色层级
- List, Grant, Revoke

## Phase 7: Handler Layer

### 7.1 route_entry_handler.go
- GET /route-entries → List (query param: resource_type)
- POST /route-entries → Create
- GET /route-entries/:id → Get
- PUT /route-entries/:id → Update
- DELETE /route-entries/:id → Delete
- POST /route-entries/:id/deploy → Deploy {cluster_id, namespace}
- GET /route-entries/:id/deployments → ListDeployments
- GET /route-entries/:id/preview → Preview {cluster_id, namespace}

### 7.2 route_plugin_handler.go
- GET /route-plugins → List
- POST /route-plugins → Create
- GET /route-plugins/:id → Get
- PUT /route-plugins/:id → Update
- DELETE /route-plugins/:id → Delete
- POST /route-plugins/:id/deploy → Deploy {cluster_id, namespace}
- GET /route-plugins/:id/deployments → ListDeployments

### 7.3 route_permission_handler.go
- GET /route-permissions → List
- POST /route-permissions → Grant
- DELETE /route-permissions/:id → Revoke

### 7.4 路由注册 + main.go 依赖注入

## Phase 8: Frontend — 插件中心（先做，路由中心依赖它）

### 8.1 Hooks
- use-route-plugins.ts — usePlugins, useCreatePlugin, useUpdatePlugin, useDeletePlugin, useDeployPlugin, usePluginDeployments

### 8.2 页面 + 组件
- /plugins/page.tsx — 插件列表页
- components/plugin/create-plugin-dialog.tsx — 名称+描述+YAML 编辑器
- components/plugin/deploy-plugin-dialog.tsx — 选集群+命名空间 → 确认
- 列表显示部署状态（已部署 N 个环境 / 有更新未部署）

## Phase 9: Frontend — 路由中心

### 9.1 Hooks
- use-route-entries.ts — useRouteEntries(resourceType), useCreateEntry, useUpdateEntry, useDeleteEntry, useDeployEntry, usePreviewEntry, useEntryDeployments
- use-route-permissions.ts — useRoutePermissions, useGrantPermission, useRevokePermission

### 9.2 页面
- /routes/page.tsx — 4-Tab 布局 + 条目列表 + 部署状态

### 9.3 创建/编辑弹窗
- components/route/create-route-dialog.tsx — 根据 resource_type 动态切换表单
- components/route/service-form.tsx — K8s Service 表单（type/selector/ports）
- components/route/ingress-form.tsx — Ingress 表单（className/tls/rules/annotations）
- components/route/ingressroute-form.tsx — Traefik IngressRoute 表单（entryPoints/routes/middlewares/tls）
- components/route/apisixroute-form.tsx — APISIX ApisixRoute 表单（rules/plugins）

### 9.4 部署 + 预览
- components/route/deploy-route-dialog.tsx — 选集群+命名空间 → 预览 YAML → 确认
- components/route/yaml-preview.tsx — YAML 预览面板

### 9.5 middleware/plugin 选择
- IngressRoute 表单中 middlewares 字段 → 从插件中心拉取 Middleware 列表多选
- ApisixRoute 表单中 plugins 字段 → 从插件中心拉取 Plugin 列表多选

## Phase 10: Sidebar 导航

- 更新 layout.tsx 侧边栏新增：路由中心(/routes)、插件中心(/plugins)

## Constitution Check

- ✅ 安全第一: RoutePermission 权限控制
- ✅ 分层架构: handler → service → repository
- ✅ 多集群隔离: 按 cluster_id 部署，独立管理
- ✅ 简洁优先: 手动部署模式，不做版本化
- ✅ CRD 支持: dynamic client for IngressRoute/ApisixRoute
