# 006 - 外部镜像支持任务列表

## Phase 1: DB Migration (1 task)

- [ ] **T001** 创建 `000026_add_external_image_fields.up/down.sql` — services: helm_env_file_path; deployments: image_source + external_image

## Phase 2: Model 更新 (2 tasks)

- [ ] **T002** 修改 `model/service.go` — 新增 HelmEnvFilePath 字段
- [ ] **T003** 修改 `model/deployment.go` — 新增 ImageSource + ExternalImage 字段

## Phase 3: 镜像地址解析工具 (2 tasks, TDD)

- [ ] **T004** 新建 `pkg/image_test.go` — TestParseImageRef 测试用例：完整地址/无tag/digest/空字符串
- [ ] **T005** 新建 `pkg/image.go` — ParseImageRef + ValidateImageRef 实现

## Phase 4: app-env.yaml 解析 (3 tasks, TDD)

- [ ] **T006** 新建 `service/deploy/env_image_test.go` — TestParseEnvImage 测试用例：标准/缺字段/空文件
- [ ] **T007** 新建 `service/deploy/env_image.go` — ParseEnvImage(yamlContent) 解析 image.* 字段
- [ ] **T008** 新建 `service/deploy/env_image_fetch.go` — FetchEnvImageFromGit 通过 Git API 获取 app-env.yaml 并解析

## Phase 5: DirectExecutor 改造 (2 tasks)

- [ ] **T009** 修改 `executor.go` — buildPodTemplate 接受 deployment 参数，external 模式直接用 ExternalImage
- [ ] **T010** 更新 executor_test.go — 新增 external image 模式测试用例

## Phase 6: HelmExecutor 改造 (3 tasks)

- [ ] **T011** 修改 `helm_job.go` — buildHelmUpgradeCmd 接受 imageSource 参数
- [ ] **T012** build 模式不变，external 模式拆分 → --set image.repository + --set image.tag
- [ ] **T013** env_file 模式不传 --set image.*，额外 -f /workspace/{env_file_path}

## Phase 7: Handler 层 (4 tasks)

- [ ] **T014** 新建 `handler/env_image.go` — GET /services/:id/env-image 端点
- [ ] **T015** 修改 `handler/deploy.go` — createDeploymentRequest 新增 image_source + external_image 字段
- [ ] **T016** 修改 `service/deploy/deploy.go` — CreateDeployment 处理 ImageSource + ExternalImage
- [ ] **T017** 修改 `cmd/server/main.go` — 注册 env-image 路由

## Phase 8: Frontend (4 tasks)

- [ ] **T018** 修改 `types/index.ts` — Deployment + Service 新增字段类型
- [ ] **T019** 新建 `hooks/use-env-image.ts` — useEnvImage(serviceId) hook
- [ ] **T020** 修改部署触发对话框 — 镜像来源三选一 + 条件渲染输入区域
- [ ] **T021** 修改部署列表 — image_source 标签显示

## Phase 9: Verification (1 task)

- [ ] **T022** 运行 go test ./... + next build 全量通过

---

**总计: 22 tasks, 9 phases**
