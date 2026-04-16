# DeployHub API 合约

本文档描述 DeployHub DevOps 部署平台的 HTTP API 与 WebSocket 约定。

| 项 | 说明 |
| --- | --- |
| **Base URL** | `/api/v1` |
| **Content-Type** | `application/json`（文件上传接口除外） |
| **认证** | 除标注为 Public 的接口外，请求头需携带 `Authorization: Bearer <JWT>` |

---

## 通用约定

### 错误响应

未特别说明时，错误体可为：

```json
{
  "error": {
    "code": "STRING_CODE",
    "message": "人类可读说明",
    "details": {}
  }
}
```

### 分页（列表类接口）

若支持分页，查询参数可为 `page`、`page_size`；响应可含 `items`、`total`、`page`、`page_size`（具体字段以实现为准）。

### ID 类型

路径中的 `:id`、`:member_id` 等若无特殊说明，为字符串或 UUID，与实现一致。

---

## Auth 模块

### POST `/api/v1/auth/register`

| 项 | 内容 |
| --- | --- |
| **说明** | 本地用户注册。 |
| **Auth** | Public |

**Request body**

```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "********"
}
```

**Response** `201 Created`

```json
{
  "user": {
    "id": "usr_xxx",
    "username": "alice",
    "email": "alice@example.com",
    "created_at": "2026-04-03T00:00:00Z"
  }
}
```

**Status codes** | `201` 成功 | `400` 参数错误 | `409` 用户名或邮箱已存在

---

### POST `/api/v1/auth/login`

| 项 | 内容 |
| --- | --- |
| **说明** | 本地登录，返回 JWT。 |
| **Auth** | Public |

**Request body**

```json
{
  "username": "alice",
  "password": "********"
}
```

**Response** `200 OK`

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": "usr_xxx",
    "username": "alice",
    "email": "alice@example.com"
  }
}
```

**Status codes** | `200` 成功 | `401` 凭据错误

---

### POST `/api/v1/auth/oauth/:provider/callback`

| 项 | 内容 |
| --- | --- |
| **说明** | OAuth2 授权码回调，交换并签发 JWT。`:provider` 为第三方标识（如 github、gitlab）。 |
| **Auth** | Public |

**Request body**

```json
{
  "code": "oauth_authorization_code",
  "state": "optional_csrf_state"
}
```

**Response** `200 OK`

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": "usr_xxx",
    "username": "alice",
    "email": "alice@example.com"
  }
}
```

**Status codes** | `200` 成功 | `400` code/state 无效 | `401` 交换失败

---

### GET `/api/v1/auth/me`

| 项 | 内容 |
| --- | --- |
| **说明** | 获取当前登录用户信息。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "id": "usr_xxx",
  "username": "alice",
  "email": "alice@example.com",
  "role": "member",
  "status": "active",
  "created_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `401` 未认证

---

### PUT `/api/v1/auth/me`

| 项 | 内容 |
| --- | --- |
| **说明** | 更新当前用户个人信息（如显示名、邮箱等，字段以实现为准）。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "display_name": "Alice",
  "email": "new@example.com"
}
```

**Response** `200 OK`

```json
{
  "id": "usr_xxx",
  "username": "alice",
  "email": "new@example.com",
  "display_name": "Alice",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `400` 校验失败 | `401` 未认证 | `409` 邮箱冲突

---

## Cluster 管理

### GET `/api/v1/clusters`

| 项 | 内容 |
| --- | --- |
| **说明** | 分页或全量返回集群列表。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "items": [
    {
      "id": "clu_xxx",
      "name": "prod-k8s",
      "display_name": "生产集群",
      "env": "production",
      "api_server": "https://k8s.example.com:6443",
      "created_at": "2026-04-03T00:00:00Z"
    }
  ],
  "total": 1
}
```

**Status codes** | `200` 成功 | `401` 未认证

---

### POST `/api/v1/clusters`

| 项 | 内容 |
| --- | --- |
| **说明** | 创建集群并保存 kubeconfig 等连接信息。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "name": "prod-k8s",
  "display_name": "生产集群",
  "env": "production",
  "api_server": "https://k8s.example.com:6443",
  "kubeconfig": "apiVersion: v1\nkind: Config\n..."
}
```

