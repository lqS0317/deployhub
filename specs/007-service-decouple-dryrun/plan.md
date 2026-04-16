# 007 - Service 解耦 + Dry-Run 实现计划

## Layer 1: DB Migration

### 受影响文件
- **新建** `backend/migrations/000027_decouple_service_cluster.up.sql` / `.down.sql`
- **新建** `backend/migrations/000028_add_preview_fields_to_deployments.up.sql` / `.down.sql`

### 实现要点
- 000027: services 表 cluster_id DROP NOT NULL, namespace DROP NOT NULL SET DEFAULT ''
- 000028: deployments 表新增 preview_yaml text nullable, preview_summary jsonb nullable

---

## Layer 2: Model 更新

### 受影响文件
- **修改** `backend/internal/model/service.go` — ClusterID 改为 *uint（nullable），Namespace 改为 nullable
- **修改** `backend/internal/model/deployment.go` — 新增 PreviewYAML + PreviewSummary 字段 + 3 个新状态常量

---

## Layer 3: Repository 层

### 受影响文件
- **修改** `backend/internal/repository/service_repo.go` — List 移除 clusterID/namespace 参数
- **修改** `backend/internal/repository/service_repo_impl.go` — List 移除 clusterID/namespace 过滤

---

## Layer 4: Service 层

### 受影响文件
- **修改** `backend/internal/service/svc/svc.go` — ServiceService.List 移除 clusterID/namespace
- **修改** `backend/internal/service/deploy/deploy.go` — CreateDeployment 接受 clusterID + namespace 参数
- **修改** `backend/internal/service/deploy/deploy.go` — 新增 Preview/GetPreview/Cancel 方法
- **修改** `backend/internal/service/build/build.go` — buildClusterID=0 时不再 fallback svc.ClusterID

---

## Layer 5: Handler 层（Service）

### 受影响文件
- **修改** `backend/internal/handler/service.go` — createServiceRequest 移除 cluster_id/namespace 必填, List 移除筛选

---

## Layer 6: Handler 层（Deploy）

### 受影响文件
- **修改** `backend/internal/handler/deploy.go` — createDeploymentRequest 新增 cluster_id/namespace
- **修改** `backend/internal/handler/deploy.go` — 新增 Preview/GetPreview/Cancel handler + 路由

---

## Layer 7: Direct Dry-Run

### 受影响文件
- **修改** `backend/internal/service/deploy/executor.go` — DirectExecutor 新增 DryRun 方法
- **新建** `backend/internal/service/deploy/preview.go` — Preview 解析器 + Secret 脱敏

---

## Layer 8: Helm Dry-Run

### 受影响文件
- **修改** `backend/internal/service/deploy/helm_executor.go` — HelmExecutor 新增 Preview 方法
- **修改** `backend/internal/service/deploy/helm_job.go` — 新增 buildHelmTemplateJob 函数

---

## Layer 9: 状态机集成

### 受影响文件
- **修改** `backend/internal/handler/deploy.go` — ExecuteDeploy 改为接受 previewed 状态
- **修改** `backend/cmd/server/main.go` — 注册新路由

---

## Layer 10: Frontend

### 受影响文件
- **修改** `frontend/src/types/index.ts` — Service cluster_id/namespace optional, Deployment 新增字段
- **修改** `frontend/src/components/service/service-create-dialog.tsx` — cluster_id/namespace 改为可选
- **修改** `frontend/src/app/(dashboard)/services/page.tsx` — 移除集群/命名空间列
- **修改** `frontend/src/hooks/use-services.ts` — List 移除 cluster_id/namespace 查询参数
- **修改** deploy 创建对话框 — 新增 cluster_id + namespace 选择
- **新建** deploy 预览面板组件 — 预览 YAML 展示 + 确认/取消

---

## Layer 11: Tests + Verification

- 更新所有涉及 Service.ClusterID/Namespace 的测试
- 运行 go test ./... + next build
