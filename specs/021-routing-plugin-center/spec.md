# Feature 021: 路由中心 + 插件中心

## 概述

新增路由中心和插件中心两个独立功能模块。路由中心管理 K8s Service/Ingress/IngressRoute/ApisixRoute 四种网络资源，插件中心管理 Traefik Middleware/APISIX Plugin 等 CRD 资源。两者都采用"编辑存 DB + 手动部署到集群"的模式。

## 数据模型

### RouteEntry（路由条目）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | |
| name | varchar(100) | 路由名称 |
| resource_type | varchar(20) | service/ingress/ingressroute/apisixroute |
| config | jsonb | 结构化配置（按 resource_type 不同） |
| created_by_id | uint FK(users) | |
| created_at, updated_at | timestamp | |

UNIQUE(name, resource_type)

### RouteDeployment（路由部署记录）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | |
| route_entry_id | uint FK CASCADE | |
| cluster_id | uint | 目标集群 |
| namespace | varchar(100) | 目标命名空间 |
| status | varchar(20) | deployed/failed |
| config_snapshot | jsonb | 部署时的配置快照 |
| rendered_yaml | text | 生成的 YAML |
| error_msg | text | 失败原因 |
| deployed_at | timestamp | |

UNIQUE(route_entry_id, cluster_id, namespace)

### RoutePermission（路由权限）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | |
| cluster_id | uint | 0=全局 |
| user_id | uint FK(users) | |
| role | varchar(20) | viewer/editor/publisher |
| created_at | timestamp | |

UNIQUE(cluster_id, user_id)

### RoutePlugin（插件）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | |
| name | varchar(100) UNIQUE | 插件名称 |
| description | text | |
| yaml_content | text | YAML 内容 |
| created_by_id | uint FK(users) | |
| created_at, updated_at | timestamp | |

### PluginDeployment（插件部署记录）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | |
| plugin_id | uint FK CASCADE | |
| cluster_id | uint | |
| namespace | varchar(100) | |
| status | varchar(20) | deployed/failed |
| yaml_snapshot | text | 部署时的 YAML 快照 |
| error_msg | text | |
| deployed_at | timestamp | |

UNIQUE(plugin_id, cluster_id, namespace)

## Config JSONB 结构（按 resource_type）

### service
```json
{
  "type": "ClusterIP",
  "selector": {"app": "my-app"},
  "ports": [{"name": "http", "port": 80, "targetPort": 8080, "protocol": "TCP"}]
}
```

### ingress
```json
{
  "ingressClassName": "nginx",
  "tls": [{"hosts": ["example.com"], "secretName": "tls-secret"}],
  "rules": [{"host": "example.com", "paths": [{"path": "/", "pathType": "Prefix", "backendService": "my-svc", "backendPort": 80}]}],
  "annotations": {"nginx.ingress.kubernetes.io/rewrite-target": "/"}
}
```

### ingressroute (Traefik)
```json
{
  "entryPoints": ["web", "websecure"],
  "routes": [{"match": "Host(`example.com`)", "services": [{"name": "my-svc", "port": 80}], "middlewares": ["auth-middleware"]}],
  "tls": {"certResolver": "letsencrypt"}
}
```

### apisixroute (APISIX)
```json
{
  "rules": [{"host": "example.com", "http": {"paths": [{"path": "/*", "backend": {"serviceName": "my-svc", "servicePort": 80}, "plugins": [{"name": "cors"}]}]}}]
}
```

## API

```
# 路由中心
GET    /route-entries?resource_type=X          — 列表（支持筛选）
POST   /route-entries                          — 创建
GET    /route-entries/:id                      — 详情
PUT    /route-entries/:id                      — 更新
DELETE /route-entries/:id                      — 删除
POST   /route-entries/:id/deploy               — 部署 {cluster_id, namespace}
GET    /route-entries/:id/deployments           — 部署状态列表
GET    /route-entries/:id/preview               — 预览 YAML {cluster_id, namespace}

# 路由权限
GET    /route-permissions                       — 列表
POST   /route-permissions                       — 授权
DELETE /route-permissions/:id                   — 移除

# 插件中心
GET    /route-plugins                           — 列表
POST   /route-plugins                           — 创建
GET    /route-plugins/:id                       — 详情
PUT    /route-plugins/:id                       — 更新
DELETE /route-plugins/:id                       — 删除
POST   /route-plugins/:id/deploy                — 部署 {cluster_id, namespace}
GET    /route-plugins/:id/deployments            — 部署状态列表
```

## YAML 生成器

4 个函数：config JSONB → K8s YAML
- `buildK8sServiceYAML(name, namespace string, config)` → core/v1 Service
- `buildIngressYAML(name, namespace string, config)` → networking.k8s.io/v1 Ingress
- `buildIngressRouteYAML(name, namespace string, config)` → traefik.io/v1alpha1 IngressRoute（Unstructured）
- `buildApisixRouteYAML(name, namespace string, config)` → apisix.apache.org/v2 ApisixRoute（Unstructured）

部署执行：
- Service/Ingress → client-go typed client
- IngressRoute/ApisixRoute → dynamic client (server-side apply)

## 前端

侧边栏新增：路由中心(/routes)、插件中心(/plugins)

路由中心页面：
- 顶部 Tab: Service | Ingress | IngressRoute | ApisixRoute
- 条目列表表格
- 创建/编辑弹窗（按类型动态表单）
- 部署弹窗（选集群+命名空间 → 预览 YAML → 确认）

插件中心页面：
- 插件列表表格
- 创建/编辑弹窗（名称+描述+YAML 编辑器）
- 部署弹窗

## 非目标

- 自动发现集群已有路由
- TLS 证书管理
- 版本化发布/回滚
- 和服务部署联动
- 插件按环境差异化

## 风险

- CRD 需要目标集群已安装对应 Ingress Controller
- dynamic client 需要正确 GVR