**Response** `201 Created`

```json
{
  "id": "clu_xxx",
  "name": "prod-k8s",
  "display_name": "生产集群",
  "env": "production",
  "api_server": "https://k8s.example.com:6443",
  "created_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `201` 成功 | `400` 参数错误 | `401` 未认证

---

### GET `/api/v1/clusters/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 获取指定集群详情（kubeconfig 等敏感字段可能脱敏或省略）。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "id": "clu_xxx",
  "name": "prod-k8s",
  "display_name": "生产集群",
  "env": "production",
  "api_server": "https://k8s.example.com:6443",
  "last_test_status": "ok",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在

---

### PUT `/api/v1/clusters/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 更新集群元数据或 kubeconfig。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "display_name": "生产 K8s",
  "api_server": "https://k8s.example.com:6443",
  "kubeconfig": "..."
}
```

**Response** `200 OK`

```json
{
  "id": "clu_xxx",
  "name": "prod-k8s",
  "display_name": "生产 K8s",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `400` 参数错误 | `401` 未认证 | `404` 不存在

---

### DELETE `/api/v1/clusters/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 删除集群记录。 |
| **Auth** | JWT required |

**Response** `204 No Content`（或 `200` 带简短确认体，以实现为准）

**Status codes** | `204`/`200` 成功 | `401` 未认证 | `404` 不存在 | `409` 仍存在依赖资源

---

### POST `/api/v1/clusters/:id/test`

| 项 | 内容 |
| --- | --- |
| **说明** | 使用已存凭证测试与 Kubernetes API 的连通性。 |
| **Auth** | JWT required |

**Request body** | 无（或空对象 `{}`）

**Response** `200 OK`

```json
{
  "success": true,
  "message": "连接成功",
  "server_version": "v1.29.0"
}
```

**Status codes** | `200` 成功（业务上可能 `success: false`） | `401` 未认证 | `404` 集群不存在

---

## Git 仓库管理

### GET `/api/v1/git-repos`

| 项 | 内容 |
| --- | --- |
| **说明** | Git 仓库列表。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "items": [
    {
      "id": "git_xxx",
      "name": "app-frontend",
      "url": "https://github.com/org/app.git",
      "provider": "github",
      "auth_type": "token",
      "created_at": "2026-04-03T00:00:00Z"
    }
  ],
  "total": 1
}
```

**Status codes** | `200` 成功 | `401` 未认证

---

### POST `/api/v1/git-repos`

| 项 | 内容 |
| --- | --- |
| **说明** | 添加 Git 仓库及凭据。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "name": "app-frontend",
  "url": "https://github.com/org/app.git",
  "provider": "github",
  "auth_type": "token",
  "credential": {
    "token": "ghp_***"
  }
}
```

**Response** `201 Created`

```json
{
  "id": "git_xxx",
  "name": "app-frontend",
  "url": "https://github.com/org/app.git",
  "provider": "github",
  "auth_type": "token",
  "created_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `201` 成功 | `400` 参数错误 | `401` 未认证

---

### GET `/api/v1/git-repos/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 仓库详情（凭据字段应脱敏）。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "id": "git_xxx",
  "name": "app-frontend",
  "url": "https://github.com/org/app.git",
  "provider": "github",
  "auth_type": "token",
  "last_test_status": "ok",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在

---

