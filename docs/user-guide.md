# DeployHub 使用指南

## 初始设置

### 1. 系统设置

登录后进入「系统设置」：

**集群管理**
- 添加 K8s 集群：名称、环境（devnet/qanet/testnet/mainnet）、kubeconfig
- 点击「测试连接」验证集群可达性
- 设置 Helm ServiceAccount（Helm 部署 Runner Job 使用的 SA）
- 点击「命名空间映射」打开映射管理：
  - 手动登记可发布的 namespace（可勾选「设为默认」），或点「从集群同步」一次性导入集群中已存在的 namespace
  - 发布流程只允许选择此处登记过的 namespace，未登记的集群无法发起部署
  - 支持单条删除；删除后已存在的部署不受影响，但回滚和新发布会被后端拒绝

**Git 仓库**
- 添加 GitHub/GitLab 仓库：URL + Token
- 点击「测试连接」验证凭证有效

**镜像仓库**
- 添加 Docker Hub / ECR / ACR 等：URL + 用户名密码

**系统配置**（运行时配置，修改后立即生效）
- `helm_job_namespace`：Helm Runner Job 运行的命名空间
- `env_values_map`：集群环境到 Helm values 文件后缀映射（如 `qanet:qa,testnet:testnet`）

---

## 服务管理

### 创建服务
1. 进入「服务管理」→ 「创建服务」
2. 填写：服务名称、部署类型（直接/Helm/其他）、Git 仓库、默认分支
3. 直接部署类型：可配置运行时默认参数（端口、副本、资源限制、启动命令、健康检查）
4. 这些默认参数在每次部署时自动继承

### 服务详情
- 成员管理：添加 owner/developer/viewer 角色
- 基本信息编辑

---

## 构建中心

### 触发构建
1. 进入「构建中心」→ 「触发构建」
2. 选择服务 → 分支 → Commit
3. 配置构建参数：构建集群、镜像仓库、镜像路径、Dockerfile 路径
4. 可自定义镜像标签（留空自动生成）

### 构建日志
- 点击「日志」实时查看 Kaniko 构建输出
- 构建失败会显示详细错误信息
- 支持「重建」和「取消」操作

---

## 发布管理

### 发起部署
1. 进入「发布管理」→ 「发起部署」
2. 选择部署方式：
   - **配置部署**：从服务运行时参数和配置中心自动读取配置，选择服务 → 集群 → 命名空间（下拉，仅展示已登记映射）→ 镜像 → 确认
   - **YAML 部署**：粘贴完整 K8s YAML，支持多资源（--- 分隔）
   - **Helm 部署**：选择 Helm Chart Git 仓库、Chart 路径、分支、镜像来源
3. 命名空间下拉规则：
   - 数据来源为「系统设置 → 集群 → 命名空间映射」，不会从集群实时拉取
   - 默认选中标记为「默认」的项；无默认时取列表首项
   - 切换集群会重置 namespace 选项并自动选默认
   - 集群没有任何映射时，提交按钮禁用并提示「该集群未配置 namespace 映射，请先在集群管理中配置」

### 部署流程
1. 创建后自动执行预览（dry-run）
2. Admin 自动审批，非 Admin 需要管理员审批
3. 审批通过后执行部署
4. 部署完成后进入 Pod 健康检查（60 秒观察期）
5. 观察期无异常 → pod_healthy；发现问题 → pod_unhealthy（含诊断信息）

### 部署详情
- 状态时间线（6 阶段）
- Helm 执行命令预览
- 部署事件日志（WebSocket 实时推送）
- Pod 容器日志
- 失败原因/Pod 异常诊断信息

---

## 配置中心

### 概念
- 每个服务的每个环境（集群）下可创建多个**配置条目**
- 配置条目类型：env（环境变量）、configmap、secret、serviceaccount
- 部署时自动匹配：查找该服务+目标集群下所有已发布的配置条目，生成对应 K8s 资源

### 配置条目管理
1. 进入「配置中心」→ 选择服务 → 选择环境（左侧集群列表）
2. 创建配置条目：名称 + 类型 + 格式（properties/yaml/json）
3. 编辑配置项（KV 表格或代码编辑器）
4. 点击「发布」生成版本快照
5. 支持「回滚」到历史版本

### 配置类型说明
| 类型 | 部署行为 | 说明 |
|------|----------|------|
| env | Pod envFrom | 注入为环境变量 |
| configmap | ConfigMap + volumeMount | 挂载到指定路径 |
| secret | Secret + volumeMount | 加密存储，挂载到指定路径 |
| serviceaccount | K8s ServiceAccount | KV 作为 SA annotations |

---

## 路由中心

### 支持的资源类型
- **K8s Service**：type/selector/ports
- **Ingress**：className/tls/rules/annotations
- **Traefik IngressRoute**：entryPoints/routes/middlewares/tls
- **APISIX ApisixRoute**：http rules/backends/plugins/timeout

### 使用流程
1. 进入「路由中心」→ 选择资源类型 Tab
2. 「新建」→ 填写结构化表单
3. 保存后点击「部署」→ 选择目标集群 → 命名空间（下拉，仅展示该集群已登记的映射）→ 预览 YAML → 确认
4. 路由是手动部署模式，只有首次上线和内容更新后需要部署

> 命名空间下拉规则与发布管理一致：未配置映射的集群禁止部署，需要先到「系统设置 → 集群 → 命名空间映射」登记。

---

## 插件中心

### 用途
管理 Traefik Middleware、APISIX Plugin 等 K8s CRD 资源。

### 使用流程
1. 进入「插件中心」→ 「新建插件」
2. 填写名称、描述、YAML 内容
3. 部署到指定集群 → 选择命名空间（仅展示该集群已登记的映射）
4. 创建路由时可引用插件名称（IngressRoute 的 middlewares、ApisixRoute 的 plugins）

---

## 通知中心

### 设置通知渠道
系统设置 → 通知渠道 → 添加飞书/钉钉/Slack/自定义 Webhook

### 配置通知规则
通知中心 → 通知规则 Tab：
- **全局 All**：默认渠道，所有事件走此渠道
- **按事件自定义**：特定事件走特定渠道（覆盖 All）
- **服务级绑定**：为服务指定专用渠道（覆盖全局）

### 支持的事件
构建成功/失败/取消、部署成功/失败/取消、待审批、Pod 异常、回滚触发

---

## 权限说明

### 全局角色
| 角色 | 权限 |
|------|------|
| admin | 全部功能 |
| user | 基本功能（需要服务权限） |

### 服务角色
| 角色 | 权限 |
|------|------|
| owner | 服务全部操作 + 成员管理 |
| developer | 构建 + 部署 + 配置编辑 |
| viewer | 只读查看 |

### 配置中心权限
按「服务 + 环境」粒度，viewer/editor/publisher 三级。
