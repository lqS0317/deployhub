# 004 - Helm 部署实现计划

## Layer 1: DB Migrations

### 受影响文件
- **新建** `backend/migrations/000022_add_helm_fields_to_services.up.sql` / `.down.sql`
- **新建** `backend/migrations/000023_add_helm_revision_to_deployments.up.sql` / `.down.sql`
- **新建** `backend/migrations/000024_create_helm_values.up.sql` / `.down.sql`

### 实现要点
- services: deploy_type varchar(10) default 'direct', helm_repo_id nullable FK, helm_chart_path, helm_values_path, helm_release_name, helm_chart_branch
- deployments: helm_revision int nullable
- helm_values: service_id + cluster_id unique, content text, version int, updated_by FK

---

## Layer 2: GORM Models

### 受影响文件
- **修改** `backend/internal/model/service.go` — 新增 DeployType + Helm 字段 + HelmRepo 关联
- **修改** `backend/internal/model/deployment.go` — 新增 HelmRevision
- **新建** `backend/internal/model/helm_values.go`

---

## Layer 3: HelmValues Repository

### 受影响文件
- **新建** `backend/internal/repository/helm_values_repo.go` (接口)
- **新建** `backend/internal/repository/helm_values_repo_impl.go` (实现)

### 接口设计
- FindByServiceAndCluster(serviceID, clusterID) → (*HelmValues, error)
- Upsert(hv *HelmValues) → error (存在则更新+版本自增，不存在则创建)
- ListByService(serviceID) → ([]HelmValues, error)
- ListHistory(serviceID, clusterID) → ([]HelmValuesVersion, error) — 通过版本号倒序

---

## Layer 4: Executor 接口抽取

### 受影响文件
- **修改** `backend/internal/service/deploy/executor.go` — 抽取 DeployExecutor 接口，现有代码重命名为 DirectExecutor
- **修改** `backend/internal/service/deploy/watcher.go` — 不变，仍用于 direct 部署

### 实现要点
```go
type DeployExecutor interface {
    Execute(ctx context.Context, deployment *model.Deployment, service *model.Service) error
}
```
- DirectExecutor 实现 DeployExecutor（重命名自 Executor）
- 所有引用 `*Executor` 的地方改为 `*DirectExecutor` 或 `DeployExecutor`

---

## Layer 5: HelmExecutor

### 受影响文件
- **新建** `backend/internal/service/deploy/helm_executor.go` — HelmExecutor 主体
- **新建** `backend/internal/service/deploy/helm_job.go` — Runner Job spec 生成

### 实现要点
- HelmExecutor: clientPool, gitRepoRepo, cryptoSvc, helmValuesRepo, deployRepo, wsHub
- Execute: 异步创建 Helm Runner Job → Watch Job → 流日志 → 更新状态+helm_revision
- buildRunnerJob: 生成 Job spec (init: git clone, main: helm upgrade --install)
- ensureGitSecret: 创建临时 git 凭证 Secret
- ensureValuesConfigMap: 创建系统 values ConfigMap
- watchHelmJob: 轮询 Job 状态，流式日志
- cleanup: 删除临时 Secret/ConfigMap/Job

---

## Layer 6: Helm Rollback

### 受影响文件
- **新建** `backend/internal/service/deploy/helm_rollback.go`

### 实现要点
- HelmExecutor.Rollback(ctx, deployment, service, revision int) error
- 创建 Runner Job: helm rollback <release> <revision> --namespace <ns>
- 监听 Job 完成 → 更新 deployment 状态

---

## Layer 7: HelmValues Service

### 受影响文件
- **新建** `backend/internal/service/deploy/helm_values.go`

### 实现要点
- HelmValuesService: helmValuesRepo
- GetByServiceAndCluster, Upsert(serviceID, clusterID, content, updatedBy), ListByService

---

## Layer 8: Deploy Service 修改

### 受影响文件
- **修改** `backend/internal/service/deploy/deploy.go` — 新增 ExecuteDeploy 方法

### 实现要点
- 新增 ExecuteDeploy(deploymentID uint) error: 加载 deployment+service → 根据 deploy_type 路由
- deploy_type=direct → directExecutor.Execute + watcher.Watch
- deploy_type=helm → helmExecutor.Execute（内部含 watch）
- StartDeploy → ExecuteDeploy 链式调用

---

## Layer 9: Handler 层

### 受影响文件
- **修改** `backend/internal/handler/deploy.go` — 新增 Approve/Execute 端点, helm-rollback
- **新建** `backend/internal/handler/helm_values.go` — Helm Values CRUD 端点
- **修改** `backend/internal/handler/service.go` — create/update 接受 helm 字段

### 端点
- POST /deployments/:id/execute — 触发执行（审批通过后调用）
- POST /deployments/:id/helm-rollback — Helm 原生回滚
- GET /services/:id/helm-values — 列出各集群 values
- PUT /services/:id/helm-values/:cluster_id — 编辑 values

---

## Layer 10: Route 注册

### 受影响文件
- **修改** `backend/cmd/server/main.go` — 初始化 HelmExecutor, HelmValuesService, 注册路由

---

## Layer 11: Frontend Hooks

### 受影响文件
- **新建** `frontend/src/hooks/use-helm-values.ts`

---

## Layer 12: Frontend UI

### 受影响文件
- **修改** `frontend/src/components/service/service-create-dialog.tsx` — 部署类型选择 + Helm 配置区
- **新建** `frontend/src/components/deploy/helm-values-editor.tsx` — YAML 编辑器
- **修改** `frontend/src/app/(dashboard)/deployments/page.tsx` — helm_revision 显示
- **修改** `frontend/src/types/index.ts` — Service helm 字段, HelmValues 类型

---

## Layer 13: Tests & Verification

- 运行 `go test ./...` 全量通过
- 运行 `next build` 前端编译通过