### PUT `/api/v1/git-repos/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 更新仓库 URL、凭据等。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "name": "app-frontend",
  "url": "https://github.com/org/app.git",
  "credential": {
    "token": "ghp_new"
  }
}
```

**Response** `200 OK`

```json
{
  "id": "git_xxx",
  "name": "app-frontend",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `400` 参数错误 | `401` 未认证 | `404` 不存在

---

### DELETE `/api/v1/git-repos/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 删除仓库配置。 |
| **Auth** | JWT required |

**Response** `204 No Content`

**Status codes** | `204`/`200` 成功 | `401` 未认证 | `404` 不存在 | `409` 存在依赖

---

### POST `/api/v1/git-repos/:id/test`

| 项 | 内容 |
| --- | --- |
| **说明** | 测试仓库克隆/拉取权限与连通性。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "success": true,
  "message": "认证成功",
  "default_branch": "main"
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在

---

### GET `/api/v1/git-repos/:id/branches`

| 项 | 内容 |
| --- | --- |
| **说明** | 列出远程分支（名称列表）。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "branches": [
    { "name": "main", "is_default": true },
    { "name": "develop", "is_default": false }
  ]
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在 | `502` 上游 Git 失败

---

## Registry 管理

### GET `/api/v1/registries`

| 项 | 内容 |
| --- | --- |
| **说明** | 容器镜像 Registry 列表。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "items": [
    {
      "id": "reg_xxx",
      "name": "company-harbor",
      "url": "https://harbor.example.com",
      "provider": "harbor",
      "created_at": "2026-04-03T00:00:00Z"
    }
  ],
  "total": 1
}
```

**Status codes** | `200` 成功 | `401` 未认证

---

### POST `/api/v1/registries`

| 项 | 内容 |
| --- | --- |
| **说明** | 添加镜像仓库及认证配置。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "name": "company-harbor",
  "url": "https://harbor.example.com",
  "provider": "harbor",
  "auth_config": {
    "username": "robot$deploy",
    "password": "***"
  }
}
```

**Response** `201 Created`

```json
{
  "id": "reg_xxx",
  "name": "company-harbor",
  "url": "https://harbor.example.com",
  "provider": "harbor",
  "created_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `201` 成功 | `400` 参数错误 | `401` 未认证

---

### GET `/api/v1/registries/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | Registry 详情。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "id": "reg_xxx",
  "name": "company-harbor",
  "url": "https://harbor.example.com",
  "provider": "harbor",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在

---

### PUT `/api/v1/registries/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 更新 Registry 地址或 `auth_config`。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "url": "https://harbor.example.com",
  "auth_config": {
    "username": "robot$deploy",
    "password": "***"
  }
}
```

**Response** `200 OK`

```json
{
  "id": "reg_xxx",
  "name": "company-harbor",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `400` 参数错误 | `401` 未认证 | `404` 不存在

---

### DELETE `/api/v1/registries/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 删除 Registry。 |
| **Auth** | JWT required |

**Response** `204 No Content`

**Status codes** | `204`/`200` 成功 | `401` 未认证 | `404` 不存在 | `409` 存在依赖

---

## Service 管理

### GET `/api/v1/services`

| 项 | 内容 |
| --- | --- |
| **说明** | 服务列表；支持查询参数 `cluster_id`、`namespace` 过滤。 |
| **Auth** | JWT required |

**Query** | `cluster_id`、`namespace`（可选）

**Response** `200 OK`

```json
{
  "items": [
    {
      "id": "svc_xxx",
      "name": "payment-api",
      "cluster_id": "clu_xxx",
      "namespace": "prod",
      "git_repo_id": "git_xxx",
      "created_at": "2026-04-03T00:00:00Z"
    }
  ],
  "total": 1
}
```

**Status codes** | `200` 成功 | `401` 未认证

---

### POST `/api/v1/services`

| 项 | 内容 |
| --- | --- |
| **说明** | 创建服务（关联集群、Git、构建参数等，字段以实现为准）。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "name": "payment-api",
  "cluster_id": "clu_xxx",
  "namespace": "prod",
  "git_repo_id": "git_xxx",
  "registry_id": "reg_xxx",
  "dockerfile_path": "Dockerfile"
}
```

**Response** `201 Created`

```json
{
  "id": "svc_xxx",
  "name": "payment-api",
  "cluster_id": "clu_xxx",
  "namespace": "prod",
  "created_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `201` 成功 | `400` 参数错误 | `401` 未认证

---

### GET `/api/v1/services/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 服务详情。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "id": "svc_xxx",
  "name": "payment-api",
  "cluster_id": "clu_xxx",
  "namespace": "prod",
  "git_repo_id": "git_xxx",
  "registry_id": "reg_xxx",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在

---

### PUT `/api/v1/services/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 更新服务配置。 |
| **Auth** | JWT required |

**Request body** | 部分字段可更新，与创建类似

**Response** `200 OK`

```json
{
  "id": "svc_xxx",
  "name": "payment-api",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `400` 参数错误 | `401` 未认证 | `404` 不存在

---

### DELETE `/api/v1/services/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 删除服务。 |
| **Auth** | JWT required |

**Response** `204 No Content`

**Status codes** | `204`/`200` 成功 | `401` 未认证 | `404` 不存在 | `409` 存在构建/发布记录

---

### POST `/api/v1/services/import`

| 项 | 内容 |
| --- | --- |
| **说明** | 批量从 YAML/JSON 文件导入服务定义；`multipart/form-data` 上传文件。 |
| **Auth** | JWT required |

**Request** | `multipart/form-data`，字段如 `file`（二进制）

**Response** `200 OK`

```json
{
  "imported": 3,
  "failed": 0,
  "results": [
    { "name": "svc-a", "id": "svc_xxx", "status": "created" }
  ]
}
```

**Status codes** | `200` 成功 | `400` 文件格式错误 | `401` 未认证 | `413` 文件过大

---

### POST `/api/v1/services/import/preview`

| 项 | 内容 |
| --- | --- |
| **说明** | 上传同样格式文件，仅解析预览，不落库。 |
| **Auth** | JWT required |

**Request** | `multipart/form-data`

**Response** `200 OK`

```json
{
  "preview": [
    {
      "name": "payment-api",
      "cluster": "prod-k8s",
      "namespace": "prod",
      "valid": true,
      "errors": []
    }
  ]
}
```

**Status codes** | `200` 成功 | `400` 解析失败 | `401` 未认证

---

### GET `/api/v1/services/:id/members`

| 项 | 内容 |
| --- | --- |
| **说明** | 服务成员与角色列表。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "items": [
    {
      "member_id": "mem_xxx",
      "user_id": "usr_xxx",
      "username": "alice",
      "role": "owner"
    }
  ]
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `403` 无权限 | `404` 服务不存在

---

### POST `/api/v1/services/:id/members`

| 项 | 内容 |
| --- | --- |
| **说明** | 为服务添加成员。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "user_id": "usr_yyy",
  "role": "developer"
}
```

**Response** `201 Created`

```json
{
  "member_id": "mem_yyy",
  "user_id": "usr_yyy",
  "role": "developer"
}
```

**Status codes** | `201` 成功 | `400` 参数错误 | `401` 未认证 | `403` 无权限 | `404` 服务/用户不存在 | `409` 已是成员

---

### PUT `/api/v1/services/:id/members/:member_id`

| 项 | 内容 |
| --- | --- |
| **说明** | 更新成员角色。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "role": "maintainer"
}
```

