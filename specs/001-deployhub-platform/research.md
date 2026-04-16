# DeployHub 技术调研与决策记录

本文档记录 DeployHub 平台已确定的技术选型，采用标准调研格式：**决策**、**理由**、**曾考虑的替代方案**。

---

## 1. HTTP 框架

- **Decision**: Gin
- **Rationale**: 轻量、性能好，中间件生态成熟，便于统一鉴权、日志与错误处理。
- **Alternatives considered**: Echo、Fiber、`net/http`

---

## 2. ORM

- **Decision**: GORM
- **Rationale**: Go 生态成熟 ORM，支持自动迁移与钩子，开发效率与可维护性平衡较好。
- **Alternatives considered**: sqlx、ent、sqlc

---

## 3. 数据库迁移工具

- **Decision**: golang-migrate
- **Rationale**: 以 SQL 为源，提供 CLI 与库调用，up/down 成对管理，便于审计与回滚。
- **Alternatives considered**: goose、atlas、GORM AutoMigrate

---

## 4. Kubernetes 客户端

- **Decision**: client-go
- **Rationale**: 官方 SDK，API 覆盖完整，支持 Watch 等控制面能力，与集群交互路径清晰。
- **Alternatives considered**: controller-runtime、`kubectl` 风格 exec

---

## 5. WebSocket

- **Decision**: gorilla/websocket
- **Rationale**: 成熟、使用广泛，符合 RFC 6455，满足实时日志与事件推送场景。
- **Alternatives considered**: nhooyr/websocket、SSE

---

## 6. 容器镜像构建

- **Decision**: Kaniko
- **Rationale**: 无需宿主机 Docker daemon，适合以 K8s Job 运行，支持层缓存与流水线集成。
- **Alternatives considered**: BuildKit、Docker-in-Docker、Buildpacks

---

## 7. 认证与授权

- **Decision**: API 使用 JWT；支持 OAuth2 多提供商；本地账号密码使用 BCrypt
- **Rationale**: JWT 适合无状态 API；OAuth2 覆盖企业 IdP；BCrypt 为密码存储常见安全基线。
- **Alternatives considered**: 纯 Session、仅 OIDC、Casbin 作为统一授权层

---

## 8. 加密

- **Decision**: AES-256-GCM（Go `crypto/aes` + `crypto/cipher`，AEAD）
- **Rationale**: 标准库即可实现认证加密，无额外依赖，密钥由配置或密钥管理注入，不硬编码。
- **Alternatives considered**: age、Vault Transit

---

## 9. 前端框架

- **Decision**: Next.js 14 App Router + React 18
- **Rationale**: SSR/SSG 灵活，App Router 利于布局与路由组织，与团队 React 栈一致。
- **Alternatives considered**: Vite + React、Remix、Vue

---

## 10. UI 组件与样式

- **Decision**: Tailwind CSS + shadcn/ui
- **Rationale**: 组件可复制到仓库内维护，可访问性基础好，暗色模式与主题扩展顺手。
- **Alternatives considered**: Ant Design、Material UI、Chakra UI

---

## 11. 前端状态管理

- **Decision**: TanStack Query（服务端状态）
- **Rationale**: 缓存、失效、重试与乐观更新与部署/流水线类数据拉取场景匹配。
- **Alternatives considered**: SWR、Redux Toolkit Query、Zustand（偏客户端状态）

---

## 12. 数据存储

- **Decision**: PostgreSQL 15+；Redis 7+
- **Rationale**: PostgreSQL 的 jsonb、CTE 等适合配置与审计类查询；Redis 用于会话、缓存及 WebSocket 广播的 pub/sub。
- **Alternatives considered**: MySQL、SQLite；Memcached、KeyDB

---

## 13. 部署策略

- **Decision**: 仅 Rolling Update，通过 client-go 对 Deployment 打补丁驱动
- **Rationale**: 实现简单，完全依赖 K8s 原生行为，满足当前发布节奏与运维复杂度。
- **Alternatives considered**: Argo Rollouts、Istio 金丝雀

---

## 14. 配置管理

- **Decision**: 内置 Go template 渲染 + 按环境变量 + 版本追踪 + 与 K8s ConfigMap/Secret 同步
- **Rationale**: 减少外部配置中心依赖，模板与集群对象对齐，便于 GitOps 与审计。
- **Alternatives considered**: Apollo、Nacos、Consul

---

## 15. RBAC 模型

- **Decision**: 服务级角色（owner / developer / viewer），面向少于 50 个服务的规模
- **Rationale**: 模型简单，权限边界清晰，与产品粒度一致，避免过度工程。
- **Alternatives considered**: Casbin 策略引擎、直接映射 K8s RBAC、全局角色矩阵

---

## 16. 通知

- **Decision**: 应用内通知（DB 持久化）+ Webhook 对接飞书 / 钉钉 / Slack
- **Rationale**: 站内可追溯；Webhook 覆盖主流协作工具，无需自建消息队列即可集成。
- **Alternatives considered**: 邮件、移动推送、以消息队列为中心的事件总线

---

## 17. 项目架构

- **Decision**: 分层架构 handler → service → repository，接口解耦
- **Rationale**: 职责清晰，测试与替换实现（如 mock repository）成本低，符合中等规模后端演进。
- **Alternatives considered**: Clean Architecture、六边形、CQRS

---

*本文件仅记录已定决策，后续若范围扩大（例如多集群、复杂发布），应复审并更新对应条目。*
