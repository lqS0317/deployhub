# 005 - Direct 部署模式 StatefulSet 支持

## 1. Overview

为 Direct 部署模式新增 StatefulSet 工作负载类型支持。Service 新增 `workload_type` 字段，DirectExecutor 和 RolloutWatcher 根据该字段操作对应的 K8s 资源（Deployment 或 StatefulSet）。Helm 部署模式不受影响。

## 2. Motivation

| 现状 | 目标 |
|------|------|
| DirectExecutor 硬编码操作 Deployment | 支持 Deployment 和 StatefulSet 两种工作负载 |
| 有状态服务无法纳入统一发布 | StatefulSet 类服务（DB、MQ、Cache）可统一管理 |

## 3. Data Model

**services 表新增字段：**

| 字段 | 类型 | 约束 |
|------|------|------|
| workload_type | varchar(20) | not null, default 'deployment' |

值域：`deployment` / `statefulset`，仅 `deploy_type=direct` 时有效。

## 4. Backend Changes

### DirectExecutor (`executor.go`)

- `KubeAppsClient` 接口新增 `StatefulSets(namespace)` 方法
- `Execute()` 根据 `service.WorkloadType` 分支
- 新增 `createK8sStatefulSet` / `updateK8sStatefulSet` 方法
- StatefulSet 与 Deployment 的差异：`ServiceName` 字段 + `UpdateStrategy: RollingUpdate`

### RolloutWatcher (`watcher.go`)

- `watchRollout()` 根据 `service.WorkloadType` 分支
- StatefulSet 就绪判断：`ReadyReplicas >= Replicas && UpdatedReplicas >= Replicas`

### Handler (`service.go`)

- createServiceRequest / updateServiceRequest 接受 `workload_type`
- 校验值域 `deployment` / `statefulset`

## 5. Frontend Changes

- Service 类型新增 `workload_type` 字段
- 创建/编辑服务表单：`deploy_type=direct` 时显示工作负载类型 Radio
- 服务列表/详情显示 workload_type 标签

## 6. Success Criteria

- [ ] StatefulSet 类服务可创建并通过 Direct 部署
- [ ] RolloutWatcher 正确监听 StatefulSet 就绪状态
- [ ] Deployment 类服务完全不受影响
- [ ] Helm 模式不受影响

## 7. Non-Goals

- StatefulSet volumeClaimTemplates 配置
- Headless Service 自动创建
- DaemonSet / Job / CronJob
- 有序更新策略 partition 配置