**Response** `200 OK`

```json
{
  "member_id": "mem_yyy",
  "user_id": "usr_yyy",
  "role": "maintainer"
}
```

**Status codes** | `200` 成功 | `400` 非法角色 | `401` 未认证 | `403` 无权限 | `404` 不存在

---

### DELETE `/api/v1/services/:id/members/:member_id`

| 项 | 内容 |
| --- | --- |
| **说明** | 移除服务成员。 |
| **Auth** | JWT required |

**Response** `204 No Content`

**Status codes** | `204`/`200` 成功 | `401` 未认证 | `403` 无权限 | `404` 不存在

---

## Build 构建

### GET `/api/v1/builds`

| 项 | 内容 |
| --- | --- |
| **说明** | 构建记录列表；支持 `service_id` 过滤。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "items": [
    {
      "id": "bld_xxx",
      "service_id": "svc_xxx",
      "status": "running",
      "git_branch": "main",
      "git_commit": "abc1234",
      "created_at": "2026-04-03T00:00:00Z"
    }
  ],
  "total": 1
}
```

**Status codes** | `200` 成功 | `401` 未认证

---

### POST `/api/v1/builds`

| 项 | 内容 |
| --- | --- |
| **说明** | 触发一次新构建。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "service_id": "svc_xxx",
  "git_branch": "main",
  "git_commit": "abc1234"
}
```

**Response** `202 Accepted`（或 `201 Created`）

```json
{
  "id": "bld_xxx",
  "service_id": "svc_xxx",
  "status": "queued",
  "git_branch": "main",
  "git_commit": "abc1234",
  "created_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `201`/`202` 已受理 | `400` 参数错误 | `401` 未认证 | `403` 无权限 | `404` 服务不存在

---

### GET `/api/v1/builds/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 构建详情（状态、镜像标签、错误信息等）。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "id": "bld_xxx",
  "service_id": "svc_xxx",
  "status": "succeeded",
  "git_branch": "main",
  "git_commit": "abc1234",
  "image": "harbor.example.com/proj/app:abc1234",
  "started_at": "2026-04-03T00:00:00Z",
  "finished_at": "2026-04-03T00:05:00Z"
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在

