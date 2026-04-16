# DeployHub 本地开发快速入门

面向开发者的本地环境搭建说明，帮助你在本机运行 DeployHub 前后端与依赖服务。

**技术栈概览：** 后端 Go 1.22+（Gin、GORM、client-go、gorilla/websocket）；前端 Node.js 20+、Next.js 14、Tailwind CSS、shadcn/ui；数据层 PostgreSQL 15+、Redis 7+；迁移工具 golang-migrate。生产构建可使用 Kaniko，本地开发无需安装。

---

## 前置条件

- **Go** 1.22 及以上
- **Node.js** 20 及以上（包管理使用 **pnpm**）
- **Docker** + **Docker Compose**（用于本地 PostgreSQL 与 Redis）
- **Make**（可选，用于执行 Makefile 中的快捷命令）

---

## 快速启动

### 1. 克隆仓库

```bash
git clone <repo-url> && cd deployhub
```

将 `<repo-url>` 替换为实际仓库地址。

### 2. 启动基础服务

在项目根目录执行：

```bash
docker-compose up -d
```

将启动 PostgreSQL 与 Redis。建议在仓库根目录提供 `docker-compose.yml`，内容示例：

```yaml
services:
  postgres:
    image: postgres:15-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: deployhub
      POSTGRES_PASSWORD: deployhub
      POSTGRES_DB: deployhub

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

### 3. 环境配置

```bash
cp backend/.env.example backend/.env
```

按实际环境编辑 `backend/.env`。`backend/.env.example` 建议包含以下变量：

```env
# 数据库（与 docker-compose 中用户/库名一致）
DATABASE_URL=postgres://deployhub:deployhub@localhost:5432/deployhub?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379/0

# 鉴权与加密（生产环境务必更换为强随机值）
JWT_SECRET=your-jwt-secret-change-in-production
# AES-256-GCM：32 字节密钥，以下为 64 位十六进制占位示例（请自行生成并保密）
AES_KEY=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef

# OAuth（可选，未配置时可跳过 GitHub 登录相关功能）
OAUTH_GITHUB_CLIENT_ID=
OAUTH_GITHUB_CLIENT_SECRET=

# 服务
SERVER_PORT=8080
LOG_LEVEL=debug
```

### 4. 数据库迁移

```bash
cd backend
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate -path migrations -database "$DATABASE_URL" up
```

确保 shell 中已导出与 `.env` 一致的 `DATABASE_URL`，或先 `export $(grep -v '^#' .env | xargs)` 再执行迁移（按团队习惯选择一种方式即可）。

### 5. 启动后端

```bash
cd backend
go run cmd/server/main.go
```

API 默认监听 **http://localhost:8080**。

### 6. 启动前端

```bash
cd frontend
pnpm install
pnpm dev
```

前端开发服务器默认 **http://localhost:3000**。

---

## 常用命令

若项目根目录提供 `Makefile`，常见目标如下：

| 命令 | 说明 |
|------|------|
| `make dev` | 启动所有服务（docker-compose + 后端 + 前端） |
| `make migrate-up` | 执行数据库迁移（向上） |
| `make migrate-down` | 回滚最后一次迁移 |
| `make migrate-create NAME=xxx` | 创建名为 `xxx` 的新迁移文件 |
| `make test-backend` | 运行后端测试 |
| `make test-frontend` | 运行前端测试 |
| `make lint` | 代码检查 |

---

## 验证

按顺序确认：

1. 浏览器打开 **http://localhost:3000**，应看到登录页。
2. 访问 **/register** 注册新账号。
3. 登录后进入控制台（dashboard）页面。
4. 访问 **http://localhost:8080/api/v1/health**，确认 API 健康检查返回正常。

全部通过即表示本地环境已就绪。
