# DeployHub Backend

Go 后端服务，提供 RESTful API + WebSocket 实时推送。

## 技术栈

- **Go 1.23+** · Gin (HTTP) · GORM (ORM) · client-go (K8s)
- **PostgreSQL** 数据持久化 · **Redis** 会话/缓存
- **gorilla/websocket** 实时日志和部署进度
- **golang-migrate** 数据库 Schema 管理
- **AES-256-GCM** 敏感数据加密 · **BCrypt** 密码哈希 · **JWT** 认证

## 开发

```bash
# 安装依赖
go mod download

# 运行迁移
make migrate-up

# 编译
make build

# 运行
./bin/server

# 测试
make test
```

## API 概览

### 认证
- `POST /api/v1/auth/register` — 注册
- `POST /api/v1/auth/login` — 登录
- `GET /api/v1/auth/me` — 当前用户

### 服务管理
- `GET/POST /api/v1/services` — 服务列表/创建
- `GET/PUT/DELETE /api/v1/services/:id` — 服务详情/更新/删除
- `GET/POST/PUT/DELETE /api/v1/services/:id/members[/:mid]` — 成员管理

### 构建中心
- `GET/POST /api/v1/builds` — 构建列表/触发
- `POST /api/v1/builds/:id/cancel` — 取消构建
- `WS /ws/builds/:id/log` — 实时构建日志

### 发布管理
- `GET/POST /api/v1/deployments` — 部署列表/创建
- `POST /api/v1/deployments/:id/execute` — 执行部署
- `POST /api/v1/deployments/:id/cancel` — 取消
- `POST /api/v1/deployments/:id/rollback` — 回滚
- `WS /ws/deployments/:id/progress` — 实时部署进度
- `WS /ws/deployments/:id/pod-logs` — Pod 日志流

### 配置中心
- `GET/POST /api/v1/services/:id/config-entries` — 配置条目
- `GET/PUT/DELETE /api/v1/config-entries/:id` — 条目操作
- `GET/POST/PUT/DELETE /api/v1/config-entries/:id/items[/:iid]` — 配置项 CRUD
- `POST /api/v1/config-entries/:id/release` — 发布
- `POST /api/v1/config-entries/:id/rollback` — 回滚

### 路由中心
- `GET/POST /api/v1/route-entries` — 路由条目
- `POST /api/v1/route-entries/:id/deploy` — 部署路由到集群
- `GET /api/v1/route-entries/:id/preview` — YAML 预览

### 插件中心
- `GET/POST /api/v1/route-plugins` — 插件
- `POST /api/v1/route-plugins/:id/deploy` — 部署插件到集群

### 通知中心
- `GET/POST /api/v1/notification-rules` — 全局通知规则
- `GET/POST /api/v1/service-notification-rules` — 服务级规则

### 系统设置
- `GET/POST /api/v1/clusters` — 集群管理
- `GET/POST/DELETE /api/v1/clusters/:id/namespaces[/:ns_id]` — 集群可发布 namespace 映射（admin 写）
- `POST /api/v1/clusters/:id/namespaces/sync` — 从集群拉取实际 namespace 同步登记（admin）
- `GET/POST /api/v1/git-repos` — Git 仓库
- `GET/POST /api/v1/registries` — 镜像仓库
- `GET/PUT /api/v1/system-settings[/:key]` — 系统配置

> 创建/回滚部署时，后端会强制校验请求中的 `namespace` 是否登记在 `cluster_namespaces(cluster_id, namespace)` 中。未登记将返回 400，杜绝脚本/旧客户端绕过前端下拉直接发到任意 namespace。

## 目录结构

```
backend/
├── cmd/server/main.go       # 入口 + DI
├── internal/
│   ├── config/              # 环境变量配置
│   ├── handler/             # HTTP Handler（~15 个文件）
│   ├── middleware/          # JWT/CORS/AuditLog/RBAC
│   ├── model/               # GORM 模型（~25 个）
│   ├── pkg/                 # 通用工具
│   ├── repository/          # 数据访问接口+实现（~30 对）
│   ├── service/             # 业务逻辑（~12 个包）
│   └── ws/                  # WebSocket Hub
├── migrations/              # SQL 迁移（000001-000068）
├── Makefile
├── go.mod / go.sum
└── .env.example
```