---

### POST `/api/v1/builds/:id/cancel`

| 项 | 内容 |
| --- | --- |
| **说明** | 取消进行中的构建。 |
| **Auth** | JWT required |

**Request body** | 无或 `{}`

**Response** `200 OK`

```json
{
  "id": "bld_xxx",
  "status": "cancelled"
}
```

**Status codes** | `200` 成功 | `400` 不可取消 | `401` 未认证 | `404` 不存在 | `409` 已结束

---

### GET `/api/v1/builds/:id/log`

| 项 | 内容 |
| --- | --- |
| **说明** | 获取构建日志文本或分页块（实现可返回 URL 或内联文本）。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "build_id": "bld_xxx",
  "log": "Step 1/5 : FROM node:20\n...",
  "truncated": false
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在

---

## Deployment 发布

### GET `/api/v1/deployments`

| 项 | 内容 |
| --- | --- |
| **说明** | 发布记录列表；支持 `service_id` 过滤。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "items": [
    {
      "id": "dep_xxx",
      "service_id": "svc_xxx",
      "build_id": "bld_xxx",
      "cluster_id": "clu_xxx",
      "namespace": "prod",
      "status": "running",
      "created_at": "2026-04-03T00:00:00Z"
    }
  ],
  "total": 1
}
```

**Status codes** | `200` 成功 | `401` 未认证

---

### POST `/api/v1/deployments`

| 项 | 内容 |
| --- | --- |
| **说明** | 使用指定构建在目标集群发起发布。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "service_id": "svc_xxx",
  "build_id": "bld_xxx",
  "cluster_id": "clu_xxx",
  "namespace": "prod",
  "replicas": 3
}
```

**Response** `202 Accepted`

```json
{
  "id": "dep_xxx",
  "service_id": "svc_xxx",
  "build_id": "bld_xxx",
  "status": "pending",
  "created_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `201`/`202` 已受理 | `400` 参数错误 | `401` 未认证 | `403` 无权限 | `404` 资源不存在

---

### GET `/api/v1/deployments/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 发布详情（阶段、副本、错误等）。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "id": "dep_xxx",
  "service_id": "svc_xxx",
  "build_id": "bld_xxx",
  "cluster_id": "clu_xxx",
  "namespace": "prod",
  "replicas": 3,
  "status": "succeeded",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在

---

### POST `/api/v1/deployments/:id/rollback`

| 项 | 内容 |
| --- | --- |
| **说明** | 回滚到某次历史发布对应状态。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "target_deployment_id": "dep_prev"
}
```

**Response** `202 Accepted`

```json
{
  "id": "dep_new",
  "rollback_from": "dep_xxx",
  "target_deployment_id": "dep_prev",
  "status": "pending"
}
```

**Status codes** | `202`/`201` 已受理 | `400` 参数错误 | `401` 未认证 | `404` 不存在 | `409` 不可回滚

---

## Approval 审批

### GET `/api/v1/approvals`

| 项 | 内容 |
| --- | --- |
| **说明** | 审批单列表；支持 `status` 过滤，默认仅 `pending`。 |
| **Auth** | JWT required |

**Query** | `status=pending|approved|rejected|all`（默认 `pending`）

**Response** `200 OK`

```json
{
  "items": [
    {
      "id": "apr_xxx",
      "resource_type": "deployment",
      "resource_id": "dep_xxx",
      "status": "pending",
      "created_at": "2026-04-03T00:00:00Z"
    }
  ],
  "total": 1
}
```

**Status codes** | `200` 成功 | `401` 未认证

---

### GET `/api/v1/approvals/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 审批单详情。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "id": "apr_xxx",
  "resource_type": "deployment",
  "resource_id": "dep_xxx",
  "status": "pending",
  "requester": { "user_id": "usr_xxx", "username": "alice" },
  "payload": {}
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在

---

### POST `/api/v1/approvals/:id/approve`

