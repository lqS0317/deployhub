# 004 - Helm 部署方式

## 1. Overview

为 DeployHub 新增 Helm 部署方式，与现有 client-go 直接部署（Direct）并存。用户在 Service 级别选择部署类型。Helm 部署通过在目标集群创建临时 Runner Job Pod（helm CLI + git）执行 `helm upgrade --install`，支持 Git 仓库中的统一 Chart 模板 + values 文件。

## 2. Motivation

| 现状 | 目标 |
|------|------|
| 只有 client-go 直接操作 K8s Deployment | 支持 Helm Chart 部署，复用团队已有 Chart 生态 |
| 无法编排复杂资源（Ingress/ConfigMap/Service 等） | Helm Chart 可包含任意 K8s 资源 |
| 无原生版本管理和回滚 | 利用 Helm release revision 原生回滚 |
| 每个服务部署配置分散 | 统一 Chart + 不同 values 文件标准化部署 |

## 3. User Stories

- **S1**: Admin 创建 Service 时选择 deploy_type=helm，配置 Chart Git 仓库路径和 values 文件路径
- **S2**: 用户发起 Helm 部署，指定镜像 tag → 系统在目标集群创建 Helm Runner Job → 执行 helm upgrade --install → 实时推送日志 → 记录 helm revision
- **S3**: 用户查看 Helm 部署的 Runner Pod 日志
- **S4**: 用户对 Helm 部署执行回滚，选择目标 revision → Runner Job 执行 helm rollback
- **S5**: Admin 在系统内编辑 Service 在不同集群的 values（YAML 编辑器）

## 4. Edge Cases

- deploy_type=direct 的服务走现有 executor 逻辑，完全不变
- Helm Runner Job 超时（10 分钟）→ 标记失败
- Git clone 失败（凭证错误/网络问题）→ init container 失败 → Job 失败
- helm upgrade 失败（Chart 错误/values 错误）→ main container 非零退出 → Job 失败
- 同一 release 并发部署 → 复用现有 FindActiveByService 锁
- 无 build_id 的纯 helm 部署 → image_tag 必须手动指定
- Service 切换 deploy_type → 允许但需确认（切换后旧方式的部署记录保留，不影响历史）

## 5. Functional Requirements

### 5.1 数据模型变更

**services 表新增字段**

| 字段 | 类型 | 约束 |
|------|------|------|
| deploy_type | varchar(10) | not null, default 'direct' |
| helm_repo_id | uint | nullable, FK → git_repos.id |
| helm_chart_path | varchar(255) | nullable |
| helm_values_path | varchar(255) | nullable |
| helm_release_name | varchar(100) | nullable |
| helm_chart_branch | varchar(100) | nullable, default 'main' |

**deployments 表新增字段**

| 字段 | 类型 | 约束 |
|------|------|------|
| helm_revision | int | nullable |

**新表 helm_values**

| 字段 | 类型 | 约束 |
|------|------|------|
| id | uint | PK |
| service_id | uint | FK → services.id, not null |
| cluster_id | uint | FK → clusters.id, not null |
| content | text | YAML 内容 |
| version | int | not null, default 1 |
| updated_by | uint | FK → users.id |
| created_at | timestamp | |
| updated_at | timestamp | |

唯一索引: `(service_id, cluster_id)`

### 5.2 Executor 接口抽象

```go
// DeployExecutor 部署执行器接口
type DeployExecutor interface {
    Execute(ctx context.Context, deployment *model.Deployment, service *model.Service) error
}
```

- **DirectExecutor**：重命名自现有 `executor.go`，实现 `DeployExecutor`
- **HelmExecutor**：新增，创建 Helm Runner Job Pod 并监听完成

### 5.3 Helm Runner Job 结构

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: helm-deploy-{deployment_id}-{timestamp}
  namespace: {target_namespace}
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 3600
  template:
    spec:
      serviceAccountName: deployhub-helm-runner
      restartPolicy: Never
      initContainers:
      - name: git-clone
        image: alpine/git:latest
        command: ["sh", "-c"]
        args: ["git clone --depth 1 -b {branch} https://{token}@{repo} /workspace"]
        volumeMounts:
        - name: workspace
          mountPath: /workspace
      containers:
      - name: helm
        image: alpine/helm:3.14
        command: ["sh", "-c"]
        args:
        - |
          helm upgrade --install {release} /workspace/{chart_path} \
            [-f /workspace/{values_path}] \
            [-f /tmp/system-values.yaml] \
            --set image.tag={image_tag} \
            --namespace {namespace} \
            --wait --timeout 10m
        volumeMounts:
        - name: workspace
          mountPath: /workspace
        - name: system-values
          mountPath: /tmp/system-values.yaml
          subPath: values.yaml
      volumes:
      - name: workspace
        emptyDir: {}
      - name: system-values
        configMap:
          name: helm-values-{deployment_id}
