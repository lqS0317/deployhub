# 006 - 外部镜像支持实现计划

## Layer 1: DB Migration

### 受影响文件
- **新建** `backend/migrations/000026_add_external_image_fields.up.sql` / `.down.sql`

### 实现要点
- services 表: 新增 helm_env_file_path varchar(255) nullable
- deployments 表: 新增 image_source varchar(20) not null default 'build', external_image varchar(500) nullable

---

## Layer 2: Model 更新

### 受影响文件
- **修改** `backend/internal/model/service.go` — 新增 HelmEnvFilePath 字段
- **修改** `backend/internal/model/deployment.go` — 新增 ImageSource + ExternalImage 字段

---

## Layer 3: 镜像地址解析工具

### 受影响文件
- **新建** `backend/internal/pkg/image.go` — ParseImageRef(fullImage) → (repository, tag)，ValidateImageRef 校验
- **新建** `backend/internal/pkg/image_test.go` — TDD 先行

### 解析规则
- `docker.io/myorg/myapp:v1.2.3` → repo=`docker.io/myorg/myapp`, tag=`v1.2.3`
- `myapp:latest` → repo=`myapp`, tag=`latest`
- `registry.io/app@sha256:abc` → repo=`registry.io/app`, tag=`sha256:abc`（digest）
- 无冒号 → repo=完整值, tag=`latest`

---

## Layer 4: app-env.yaml 解析服务

### 受影响文件
- **新建** `backend/internal/service/deploy/env_image.go` — ParseEnvImage(yamlContent) → EnvImageInfo
- **新建** `backend/internal/service/deploy/env_image_test.go` — TDD 先行
- **新建** `backend/internal/service/deploy/env_image_fetch.go` — FetchEnvImageFromGit(gitRepoID, branch, filePath) 通过 Git API 获取文件

### 解析结构
```go
type EnvImageInfo struct {
    Repository      string `json:"repository"`
    Tag             string `json:"tag"`
    ImagePullPolicy string `json:"image_pull_policy"`
    FullImage       string `json:"full_image"`
}
```

---

## Layer 5: DirectExecutor 改造

### 受影响文件
- **修改** `backend/internal/service/deploy/executor.go` — buildPodTemplate 中根据 deployment.ImageSource 选择镜像地址

### 逻辑
- image_source == "external" && ExternalImage != "" → image = ExternalImage
- 其他 → image = ImageRepo:ImageTag（现有逻辑）

---

## Layer 6: HelmExecutor 改造

### 受影响文件
- **修改** `backend/internal/service/deploy/helm_job.go` — buildHelmUpgradeCmd 根据 image_source 分支

### 逻辑
- build → `--set image.tag={tag}`（不变）
- external → `--set image.repository={repo} --set image.tag={tag}`（从 ExternalImage 拆分）
- env_file → 不传 `--set image.*`，额外 `-f /workspace/{env_file_path}`

---

## Layer 7: Handler 层

### 受影响文件
- **新建** `backend/internal/handler/env_image.go` — GET /services/:id/env-image 端点
- **修改** `backend/internal/handler/deploy.go` — createDeploymentRequest 新增 image_source + external_image
- **修改** `backend/internal/service/deploy/deploy.go` — CreateDeployment 处理新字段
- **修改** `backend/cmd/server/main.go` — 注册 env-image 路由

---

## Layer 8: Frontend 部署表单

### 受影响文件
- **修改** `frontend/src/types/index.ts` — Deployment 新增 image_source + external_image, Service 新增 helm_env_file_path
- **修改** `frontend/src/components/deploy/deploy-dialog.tsx` 或触发构建对话框 — 镜像来源选择
- **新建** `frontend/src/hooks/use-env-image.ts` — useEnvImage(serviceId) hook

---

## Layer 9: Frontend 部署列表

### 受影响文件
- **修改** `frontend/src/app/(dashboard)/deployments/page.tsx` — image_source 标签 + external_image 显示

---

## Layer 10: Tests + Verification

- 运行 `go test ./...` 全量通过
- 运行 `next build` 前端编译通过