| 项 | 内容 |
| --- | --- |
| **说明** | 审批通过。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "comment": "同意上线"
}
```

**Response** `200 OK`

```json
{
  "id": "apr_xxx",
  "status": "approved",
  "comment": "同意上线",
  "decided_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `403` 非审批人 | `404` 不存在 | `409` 已处理

---

### POST `/api/v1/approvals/:id/reject`

| 项 | 内容 |
| --- | --- |
| **说明** | 审批拒绝。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "comment": "风险未评估"
}
```

**Response** `200 OK`

```json
{
  "id": "apr_xxx",
  "status": "rejected",
  "comment": "风险未评估",
  "decided_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `403` 非审批人 | `404` 不存在 | `409` 已处理

---

## Config 配置

### GET `/api/v1/services/:id/configs`

| 项 | 内容 |
| --- | --- |
| **说明** | 某服务下配置模板列表。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "items": [
    {
      "id": "cfg_xxx",
      "name": "app-env",
      "config_type": "env",
      "updated_at": "2026-04-03T00:00:00Z"
    }
  ]
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 服务不存在

---

### POST `/api/v1/services/:id/configs`

| 项 | 内容 |
| --- | --- |
| **说明** | 创建配置模板。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "name": "app-env",
  "config_type": "env",
  "template_content": "FOO={{ .FOO }}\nBAR={{ .BAR }}"
}
```

**Response** `201 Created`

```json
{
  "id": "cfg_xxx",
  "service_id": "svc_xxx",
  "name": "app-env",
  "config_type": "env",
  "created_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `201` 成功 | `400` 参数错误 | `401` 未认证 | `404` 服务不存在

---

### GET `/api/v1/configs/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 配置模板详情。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "id": "cfg_xxx",
  "service_id": "svc_xxx",
  "name": "app-env",
  "config_type": "env",
  "template_content": "FOO={{ .FOO }}"
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在

---

### PUT `/api/v1/configs/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 更新模板名称或内容。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "name": "app-env-v2",
  "template_content": "FOO={{ .FOO }}\nBAZ={{ .BAZ }}"
}
```

**Response** `200 OK`

```json
{
  "id": "cfg_xxx",
  "version": 2,
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `400` 参数错误 | `401` 未认证 | `404` 不存在

---

### DELETE `/api/v1/configs/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 删除配置模板。 |
| **Auth** | JWT required |

**Response** `204 No Content`

**Status codes** | `204`/`200` 成功 | `401` 未认证 | `404` 不存在

---

### GET `/api/v1/configs/:id/env-values`

| 项 | 内容 |
| --- | --- |
| **说明** | 获取各环境（集群）下的变量键值（敏感值可掩码）。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "values_by_cluster": [
    {
      "cluster_id": "clu_xxx",
      "cluster_name": "prod-k8s",
      "vars": { "FOO": "prod-value", "BAR": "***" }
    }
  ]
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在

---

### PUT `/api/v1/configs/:id/env-values/:cluster_id`

| 项 | 内容 |
| --- | --- |
| **说明** | 更新指定集群环境下的变量集合。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "vars": {
    "FOO": "new-value",
    "BAR": "secret"
  }
}
```

**Response** `200 OK`

```json
{
  "config_id": "cfg_xxx",
  "cluster_id": "clu_xxx",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `400` 参数错误 | `401` 未认证 | `404` 不存在

---

### POST `/api/v1/configs/:id/render`

| 项 | 内容 |
| --- | --- |
| **说明** | 按某集群变量预览渲染后的配置内容。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "cluster_id": "clu_xxx"
}
```

**Response** `200 OK`

```json
{
  "rendered": "FOO=prod-value\nBAR=secret"
}
```

**Status codes** | `200` 成功 | `400` 参数错误 | `401` 未认证 | `404` 不存在

---

### GET `/api/v1/configs/:id/versions`

| 项 | 内容 |
| --- | --- |
| **说明** | 配置模板历史版本列表。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "items": [
    {
      "version_id": "ver_2",
      "version": 2,
      "created_at": "2026-04-03T00:00:00Z",
      "created_by": "usr_xxx"
    }
  ]
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在

---

### GET `/api/v1/configs/:id/versions/:version_id/diff`

| 项 | 内容 |
| --- | --- |
| **说明** | 对比两个版本差异；目标版本由查询参数指定。 |
| **Auth** | JWT required |

**Query** | `target_version_id=ver_1`

**Response** `200 OK`

```json
{
  "from_version_id": "ver_2",
  "to_version_id": "ver_1",
  "diff": "--- a\n+++ b\n@@ ... @@"
}
```

**Status codes** | `200` 成功 | `400` 缺少参数 | `401` 未认证 | `404` 版本不存在

---

### POST `/api/v1/configs/:id/deploy`

| 项 | 内容 |
| --- | --- |
| **说明** | 将渲染后的配置下发到指定集群与命名空间（如 ConfigMap/Secret）。 |
| **Auth** | JWT required |

**Request body**

```json
{
  "cluster_id": "clu_xxx",
  "namespace": "prod"
}
```

**Response** `202 Accepted`

```json
{
  "job_id": "job_xxx",
  "status": "queued"
}
```

**Status codes** | `202`/`200` 已受理 | `400` 参数错误 | `401` 未认证 | `404` 不存在 | `502` 集群操作失败

---

## Notification 通知

### GET `/api/v1/notifications`

| 项 | 内容 |
| --- | --- |
| **说明** | 当前用户通知列表；支持 `is_read=true|false` 过滤。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "items": [
    {
      "id": "ntf_xxx",
      "title": "构建成功",
      "body": "服务 payment-api 构建完成",
      "is_read": false,
      "created_at": "2026-04-03T00:00:00Z"
    }
  ],
  "total": 1
}
```

