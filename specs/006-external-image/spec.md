# 006 - 外部镜像支持

## 1. Overview

为 DeployHub 新增三种镜像来源模式：系统构建（build）、手动指定外部镜像（external）、从 cicd 仓库 app-env.yaml 读取（env_file）。Direct 和 Helm 部署均支持 external 模式，Helm 部署额外支持 env_file 模式。

## 2. Motivation

| 现状 | 目标 |
|------|------|
| 只支持 Kaniko 构建产物 | 支持外部 CI 构建的镜像 |
| Helm 部署 --set image.tag 硬编码 | env_file 模式自动从 app-env.yaml 读取镜像信息 |
| 手动输入 tag 时需要知道 ImageRepo 前缀 | external 模式直接填写完整镜像地址 |

## 3. Data Model Changes

### Deployment 表新增字段

| 字段 | 类型 | 说明 |
|------|------|------|
| image_source | varchar(20) | not null, default 'build'。枚举: build / external / env_file |
| external_image | varchar(500) | nullable, external 模式时的完整镜像地址 |

### Service 表新增字段

| 字段 | 类型 | 说明 |
|------|------|------|
| helm_env_file_path | varchar(255) | nullable, app-env.yaml 相对路径（默认 services/{name}/app-{env}.yaml） |

## 4. 三种模式行为

### 4.1 build 模式（现有，不变）

- 用户选择 Build 记录或手动输入 image_tag
- Direct: `{service.ImageRepo}:{deployment.ImageTag}`
- Helm: `--set image.tag={deployment.ImageTag}`
- `image_source = 'build'`, `external_image = null`

### 4.2 external 模式

- 用户输入完整镜像地址，如 `docker.io/myorg/myapp:v1.2.3`
- 存入 `deployment.external_image`，`image_tag` 可为空或存 tag 部分
- Direct: 容器 Image 直接用 `deployment.ExternalImage`（不拼接 ImageRepo）
- Helm: 从 external_image 拆分出 repository 和 tag → `--set image.repository=xxx --set image.tag=yyy`
- `image_source = 'external'`

### 4.3 env_file 模式（仅 Helm）

- 系统从 cicd 仓库拉取 app-env.yaml，解析 `image.repository`、`image.tag` 字段
- Helm 命令中**不传** `--set image.tag`（由 app-env.yaml 作为 values 文件提供）
- app-env.yaml 作为额外 `-f` 参数加入 helm upgrade 命令
- `image_source = 'env_file'`, `image_tag` 存从文件解析出的值（仅记录）

## 5. API Changes

### 新增 API

| Method | Path | 说明 |
|--------|------|------|
| GET | /api/v1/services/:id/env-image | 解析 cicd 仓库 app-env.yaml 返回镜像信息 |

Response:
```json
{
  "repository": "docker.io/myorg/myapp",
  "tag": "v1.2.3",
  "image_pull_policy": "IfNotPresent",
  "full_image": "docker.io/myorg/myapp:v1.2.3"
}
```

### 修改 API

- `POST /api/v1/deployments` — request 新增 `image_source`（默认 build）和 `external_image`
- `POST /api/v1/deployments/:id/execute` — 根据 image_source 分支执行逻辑

## 6. Backend Changes

### DirectExecutor (`executor.go`)

```go
// Execute 中构造 image 的逻辑改为：
if deployment.ImageSource == "external" && deployment.ExternalImage != "" {
    image = deployment.ExternalImage
} else {
    image = fmt.Sprintf("%s:%s", service.ImageRepo, deployment.ImageTag)
}
```

### HelmExecutor (`helm_job.go`)

```
buildHelmUpgradeCmd 改造：
- image_source == "build":  --set image.tag={tag}（现有逻辑）
- image_source == "external": --set image.repository={repo} --set image.tag={tag}
- image_source == "env_file": 不传 --set image.*（由 app-env.yaml values 文件提供）
  额外添加 -f /workspace/{helm_env_file_path}
```

### env-image 端点 (`handler/service.go`)

- 通过 Git API 或 clone 获取 app-env.yaml 内容
- 解析 YAML 提取 `image.repository`、`image.tag`、`image.imagePullPolicy`
- 返回结构化结果

## 7. Frontend Changes

### 创建部署对话框

- 新增「镜像来源」选择：系统构建 / 外部镜像 / 从文件读取（Helm only）
- **系统构建**：现有 Build 选择 + image_tag 输入（不变）
- **外部镜像**：显示完整镜像地址输入框
- **从文件读取**：调用 env-image API 预览镜像信息，确认后使用

### 部署列表

- 新增 image_source 列标签
- external 模式显示完整镜像地址

## 8. Success Criteria

- [ ] external 模式：Direct 部署使用完整外部镜像地址
- [ ] external 模式：Helm 部署正确拆分并 --set image.repository + image.tag
- [ ] env_file 模式：Helm 部署从 app-env.yaml 读取镜像信息，不传 --set image.tag
- [ ] env-image API 正确解析 app-env.yaml
- [ ] build 模式完全不变
- [ ] 所有测试通过

## 9. Non-Goals

- 验证外部镜像是否存在
- 支持 app-env.yaml 以外的文件
- 嵌套复杂镜像配置
- 多 Registry 凭证
- 修改 Kaniko 构建
