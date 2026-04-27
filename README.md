# DeployHub

面向小团队（5-20 人）的 Kubernetes 原生运维发布平台，管理 <50 个微服务的完整生命周期。

## 功能概览

| 模块 | 功能 |
|------|------|
| **服务管理** | 服务 CRUD、成员权限、运行时参数（端口/副本/资源/探针/启动命令） |
| **构建中心** | Kaniko 集群内构建、Git 分支+Commit 选择、实时日志流 |
| **发布管理** | Direct 配置部署 / YAML 部署 / Helm 部署、Namespace 强制映射选择、审批流程、滚动更新、回滚、Pod 健康检查 |
| **配置中心** | 多配置条目（env/configmap/secret/serviceaccount）、发布/回滚版本化、部署时自动匹配注入 |
| **路由中心** | K8s Service / Ingress / Traefik IngressRoute / APISIX ApisixRoute 结构化管理与部署 |
| **插件中心** | Traefik Middleware / APISIX Plugin 等 CRD 资源 YAML 管理与部署 |
| **通知中心** | 飞书/钉钉/Slack Webhook 推送、全局/服务级规则、发送记录 |
| **用户管理** | 本地账号、RBAC（admin/owner/developer/viewer）、组权限 |
| **系统设置** | 集群管理、Git 仓库、镜像仓库、通知渠道、运行时配置 |

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.23+ · Gin · GORM · PostgreSQL · Redis · client-go |
| 前端 | Next.js 14 · React · Tailwind CSS · shadcn/ui · TanStack Query |
| 构建 | Kaniko（K8s Job 集群内构建） |
| 部署 | K8s Deployment/StatefulSet · Helm · Dynamic Client (CRD) |
| 通知 | Webhook (飞书/钉钉/Slack/自定义) |
| 认证 | JWT · BCrypt · AES-256-GCM |

## 快速开始

### 前置条件

- Go 1.23+
- Node.js 20+ / pnpm
- PostgreSQL 14+
- Redis 6+
- K8s 集群 + kubeconfig（用于构建和部署）

### 后端

```bash
cd backend

# 复制并编辑环境变量
cp .env.example .env
# 编辑 .env，填入数据库连接串、JWT 密钥、AES 密钥等

# 安装依赖
go mod download

# 运行数据库迁移
make migrate-up
# 或手动: migrate -database "${DATABASE_URL}" -path migrations up

# 编译运行
make build
./bin/server
```

### 前端

```bash
cd frontend

# 安装依赖
pnpm install

# 开发模式
pnpm dev

# 生产构建
pnpm build
pnpm start
```

### 初始配置

1. 访问 `http://localhost:3000`，注册第一个用户
2. 手动将用户 role 更新为 admin：`UPDATE users SET role='admin' WHERE id=1;`
3. 在系统设置中添加：
   - **集群**：填入 kubeconfig（用于构建和部署）
   - **集群 → 命名空间映射**：每个集群点击「命名空间映射」，登记本集群允许发布的 namespace（必填项；发布弹窗只允许选择此处登记过的 namespace）
   - **Git 仓库**：GitHub/GitLab 仓库 + Token
   - **镜像仓库**：Docker Hub / ECR / ACR 等
4. 在系统配置中设置：
   - `helm_job_namespace`：Helm Runner Job 运行的命名空间
   - `env_values_map`：集群环境到 Helm values 文件后缀的映射

## 项目结构

```
deployhub/
├── backend/                 # Go 后端
│   ├── cmd/server/          # 入口 + 依赖注入
│   ├── internal/
│   │   ├── config/          # 配置加载
│   │   ├── handler/         # HTTP Handler 层
│   │   ├── middleware/      # JWT/CORS/审计/RBAC 中间件
│   │   ├── model/           # GORM 模型
│   │   ├── pkg/             # 通用工具（响应/分页/掩码）
│   │   ├── repository/      # 数据访问层
│   │   ├── service/         # 业务逻辑层
│   │   │   ├── approval/    # 审批服务
│   │   │   ├── auth/        # 认证服务
│   │   │   ├── build/       # 构建服务 + Kaniko 执行器
│   │   │   ├── cluster/     # 集群管理 + ClientsetPool
│   │   │   ├── configcenter/# 配置中心服务
│   │   │   ├── crypto/      # AES-256-GCM 加密
│   │   │   ├── deploy/      # 部署服务 + 执行器
│   │   │   ├── notification/# 通知调度 + Webhook
│   │   │   ├── routing/     # 路由中心 + 插件中心
│   │   │   └── setting/     # 系统配置
│   │   └── ws/              # WebSocket Hub
│   └── migrations/          # 数据库迁移文件
├── frontend/                # Next.js 前端
│   └── src/
│       ├── app/(dashboard)/ # 页面路由
│       ├── components/      # UI 组件
│       ├── hooks/           # TanStack Query Hooks
│       ├── lib/             # API Client / WS Client
│       └── types/           # TypeScript 类型
└── specs/                   # 功能规格文档
```

## 架构

```
┌─────────────┐    ┌──────────────────────────────────────────┐
│   Frontend   │───▶│              Backend (Go)                │
│  Next.js 14  │◀───│  Handler → Service → Repository         │
└─────────────┘    ├──────────────────────────────────────────┤
                   │  PostgreSQL  │  Redis  │  K8s Clusters   │
                   └──────────────────────────────────────────┘
```

- **Handler 层**：HTTP 请求解析、参数校验、权限检查、响应序列化
- **Service 层**：业务逻辑、事务编排、跨实体协调
- **Repository 层**：GORM 数据访问，接口解耦

## 环境变量

| 变量 | 说明 | 必填 |
|------|------|------|
| DATABASE_URL | PostgreSQL 连接串 | ✅ |
| REDIS_URL | Redis 连接串 | ✅ |
| JWT_SECRET | JWT 签名密钥 | ✅ |
| AES_KEY | AES-256-GCM 密钥（64 位 hex） | ✅ |
| SERVER_PORT | 服务端口（默认 8080） | |
| LOG_LEVEL | 日志级别（debug/info/warn/error） | |
| S3_ENDPOINT | S3 兼容存储端点（头像上传） | |
| DEPLOYHUB_BASE_URL | 平台 URL（通知跳转链接） | |

## License

MIT