**Status codes** | `200` 成功 | `401` 未认证

---

### PUT `/api/v1/notifications/:id/read`

| 项 | 内容 |
| --- | --- |
| **说明** | 将单条通知标记为已读。 |
| **Auth** | JWT required |

**Request body** | 无或 `{}`

**Response** `200 OK`

```json
{
  "id": "ntf_xxx",
  "is_read": true
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `404` 不存在

---

### PUT `/api/v1/notifications/read-all`

| 项 | 内容 |
| --- | --- |
| **说明** | 当前用户全部通知标记为已读。 |
| **Auth** | JWT required |

**Request body** | 无

**Response** `200 OK`

```json
{
  "updated": 42
}
```

**Status codes** | `200` 成功 | `401` 未认证

---

### GET `/api/v1/notifications/unread-count`

| 项 | 内容 |
| --- | --- |
| **说明** | 未读通知数量。 |
| **Auth** | JWT required |

**Response** `200 OK`

```json
{
  "unread_count": 5
}
```

**Status codes** | `200` 成功 | `401` 未认证

---

## AuditLog 审计

### GET `/api/v1/audit-logs`

| 项 | 内容 |
| --- | --- |
| **说明** | 审计日志列表；支持 `user_id`、`action`、`resource_type`、时间范围（如 `from`、`to`）过滤。 |
| **Auth** | JWT required（通常仅管理员可访问，以实现为准） |

**Response** `200 OK`

```json
{
  "items": [
    {
      "id": "aud_xxx",
      "user_id": "usr_xxx",
      "action": "deployment.create",
      "resource_type": "deployment",
      "resource_id": "dep_xxx",
      "ip": "10.0.0.1",
      "created_at": "2026-04-03T00:00:00Z"
    }
  ],
  "total": 1
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `403` 无权限

---

## User 管理（Admin）

### GET `/api/v1/users`

| 项 | 内容 |
| --- | --- |
| **说明** | 全站用户列表（分页）。 |
| **Auth** | Admin only |

**Response** `200 OK`

```json
{
  "items": [
    {
      "id": "usr_xxx",
      "username": "alice",
      "email": "alice@example.com",
      "role": "member",
      "status": "active"
    }
  ],
  "total": 1
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `403` 非管理员

---

### PUT `/api/v1/users/:id/role`

| 项 | 内容 |
| --- | --- |
| **说明** | 更新用户全局角色。 |
| **Auth** | Admin only |

**Request body**

```json
{
  "role": "admin"
}
```

**Response** `200 OK`

```json
{
  "id": "usr_xxx",
  "role": "admin",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `400` 非法 role | `401` 未认证 | `403` 非管理员 | `404` 用户不存在

---

### PUT `/api/v1/users/:id/status`

| 项 | 内容 |
| --- | --- |
| **说明** | 启用或禁用用户账号。 |
| **Auth** | Admin only |

**Request body**

```json
{
  "status": "disabled"
}
```

**Response** `200 OK`

```json
{
  "id": "usr_xxx",
  "status": "disabled",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `400` 非法 status | `401` 未认证 | `403` 非管理员 | `404` 不存在

---

## Notification Channel 通知渠道（Admin）

### GET `/api/v1/notification-channels`

| 项 | 内容 |
| --- | --- |
| **说明** | 系统级通知渠道（飞书/钉钉/Slack 等）列表。 |
| **Auth** | Admin only |

**Response** `200 OK`

```json
{
  "items": [
    {
      "id": "nch_xxx",
      "name": "feishu-deploy",
      "type": "feishu",
      "webhook_url": "https://open.feishu.cn/...",
      "created_at": "2026-04-03T00:00:00Z"
    }
  ]
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `403` 非管理员

---

### POST `/api/v1/notification-channels`

| 项 | 内容 |
| --- | --- |
| **说明** | 创建通知渠道。 |
| **Auth** | Admin only |

**Request body**

```json
{
  "type": "feishu",
  "name": "feishu-deploy",
  "webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx"
}
```

**Response** `201 Created`

```json
{
  "id": "nch_xxx",
  "type": "feishu",
  "name": "feishu-deploy",
  "created_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `201` 成功 | `400` 参数错误 | `401` 未认证 | `403` 非管理员

---

### PUT `/api/v1/notification-channels/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 更新渠道名称或 Webhook。 |
| **Auth** | Admin only |

**Request body**

```json
{
  "name": "feishu-deploy-v2",
  "webhook_url": "https://open.feishu.cn/..."
}
```

**Response** `200 OK`

```json
{
  "id": "nch_xxx",
  "updated_at": "2026-04-03T00:00:00Z"
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `403` 非管理员 | `404` 不存在

---

### DELETE `/api/v1/notification-channels/:id`

| 项 | 内容 |
| --- | --- |
| **说明** | 删除通知渠道。 |
| **Auth** | Admin only |

**Response** `204 No Content`

**Status codes** | `204`/`200` 成功 | `401` 未认证 | `403` 非管理员 | `404` 不存在

---

### POST `/api/v1/notification-channels/:id/test`

| 项 | 内容 |
| --- | --- |
| **说明** | 向该渠道发送测试消息以验证 Webhook。 |
| **Auth** | Admin only |

**Response** `200 OK`

```json
{
  "success": true,
  "message": "测试消息已发送"
}
```

**Status codes** | `200` 成功 | `401` 未认证 | `403` 非管理员 | `404` 不存在 | `502` 上游失败

---

## WebSocket 实时推送

连接时需携带 JWT：查询参数 `token=<JWT>`（或实现约定的 `access_token`），**不得**将长期令牌写入可被日志采集的公开 URL 日志；生产环境应优先短期令牌或子协议协商方案（实现可扩展）。

### `WS /ws/builds/:id/log`

| 项 | 内容 |
| --- | --- |
| **说明** | 构建日志实时流式推送（文本帧或 JSON 行）。 |
| **Auth** | JWT required（Query: `?token=...`） |

**服务端消息示例**

```json
{
  "type": "log",
  "chunk": "[build] Step 2/5 : RUN npm ci\n",
  "seq": 42
}
```

**关闭** | 构建结束或客户端断开 | 错误时可发 `type: "error"` 后关闭

---

### `WS /ws/deployments/:id/progress`

| 项 | 内容 |
| --- | --- |
| **说明** | 发布进度与阶段事件实时推送。 |
| **Auth** | JWT required（Query: `?token=...`） |

**服务端消息示例**

```json
{
  "type": "progress",
  "phase": "rolling_update",
  "ready_replicas": 2,
  "desired_replicas": 3,
  "message": "Waiting for rollout"
}
```

**关闭** | 发布终态（成功/失败）或断开

---

## 修订记录

| 版本 | 日期 | 说明 |
| --- | --- | --- |
| 0.1.0 | 2026-04-03 | 初稿：与 DeployHub 001 规格对齐的 REST/WebSocket 合约 |
