# 007 - Service 与集群解耦 + Dry-Run 预览

## 1. Overview

将 Service 从"绑定特定集群+命名空间"改为"纯服务定义"，部署目标（cluster_id + namespace）移至 Deployment 创建时指定。同时新增发布前 dry-run 预览功能，扩展部署状态机。

## 2. Motivation

| 现状 | 目标 |
|------|------|
| Service 绑定 cluster_id + namespace | Service 只定义代码/镜像/Chart，不绑定集群 |
| 同一服务只能部署到一个集群 | 一个 Service 可部署到多个集群/环境 |
| 部署直接执行，无预览 | dry-run 预览后确认再执行 |

## 3. Impact Analysis

### 3.1 必须修改（Group A：Service → Deployment 迁移）

**后端（17 个文件）：**

| 文件 | 变更 |
|------|------|
| `model/service.go` | ClusterID → nullable（渐进式），Namespace → nullable |
| `migrations/000027` | ALTER services SET cluster_id nullable, namespace nullable |
| `repository/service_repo.go` + `_impl.go` | List 移除 clusterID/namespace 参数 |
| `service/svc/svc.go` | List 移除 clusterID/namespace 参数 |
| `handler/service.go` | createServiceRequest 移除 cluster_id/namespace 必填；List 移除筛选 |
| `service/deploy/deploy.go` | CreateDeployment 接受 clusterID/namespace 参数（不再从 svc 取）|
| `service/deploy/deploy.go` | Rollback 从原部署取 clusterID/namespace（不变） |
| `handler/deploy.go` | createDeploymentRequest 新增 cluster_id + namespace 必填 |
| `service/build/build.go` | buildClusterID=0 时不再 fallback 到 svc.ClusterID |
| `service/svc/import_yaml.go` | 导入时 namespace 变为可选信息 |

**前端（6 个文件）：**

| 文件 | 变更 |
|------|------|
| `types/index.ts` | Service.cluster_id/namespace 改为 optional |
| `service-create-dialog.tsx` | 移除 cluster_id/namespace 必填 |
| `services/page.tsx` | 列表移除集群/命名空间列 |
| `services/[id]/page.tsx` | 详情移除集群显示 |
| `use-services.ts` | List 移除 cluster_id/namespace 查询参数 |
| deploy 对话框 | 新增 cluster_id + namespace 输入 |

### 3.2 保持不变（Group C）

- DirectExecutor/HelmExecutor：使用 `deployment.ClusterID`/`deployment.Namespace`（已解耦）
- RolloutWatcher：同上
- HelmValues：service_id + cluster_id 索引保留（一个服务在不同集群有不同 values 是合理的）
- Build executor：使用 `build.BuildClusterID`（独立的构建集群概念）
- Config 模块：独立的 cluster_id 引用

## 4. Data Model Changes

### 4.1 Service 表变更

```sql
ALTER TABLE services ALTER COLUMN cluster_id DROP NOT NULL;
ALTER TABLE services ALTER COLUMN namespace DROP NOT NULL;
ALTER TABLE services ALTER COLUMN namespace SET DEFAULT '';
```

### 4.2 Deployment 表新增字段

| 字段 | 类型 | 说明 |
|------|------|------|
| preview_yaml | text | nullable, dry-run 结果的完整 YAML |
| preview_summary | jsonb | nullable, 结构化预览摘要 |

### 4.3 新增部署状态

```go
DeployStatusPreviewing = "previewing"   // dry-run 执行中
DeployStatusPreviewed  = "previewed"    // 预览完成等待确认
DeployStatusCancelled  = "cancelled"    // 取消
```

### 4.4 状态流

```
创建 → pending_approval → approved → previewing → previewed → deploying → success/failed
                                                    ↓
                                                 cancelled
```

## 5. API Changes

### 5.1 修改 API

| Method | Path | 变更 |
|--------|------|------|
| POST | /services | 移除 cluster_id/namespace 必填 |
| PUT | /services/:id | 同上 |
| POST | /deployments | 新增 cluster_id + namespace 必填 |
| GET | /services | 移除 cluster_id/namespace 筛选参数 |

### 5.2 新增 API

| Method | Path | 说明 |
|--------|------|------|
| POST | /deployments/:id/preview | 触发 dry-run 预览 |
| GET | /deployments/:id/preview | 获取预览结果 |
| POST | /deployments/:id/cancel | 取消部署 |

### 5.3 预览结果结构

```json
{
  "preview_yaml": "apiVersion: apps/v1\nkind: Deployment...",
  "summary": {
    "resources": [
      { "kind": "Deployment", "name": "myapp", "action": "create" },
      { "kind": "Service", "name": "myapp", "action": "unchanged" }
    ],
    "changes": {
      "image": "myapp:v1.0.0 → myapp:v2.0.0",
      "replicas": "3 → 5"
    }
  }
}
```

## 6. Dry-Run Implementation

### 6.1 Direct Dry-Run

```go
// 使用 K8s server-side dry-run
_, err := client.Create(ctx, k8sDep, metav1.CreateOptions{
    DryRun: []string{metav1.DryRunAll},
})
```

将构建好的资源 spec 序列化为 YAML 作为 preview_yaml。

### 6.2 Helm Dry-Run

Runner Job 执行 `helm template`（不连接集群，纯本地渲染）：

```bash
helm template {release} /workspace/{chart_path} \
  -f /workspace/services/{name}/app.yaml \
  --set image.tag={tag} \
  --namespace {namespace}
```

输出作为 preview_yaml。

## 7. Frontend Changes

- **创建部署对话框**：新增 cluster_id + namespace 下拉/输入
- **部署详情页**：previewed 状态显示预览面板（资源列表 + YAML 展开）+ 确认/取消按钮
- **创建/编辑服务**：cluster_id/namespace 变为可选（或移除）
- **服务列表**：移除集群/命名空间列

## 8. Success Criteria

- [ ] Service 不再强制绑定 cluster_id/namespace
- [ ] 创建部署时由用户指定 cluster_id + namespace
- [ ] 同一服务可部署到不同集群
- [ ] Direct dry-run 返回预览 YAML
- [ ] Helm dry-run 返回 helm template 输出
- [ ] previewed 状态可确认执行或取消
- [ ] 现有 build/config/helm-values 功能不受影响
- [ ] 全部测试通过

## 9. Non-Goals

- 两次预览之间的 diff
- dry-run 缓存
- 修改构建流程
- 服务 → 集群默认关联
- namespace 自动发现

## 10. Migration Strategy

1. **Phase A**：数据库迁移 — cluster_id/namespace nullable
2. **Phase B**：后端代码 — 创建/编辑服务不再要求 cluster_id，部署接受 cluster_id
3. **Phase C**：前端 — 调整表单和列表
4. **Phase D**：清理 — 后续版本可删除 Service 上的 cluster_id/namespace 字段
