# 005 - StatefulSet 支持实现计划

## Layer 1: DB Migration

- **新建** `backend/migrations/000025_add_workload_type_to_services.up.sql` / `.down.sql`
- services 表新增 workload_type varchar(20) not null default 'deployment'

## Layer 2: Service Model

- **修改** `backend/internal/model/service.go` — 新增 WorkloadType 字段

## Layer 3: DirectExecutor 扩展

- **修改** `backend/internal/service/deploy/executor.go`
  - KubeAppsClient 接口新增 StatefulSets(namespace) 方法
  - NewDirectExecutor 返回同时支持 Deployment 和 StatefulSet 的客户端
  - Execute() 根据 service.WorkloadType 分支
  - 新增 createK8sStatefulSet / updateK8sStatefulSet 方法

## Layer 4: RolloutWatcher 扩展

- **修改** `backend/internal/service/deploy/watcher.go`
  - watchRollout() 根据 service.WorkloadType 分支
  - StatefulSet: cs.AppsV1().StatefulSets(ns).Get() → 判断 ReadyReplicas

## Layer 5: Handler 更新

- **修改** `backend/internal/handler/service.go` — createServiceRequest 新增 workload_type 字段

## Layer 6: Frontend

- **修改** `frontend/src/types/index.ts` — Service 新增 workload_type
- **修改** `frontend/src/components/service/service-create-dialog.tsx` — 工作负载类型选择器

## Layer 7: Tests

- **修改** `backend/internal/service/deploy/executor_test.go` — StatefulSet 测试用例
- 运行 go test ./... + next build
