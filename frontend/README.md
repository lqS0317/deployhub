# DeployHub Frontend

Next.js 14 前端应用，提供 DeployHub 的管理界面。

## 技术栈

- **Next.js 14** + React 18
- **Tailwind CSS** + **shadcn/ui** 组件
- **TanStack Query** 服务端状态管理
- **Axios** HTTP 客户端
- **WebSocket** 实时日志和部署进度

## 开发

```bash
# 安装依赖
pnpm install

# 开发模式（http://localhost:3000）
pnpm dev

# 生产构建
pnpm build
pnpm start
```

## 环境变量

创建 `.env.local`：

```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## 页面结构

| 路由 | 页面 | 功能 |
|------|------|------|
| `/login` | 登录页 | 用户登录/注册 |
| `/services` | 服务管理 | 服务列表、创建/编辑服务 |
| `/services/[id]` | 服务详情 | 成员管理、基本信息 |
| `/builds` | 构建中心 | 构建列表、触发构建、实时日志 |
| `/deployments` | 发布管理 | 部署列表、发起部署、部署详情 |
| `/deployments/[id]` | 部署详情 | 状态时间线、命令预览、Pod 日志 |
| `/configs` | 配置中心 | 服务选择 → 配置管理 |
| `/configs/[serviceId]` | 服务配置 | 环境切换、配置条目、发布/回滚 |
| `/routes` | 路由中心 | Service/Ingress/IngressRoute/ApisixRoute |
| `/plugins` | 插件中心 | Middleware/Plugin YAML 管理 |
| `/approvals` | 审批中心 | 待审批/已处理列表 |
| `/notifications` | 通知中心 | 站内通知、通知规则、发送记录 |
| `/settings/*` | 系统设置 | 集群/Git/Registry/用户/组/通知渠道/系统配置 |
| `/profile` | 个人中心 | 资料编辑、修改密码 |

## 目录结构

```
frontend/src/
├── app/
│   ├── (dashboard)/          # 需认证的页面
│   │   ├── services/         # 服务管理
│   │   ├── builds/           # 构建中心
│   │   ├── deployments/      # 发布管理
│   │   ├── configs/          # 配置中心
│   │   ├── routes/           # 路由中心
│   │   ├── plugins/          # 插件中心
│   │   ├── approvals/        # 审批中心
│   │   ├── notifications/    # 通知中心
│   │   ├── settings/         # 系统设置
│   │   └── profile/          # 个人中心
│   ├── login/                # 登录页
│   ├── layout.tsx            # 根布局
│   └── providers.tsx         # QueryClient + Toast
├── components/
│   ├── build/                # 构建相关组件
│   ├── config/               # 配置中心组件
│   ├── deploy/               # 部署相关组件
│   ├── plugin/               # 插件中心组件
│   ├── route/                # 路由中心组件
│   ├── service/              # 服务管理组件
│   ├── ui/                   # 通用 UI（Toast）
│   └── layout/               # 布局组件（Sidebar）
├── hooks/                    # TanStack Query Hooks
│   ├── use-auth.ts           # 认证
│   ├── use-services.ts       # 服务
│   ├── use-builds.ts         # 构建
│   ├── use-deployments.ts    # 部署
│   ├── use-config-center.ts  # 配置中心
│   ├── use-route-entries.ts  # 路由中心
│   ├── use-route-plugins.ts  # 插件中心
│   └── ...
├── lib/
│   ├── api-client.ts         # Axios 实例 + 拦截器
│   └── ws-client.ts          # WebSocket 客户端
└── types/
    └── index.ts              # TypeScript 类型定义
```

## 组件设计

- **结构化表单**：路由中心的 4 种资源类型各有专用表单组件（service-form、ingress-form、ingressroute-form、apisixroute-form）
- **实时预览**：部署对话框展示已发布配置条目、YAML 预览
- **WebSocket 集成**：构建日志、部署进度、Pod 日志均通过 WS 实时推送
- **响应式布局**：配置中心左右分栏、路由中心 Tab 切换