```

### 5.4 API 契约

#### Helm Values 管理

| Method | Path | Body | Response |
|--------|------|------|----------|
| GET | /api/v1/services/:id/helm-values | — | `{ items: HelmValues[] }` |
| PUT | /api/v1/services/:id/helm-values/:cluster_id | `{ content: "yaml..." }` | `HelmValues` |
| GET | /api/v1/services/:id/helm-values/:cluster_id/history | — | `{ items: HelmValuesVersion[] }` |

#### Helm 回滚

| Method | Path | Body | Response |
|--------|------|------|----------|
| POST | /api/v1/deployments/:id/helm-rollback | `{ revision: int }` | `201 Deployment` |

#### 现有 API 变更

- `POST /api/v1/services` — 接受新增 helm 字段
- `PUT /api/v1/services/:id` — 可更新 helm 字段
- `POST /api/v1/deployments` — 创建部署后根据 deploy_type 路由到对应 executor
- `POST /api/v1/deployments/:id/rollback` — direct 类型走现有逻辑，helm 类型走 helm rollback

### 5.5 Deploy Handler 激活 Executor + Watcher

**当前问题**：`Executor` 和 `RolloutWatcher` 虽然注入了 DeployHandler，但 handler 中没有调用它们。需要：

1. 审批通过后调用 `StartDeploy` + `executor.Execute` + `watcher.Watch`
2. 或者在 DeployHandler.Create 中，直接调用 executor（如果不走审批流）

**本次方案**：
- `CreateDeployment` 创建后状态仍为 `pending_approval`
- 新增 `POST /api/v1/deployments/:id/approve` — 审批通过后触发执行
- approve handler 中：`StartDeploy` → 根据 deploy_type 选择 executor → `executor.Execute` → `watcher.Watch`（direct）或 Job Watch（helm）

### 5.6 前端变更

**创建/编辑服务表单**
- 新增「部署类型」Radio：直接部署 / Helm
- 选择 Helm 后展开 Helm 配置区域：Chart Git 仓库选择、Chart 路径、Values 路径、Release 名称、分支

**服务详情页**
- Helm 类型显示 Chart 配置信息
- 新增 Helm Values 编辑 Tab（YAML 文本域，按集群切换）

**部署列表/详情**
- Helm 类型显示 helm_revision 列
- 回滚操作：Helm 类型显示 revision 输入 → 原生回滚

### 5.7 安全约束

- Git 凭证通过临时 K8s Secret 注入 init container，Job 完成后清理
- 系统 values 通过 ConfigMap 挂载，Job 完成后清理
- Runner Pod 使用 ServiceAccount in-cluster 认证
- Runner Job 超时 10 分钟
- Job/Pod/临时资源完成后自动清理（TTL 1 小时）

## 6. Success Criteria

- [ ] Service 支持 deploy_type 选择（direct/helm）
- [ ] Direct 部署逻辑完全不变
- [ ] Helm 部署通过 Runner Job 执行 helm upgrade --install
- [ ] Helm 回滚通过 Runner Job 执行 helm rollback
- [ ] Runner Job 日志实时推送到前端
- [ ] Helm Values 系统内编辑和版本管理
- [ ] Executor 接口抽象，deploy_type 路由
- [ ] 部署审批通过后自动触发执行
- [ ] 全部后端测试通过

## 7. Assumptions

- 目标集群各 namespace 已预创建 `deployhub-helm-runner` ServiceAccount + ClusterRoleBinding
- 目标集群可拉取 `alpine/git` 和 `alpine/helm:3.14` 镜像
- Chart Git 仓库使用 HTTPS + token 认证
- 现有审批流程（ApprovalService）可用但未集成，本次可选集成

## 8. Non-Goals

- Helm Chart OCI Registry / ChartMuseum
- Helm Chart 版本管理（使用 Git 分支/Tag）
- helm template --dry-run 预览
- Helm hooks 管理
- Helmfile / Kustomize 集成
- 跨集群部署（Runner Pod 在目标集群运行）
- Helm values 中的敏感值 Secret 化（本次用 ConfigMap，后续可增强）
