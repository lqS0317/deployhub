# Implementation Plan: DeployHub — K8s 原生运维发布平台

**Branch**: `001-deployhub-platform` | **Date**: 2026-04-03 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-deployhub-platform/spec.md`

## Summary

构建一个面向小团队（5-20 人）的 Kubernetes 原生运维发布平台，管理 <50 个微服务的完整生命周期。Go 后端（Gin + GORM + client-go）提供 REST API 和 WebSocket 实时推送，Next.js 14 前端提供 7 个功能页面。核心能力包括：Kaniko 集群内镜像构建、client-go 滚动更新发布、审批工作流、Go template 配置管理、全链路审计。

## Technical Context

**Language/Version**: Go 1.22+ (backend), TypeScript / Node.js 20+ (frontend)  
**Primary Dependencies**: Gin (HTTP), GORM (ORM), client-go (K8s), gorilla/websocket (WS), golang-migrate (migration) | Next.js 14, React 18, Tailwind CSS, shadcn/ui, TanStack Query  
**Storage**: PostgreSQL 15+ (business data), Redis 7+ (session/cache)  
**Testing**: Go `testing` + `testify` + `httptest` (backend TDD), Jest + React Testing Library (frontend)  
**Target Platform**: Linux server (K8s deployment), modern browsers (Chrome/Firefox/Edge)  
**Project Type**: Web application (Go API + Next.js SPA)  
**Performance Goals**: 构建日志延迟 <2s, 发布进度延迟 <3s, 全流程 <5min  
**Constraints**: <50 services, <5 clusters, 5-20 concurrent users, AES-256-GCM encryption for all sensitive data  
**Scale/Scope**: 14 entities, 7 frontend pages, 5 core workflows, ~20 API resource groups

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principle | Status | Verification |
|---|-----------|--------|-------------|
| I | 安全第一 | ✅ PASS | AES-256-GCM 加密所有凭证, JWT+OAuth2 认证, BCrypt 密码, RBAC 中间件, 日志脱敏 |
| II | 测试驱动开发 | ✅ PASS | TDD 强制: Go testing+testify+httptest, Jest+RTL, 核心引擎单测+API 集成测试 |
| III | 分层架构 | ✅ PASS | handler → service → repository 三层, 接口解耦, 禁止跨层调用 |
| IV | 多集群隔离 | ✅ PASS | 每集群独立 clientset, kubeconfig 加密存储, Config-as-Code 模板渲染 |
| V | 审批驱动发布 | ✅ PASS | Owner 跳过/Developer 需审批/Admin 紧急通道, Rolling Update only |
| VI | 实时可观测 | ✅ PASS | gorilla/websocket 推送构建日志和发布进度, client-go Watch, 结构化日志 |
| VII | 简洁优先 | ✅ PASS | MVP 切片交付, <50 服务, 不做多租户, jsonb 扩展字段 |

**Non-Goals check**: 无 CI 编排 ✅ | 无监控告警 ✅ | 无金丝雀/蓝绿 ✅ | 无多租户 ✅ | 无外部配置中心 ✅

## Project Structure

### Documentation (this feature)

```text
specs/001-deployhub-platform/
├── plan.md              # This file
├── research.md          # Phase 0: technology decisions
├── data-model.md        # Phase 1: 14 entities with fields and relationships
├── quickstart.md        # Phase 1: dev environment setup guide
├── contracts/           # Phase 1: REST API contracts
│   └── api.md           # All API endpoints grouped by module
└── tasks.md             # Phase 2: implementation tasks (/speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── server/
│       └── main.go                  # 应用入口
├── internal/
│   ├── config/                      # 配置加载（环境变量）
│   ├── middleware/                   # JWT 认证、RBAC、审计日志、请求日志
│   ├── model/                       # GORM 模型（14 个实体）
│   ├── repository/                  # 数据访问层（每实体一个 repo）
│   ├── service/                     # 业务逻辑层
│   │   ├── auth/                    # 认证服务（JWT + OAuth2 + 本地）
│   │   ├── cluster/                 # 集群管理（client-go clientset 池）
│   │   ├── gitrepo/                 # Git 仓库管理
│   │   ├── registry/                # 镜像仓库管理
│   │   ├── svc/                     # 服务管理 + 成员 RBAC
│   │   ├── build/                   # 构建引擎（Kaniko Job）
│   │   ├── deploy/                  # 发布引擎（Rolling Update）
│   │   ├── approval/                # 审批引擎
│   │   ├── config/                  # 配置中心（模板渲染 + 版本管理）
│   │   ├── notification/            # 通知引擎（站内 + webhook）
│   │   ├── audit/                   # 审计日志
│   │   └── crypto/                  # AES-256-GCM 加密/解密
│   ├── handler/                     # HTTP handler 层（Gin routes）
│   │   ├── auth.go
│   │   ├── cluster.go
│   │   ├── gitrepo.go
│   │   ├── registry.go
│   │   ├── service.go
│   │   ├── build.go
│   │   ├── deploy.go
│   │   ├── approval.go
│   │   ├── config.go
│   │   ├── notification.go
│   │   ├── audit.go
│   │   └── ws.go                    # WebSocket handler（日志流 + 发布进度）
│   ├── ws/                          # WebSocket hub 和连接管理
│   └── pkg/                         # 共享工具（分页、响应格式、错误码）
├── migrations/                      # golang-migrate SQL 文件
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   └── ...
├── go.mod
├── go.sum
├── Dockerfile
└── Makefile

frontend/
├── src/
│   ├── app/                         # Next.js 14 App Router
│   │   ├── (auth)/                  # 登录/注册页面
│   │   ├── (dashboard)/             # 主布局（侧边栏）
│   │   │   ├── services/            # 服务管理
│   │   │   ├── builds/              # 构建中心
│   │   │   ├── deployments/         # 发布管理
│   │   │   ├── configs/             # 配置中心
│   │   │   ├── approvals/           # 审批中心
│   │   │   ├── notifications/       # 通知中心
│   │   │   └── settings/            # 系统设置
│   │   └── layout.tsx
│   ├── components/                  # shadcn/ui 组件 + 业务组件
│   │   ├── ui/                      # shadcn/ui 基础组件
│   │   ├── service/                 # 服务相关组件
│   │   ├── build/                   # 构建相关组件（日志查看器）
│   │   ├── deploy/                  # 发布相关组件（进度条、Pod 状态）
│   │   ├── config/                  # 配置相关组件（模板编辑器、diff 视图）
│   │   └── layout/                  # 布局组件（侧边栏、顶栏）
│   ├── lib/                         # 工具函数、API client、WebSocket client
│   ├── hooks/                       # TanStack Query hooks
│   └── types/                       # TypeScript 类型定义
├── public/
├── next.config.js
├── tailwind.config.ts
├── package.json
└── Dockerfile

docker-compose.yml                   # 本地开发：PostgreSQL + Redis
```

**Structure Decision**: Web application 模式（backend/ + frontend/），Go 后端使用 `internal/` 保持包私有，三层架构对应 `handler/` → `service/` → `repository/`。

## Complexity Tracking

> No constitution violations. All design decisions align with the 7 core principles.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (none)    | —          | —                                   |
