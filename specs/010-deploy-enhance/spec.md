# 010 - 发布管理增强：编辑/删除 + 失败原因 + 部署日志

## 1. Overview

增强发布管理的三项核心能力：部署记录编辑/删除、失败原因展示、实时部署事件和 Pod 容器日志查看。

## 2. Data Model Changes

### Deployment 表新增字段

| 字段 | 类型 | 说明 |
|------|------|------|
| fail_reason | text | nullable, 失败原因 |

## 3. API Changes

### 新增 API

| Method | Path | 说明 |
|--------|------|------|
| PUT | /deployments/:id | 编辑部署（仅 pending_approval 状态） |
| DELETE | /deployments/:id | 删除部署（级联删除审批记录） |
| GET | /deployments/:id/pods | Pod 列表（名称、状态、容器列表） |
| WS | /deployments/:id/pod-logs | Pod 容器日志流 |

## 4. Edit Rules

- 允许编辑状态：仅 `pending_approval`
- 可编辑字段：cluster_id, namespace, image_tag, external_image, replicas, deploy_type, workload_type, port, 资源配置, Helm 参数
- 不可编辑：service_id, trigger_user_id
- 编辑后：清空 preview_yaml/preview_summary，状态重置为 `previewing`，自动重新触发预览

## 5. Delete Rules

- 允许删除：任何终态（success/failed/cancelled/rejected）+ pending_approval/approved/previewed
- 不允许直接删除：deploying/previewing（需先取消）
- 级联删除：关联的 Approval 记录

## 6. Fail Reason

所有将状态设为 failed 的调用方同时写入 fail_reason：
- `RolloutWatcher.markFailed` — 超时/集群连接失败
- `DeployHandler.autoPreview` — 预览失败原因
- `HelmExecutor` — Helm 命令执行失败
- `BuildExecutor` — 构建失败原因（已有）

## 7. Pod Logs

### HTTP: GET /deployments/:id/pods

```json
{
  "items": [
    {
      "name": "myapp-abc123",
      "status": "Running",
      "ready": true,
      "containers": ["myapp", "sidecar"],
      "created_at": "2026-04-10T15:00:00Z"
    }
  ]
}
```

### WS: /deployments/:id/pod-logs

Query params: `pod`, `container`, `tail` (default 100)

校验逻辑：
- Pod 不存在 → error message
- Container Waiting (CrashLoopBackOff 等) → error with reason
- Container Running/Terminated → Follow 流式推送

## 8. Frontend Changes

- 部署列表：`pending_approval` 状态增加「编辑」按钮；所有状态增加「删除」按钮；`failed` 行 hover 显示失败原因
- 部署详情：失败状态显示红色告警卡片（fail_reason）；新增「事件日志」和「容器日志」Tab
- 日志面板：终端风格，自动滚动，Pod/Container 下拉选择

## 9. Success Criteria

- [ ] 编辑部署后自动重新预览
- [ ] 删除部署级联清理审批记录
- [ ] 失败原因在详情页和列表页可见
- [ ] 实时部署事件日志可查看
- [ ] Pod 容器日志可实时查看
- [ ] 全部测试通过

## 10. Non-Goals

- 日志持久化
- 日志搜索/过滤
- 批量删除
- 日志回看
