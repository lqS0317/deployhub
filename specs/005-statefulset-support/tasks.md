# 005 - StatefulSet 支持任务列表

## Phase 1: DB Migration (1 task)

- [ ] **T001** 创建 `000025_add_workload_type_to_services.up/down.sql`

## Phase 2: Service Model (1 task)

- [ ] **T002** 修改 `model/service.go` — 新增 WorkloadType 字段

## Phase 3: DirectExecutor 扩展 (4 tasks)

- [ ] **T003** 扩展 KubeAppsClient 接口 — 新增 StatefulSets(namespace) 方法
- [ ] **T004** 更新 NewDirectExecutor — 返回完整 AppsV1 客户端（已支持 StatefulSets）
- [ ] **T005** Execute 分支 — 根据 service.WorkloadType 路由到 Deployment 或 StatefulSet
- [ ] **T006** 新增 createK8sStatefulSet + updateK8sStatefulSet 方法

## Phase 4: RolloutWatcher 扩展 (2 tasks)

- [ ] **T007** watchRollout 分支 — 根据 service.WorkloadType 查 Deployment 或 StatefulSet
- [ ] **T008** StatefulSet 就绪判断 — ReadyReplicas >= Replicas && UpdatedReplicas >= Replicas

## Phase 5: Handler + 前端 (3 tasks)

- [ ] **T009** 修改 handler/service.go — createServiceRequest 新增 workload_type，写入 model
- [ ] **T010** 修改 frontend types + service-create-dialog — 工作负载类型选择器
- [ ] **T011** 修改 frontend service list — 显示 workload_type 标签

## Phase 6: Tests (2 tasks)

- [ ] **T012** 更新 executor_test.go — 补充 StatefulSet 测试用例
- [ ] **T013** 运行 go test ./... + next build 全量通过

---

**总计: 13 tasks, 6 phases**
