# 010 - 发布管理增强实现计划

## Layer 1: DB Migration
- **新建** `000035_add_fail_reason_to_deployments.up/down.sql`

## Layer 2: Model
- **修改** `model/deployment.go` — 新增 FailReason *string 字段

## Layer 3: Repository
- **修改** `repository/deployment_repo.go` — 接口新增 Delete, UpdateStatusWithReason
- **修改** `repository/deployment_repo_impl.go` — GORM 实现

## Layer 4: Service
- **修改** `service/deploy/deploy.go` — 新增 UpdateStatusWithReason, UpdateDeployment, DeleteDeployment

## Layer 5: Handler — 编辑/删除
- **修改** `handler/deploy.go` — PUT /deployments/:id, DELETE /deployments/:id

## Layer 6: Fail Reason 集成
- **修改** `service/deploy/watcher.go` — markFailed 传 reason
- **修改** `handler/deploy.go` — autoPreview 失败写 reason
- **修改** `service/deploy/helm_executor.go` — failDeploy 写 reason

## Layer 7: Pod API
- **修改** `handler/deploy.go` — GET /deployments/:id/pods

## Layer 8: Pod Log WS
- **修改** `handler/ws.go` — 新增 HandlePodLogWS

## Layer 9: Frontend 编辑/删除
- **修改** deployments/page.tsx + [id]/page.tsx + hooks

## Layer 10: Frontend 失败原因 + 日志面板
- **修改** 详情页 + 新建日志组件

## Layer 11: Tests + Verification
