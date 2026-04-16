# DeployHub 数据模型（PostgreSQL + GORM）

本文档描述 DeployHub 的 **15** 个领域实体（用户所述「14」与下列清单不一致时，以本清单为准）、字段、关系、校验、状态机与索引。数据库为 **PostgreSQL**；敏感字段经 **AES-256-GCM** 加密后以 **Base64 编码** 存入 `text` 列；时间戳为 `time.Time`；主键为自增 `uint`；JSON 列使用 `gorm.io/datatypes` 的 `datatypes.JSON` 映射为 `jsonb`。

---

## 1. User（用户）

**说明**：平台登录用户，支持本地密码或 OAuth；与角色、状态及服务成员关系关联。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| username | varchar(50) | unique，not null | 登录名 |
| email | varchar(100) | unique，not null | 邮箱 |
| password_hash | varchar(255) | null | 密码哈希；OAuth 专用用户可为 null |
| oauth_provider | varchar(20) | null | github / gitlab / google |
| oauth_id | varchar(100) | null | 第三方用户 ID |
| role | varchar(10) | not null | admin / member |
| avatar | varchar(500) | | 头像 URL |
| status | varchar(10) | not null，默认 active | active / disabled |
| created_at | timestamptz | not null | 创建时间 |
| updated_at | timestamptz | not null | 更新时间 |

### Go 模型（GORM）

```go
type User struct {
	ID            uint   `gorm:"primaryKey"`
	Username      string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Email         string `gorm:"type:varchar(100);uniqueIndex;not null"`
	PasswordHash  *string `gorm:"type:varchar(255)"`
	OAuthProvider *string `gorm:"type:varchar(20)"`
	OAuthID       *string `gorm:"type:varchar(100)"`
	Role          string `gorm:"type:varchar(10);not null"`
	Avatar        string `gorm:"type:varchar(500)"`
	Status        string `gorm:"type:varchar(10);not null;default:active"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
```

> **部分唯一索引**：`(oauth_provider, oauth_id)` 在两者均非 null 时唯一。GORM 需在迁移中用原始 SQL 或 `CREATE UNIQUE INDEX ... WHERE oauth_provider IS NOT NULL AND oauth_id IS NOT NULL`，见下文索引节。

### 关系

- **Has many**：`ServiceMember`、`Service`（`owner_id`）、`Build`（`trigger_user_id`）、`Deployment`（`trigger_user_id`）、`Approval`（`requester_id` / `approver_id`）、`Notification`、`AuditLog`、`ConfigVersion`（`created_by_id`）、`ConfigDeployment`（`deployed_by_id`）

### 校验规则

- `username`、`email`：非空；`email` 格式校验。
- `role` ∈ {`admin`, `member`}。
- `status` ∈ {`active`, `disabled`}。
- `oauth_provider` ∈ {`github`, `gitlab`, `google`} 或为空。
- 若 `password_hash` 为空，则必须存在有效的 `(oauth_provider, oauth_id)` 组合（OAuth 用户）。
- 若 `password_hash` 非空，本地登录路径需校验密码策略（长度、复杂度等，应用层定义）。

### 状态迁移

无业务状态机（`status` 为账号启用/停用，由管理操作切换）。

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| unique | username | 唯一 |
| unique | email | 唯一 |
| unique (partial) | oauth_provider, oauth_id | `WHERE oauth_provider IS NOT NULL AND oauth_id IS NOT NULL` |

---

## 2. ServiceMember（服务成员）

**说明**：用户与服务的多对多成员关系及在服务内的角色。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| service_id | uint | FK → Service，not null | 服务 |
| user_id | uint | FK → User，not null | 用户 |
| role | varchar(10) | not null | owner / developer / viewer |
| created_at | timestamptz | not null | 加入时间 |

### Go 模型（GORM）

```go
type ServiceMember struct {
	ID        uint      `gorm:"primaryKey"`
	ServiceID uint      `gorm:"not null;index"`
	UserID    uint      `gorm:"not null;index"`
	Role      string    `gorm:"type:varchar(10);not null"`
	CreatedAt time.Time

	Service Service `gorm:"foreignKey:ServiceID"`
	User    User    `gorm:"foreignKey:UserID"`
}
```

### 关系

- **Belongs to**：`Service`、`User`

### 校验规则

- `role` ∈ {`owner`, `developer`, `viewer`}。
- 同一 `(service_id, user_id)` 仅一条记录。

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| unique | service_id, user_id | 联合唯一 |
| index | service_id | 查询服务下成员 |
| index | user_id | 查询用户所属服务 |

---

## 3. Cluster（Kubernetes 集群）

**说明**：纳管的 K8s 集群；`kubeconfig` 等敏感连接信息加密存储。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| name | varchar(50) | unique，not null | 内部标识名 |
| display_name | varchar(100) | | 展示名 |
| env | varchar(10) | not null | dev / staging / prod |
| api_server | varchar(500) | | API Server 地址（展示/校验用，可选） |
| kubeconfig_encrypted | text | not null | AES-256-GCM + Base64 |
| status | varchar(10) | not null | active / inactive |
| created_at | timestamptz | not null | |
| updated_at | timestamptz | not null | |

### Go 模型（GORM）

```go
type Cluster struct {
	ID                  uint   `gorm:"primaryKey"`
	Name                string `gorm:"type:varchar(50);uniqueIndex;not null"`
	DisplayName         string `gorm:"type:varchar(100)"`
	Env                 string `gorm:"type:varchar(10);not null"`
	APIServer           string `gorm:"type:varchar(500)"`
	KubeconfigEncrypted string `gorm:"type:text;not null"` // 应用层 AES-256-GCM，存 Base64
	Status              string `gorm:"type:varchar(10);not null"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
```

### 关系

- **Has many**：`Service`、`Build`（`build_cluster_id`）、`Deployment`、`ConfigEnvValue`、`ConfigVersion`、`ConfigDeployment`

### 校验规则

- `name`：非空，符合 DNS 子域/内部命名规范（应用层）。
- `env` ∈ {`dev`, `staging`, `prod`}。
- `status` ∈ {`active`, `inactive`}。
- `kubeconfig_encrypted`：写入前加密，禁止明文落库。

### 状态迁移

```
active  ──(手动或 kubeconfig 过期自动)──►  inactive
inactive ─────────(重新校验/更新凭据)────────►  active
```

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| unique | name | 唯一 |

---

## 4. GitRepo（Git 仓库）

**说明**：构建拉代码所用的仓库及认证信息（加密）。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| name | varchar(100) | unique，not null | 仓库显示名/标识 |
| url | varchar(500) | not null | 克隆 URL |
| provider | varchar(20) | not null | github / gitlab / other |
| auth_type | varchar(10) | not null | token / ssh_key |
| credential_encrypted | text | not null | AES-256-GCM + Base64（token 或私钥等） |
| created_at | timestamptz | not null | |
| updated_at | timestamptz | not null | |

### Go 模型（GORM）

```go
type GitRepo struct {
	ID                    uint   `gorm:"primaryKey"`
	Name                  string `gorm:"type:varchar(100);uniqueIndex;not null"`
	URL                   string `gorm:"type:varchar(500);not null"`
	Provider              string `gorm:"type:varchar(20);not null"`
	AuthType              string `gorm:"type:varchar(10);not null"`
	CredentialEncrypted   string `gorm:"type:text;not null"` // AES-256-GCM
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
```

### 关系

- **Has many**：`Service`

### 校验规则

- `provider` ∈ {`github`, `gitlab`, `other`}。
- `auth_type` ∈ {`token`, `ssh_key`}。
- `url`：合法 Git URL（https/ssh）。
- `credential_encrypted`：仅密文存储。

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| unique | name | 唯一 |

---

## 5. Registry（镜像仓库）

**说明**：镜像推送/拉取的注册中心配置；认证 JSON 加密存储。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| name | varchar(100) | unique，not null | 名称 |
| url | varchar(500) | not null | 仓库根地址 |
| provider | varchar(20) | not null | ecr / acr / tcr / harbor / other |
| auth_config_encrypted | text | not null | AES-256-GCM + Base64，明文为 JSON |
| is_default | bool | not null，default false | 是否默认仓库 |
| created_at | timestamptz | not null | |
| updated_at | timestamptz | not null | |

### Go 模型（GORM）

```go
type Registry struct {
	ID                   uint   `gorm:"primaryKey"`
	Name                 string `gorm:"type:varchar(100);uniqueIndex;not null"`
	URL                  string `gorm:"type:varchar(500);not null"`
	Provider             string `gorm:"type:varchar(20);not null"`
	AuthConfigEncrypted  string `gorm:"type:text;not null"` // JSON 结构加密后 Base64
	IsDefault            bool   `gorm:"not null;default:false"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
```

### 关系

- **Has many**：`Service`

### 校验规则

- `provider` ∈ {`ecr`, `acr`, `tcr`, `harbor`, `other`}。
- 同一租户/系统内「默认仓库」至多一个（应用层或部分唯一索引，视多租户设计而定）。

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| unique | name | 唯一 |

---

## 6. Service（服务）

**说明**：平台核心实体，关联代码仓库、镜像仓库、目标集群及资源与运行配置。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| name | varchar(100) | unique，not null | 服务标识 |
| display_name | varchar(200) | | 展示名 |
| description | text | | 描述 |
| git_repo_id | uint | FK → GitRepo，not null | |
| git_branch | varchar(200) | not null，default main | 默认分支 |
| dockerfile_path | varchar(500) | not null，default ./Dockerfile | |
| registry_id | uint | FK → Registry，not null | |
| image_repo | varchar(500) | not null | 完整镜像仓库路径 |
| cluster_id | uint | FK → Cluster，not null | 部署目标集群 |
| namespace | varchar(100) | not null，default default | K8s 命名空间 |
| replicas | int | not null，default 1 | |
| cpu_request | varchar(20) | | K8s 资源字符串 |
| mem_request | varchar(20) | | |
| cpu_limit | varchar(20) | | |
| mem_limit | varchar(20) | | |
| port | int | not null | 容器端口 |
| health_check_path | varchar(200) | | 健康检查 HTTP 路径 |
| env_vars | jsonb | | 环境变量（结构由应用约定） |
| volumes | jsonb | | 卷配置 |
| owner_id | uint | FK → User，not null | 创建者/默认负责人 |
| created_at | timestamptz | not null | |
| updated_at | timestamptz | not null | |

### Go 模型（GORM）

```go
type Service struct {
	ID              uint           `gorm:"primaryKey"`
	Name            string         `gorm:"type:varchar(100);uniqueIndex;not null"`
	DisplayName     string         `gorm:"type:varchar(200)"`
	Description     string         `gorm:"type:text"`
	GitRepoID       uint           `gorm:"not null;index"`
	GitBranch       string         `gorm:"type:varchar(200);not null;default:main"`
	DockerfilePath  string         `gorm:"type:varchar(500);not null;default:./Dockerfile"`
	RegistryID      uint           `gorm:"not null;index"`
	ImageRepo       string         `gorm:"type:varchar(500);not null"`
	ClusterID       uint           `gorm:"not null;index:idx_service_cluster_namespace"`
	Namespace       string         `gorm:"type:varchar(100);not null;default:default;index:idx_service_cluster_namespace"`
	Replicas        int            `gorm:"not null;default:1"`
	CPURequest      string         `gorm:"type:varchar(20)"`
	MemRequest      string         `gorm:"type:varchar(20)"`
	CPULimit        string         `gorm:"type:varchar(20)"`
	MemLimit        string         `gorm:"type:varchar(20)"`
	Port            int            `gorm:"not null"`
	HealthCheckPath string         `gorm:"type:varchar(200)"`
	EnvVars         datatypes.JSON `gorm:"type:jsonb"`
	Volumes         datatypes.JSON `gorm:"type:jsonb"`
	OwnerID         uint           `gorm:"not null;index"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	GitRepo  GitRepo  `gorm:"foreignKey:GitRepoID"`
	Registry Registry `gorm:"foreignKey:RegistryID"`
	Cluster  Cluster  `gorm:"foreignKey:ClusterID"`
	Owner    User     `gorm:"foreignKey:OwnerID"`
}
```

### 关系

- **Belongs to**：`GitRepo`、`Registry`、`Cluster`、`User`（owner）
- **Has many**：`ServiceMember`、`Build`、`Deployment`、`ConfigTemplate`

### 校验规则

- `name`：唯一、非空；`replicas` ≥ 1；`port` ∈ 合法端口范围。
- `env_vars` / `volumes`：JSON Schema 校验（应用层）。
- 资源字段符合 K8s quantity 格式（可选严格校验）。

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| unique | name | 唯一 |
| index | cluster_id, namespace | 联合索引 `idx_service_cluster_namespace` |
| index | git_repo_id, registry_id, owner_id | 外键查询 |

---

## 7. Build（构建）

**说明**：针对某服务的镜像构建任务及日志；在指定构建集群执行 Kaniko 等。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| service_id | uint | FK → Service，not null | |
| trigger_user_id | uint | FK → User，not null | 触发人 |
| git_branch | varchar(200) | not null | 构建所用分支 |
| git_commit | varchar(40) | | commit SHA |
| image_tag | varchar(200) | | 产出镜像标签 |
| status | varchar(20) | not null | 见状态机 |
| build_cluster_id | uint | FK → Cluster，not null | 执行构建的集群 |
| kaniko_job_name | varchar(200) | | K8s Job 名 |
| log | text | | 构建日志全文 |
| started_at | timestamptz | null | 开始时间 |
| finished_at | timestamptz | null | 结束时间 |
| created_at | timestamptz | not null | 创建时间 |

### Go 模型（GORM）

```go
type Build struct {
	ID              uint       `gorm:"primaryKey"`
	ServiceID       uint       `gorm:"not null;index"`
	TriggerUserID   uint       `gorm:"not null;index"`
	GitBranch       string     `gorm:"type:varchar(200);not null"`
	GitCommit       string     `gorm:"type:varchar(40)"`
	ImageTag        string     `gorm:"type:varchar(200)"`
	Status          string     `gorm:"type:varchar(20);not null"`
	BuildClusterID  uint       `gorm:"not null;index"`
	KanikoJobName   string     `gorm:"type:varchar(200)"`
	Log             string     `gorm:"type:text"`
	StartedAt       *time.Time
	FinishedAt      *time.Time
	CreatedAt       time.Time

	Service      Service `gorm:"foreignKey:ServiceID"`
	TriggerUser  User    `gorm:"foreignKey:TriggerUserID"`
	BuildCluster Cluster `gorm:"foreignKey:BuildClusterID"`
}
```

### 关系

- **Belongs to**：`Service`、`User`（trigger）、`Cluster`（build）
- **Has many**：`Deployment`（可选关联 `build_id`）

### 校验规则

- `status` ∈ {`pending`, `building`, `success`, `failed`, `cancelled`}。
- `finished_at` ≥ `started_at`（若均非空）。

### 状态迁移

```
                    ┌──► success
pending ──► building ──┤
                    └──► failed

pending ──► cancelled
building ──► cancelled
```

（`cancelled` 仅能从 `pending` 或 `building` 进入。）

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| index | service_id | 按服务查构建历史 |
| index | trigger_user_id, build_cluster_id | 过滤 |

---

## 8. Deployment（部署/发布）

**说明**：一次发布或回滚记录，含审批门禁与集群上的镜像版本。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| service_id | uint | FK → Service，not null | |
| build_id | uint | FK → Build，null | 回滚等场景可不关联新构建 |
| trigger_user_id | uint | FK → User，not null | |
| cluster_id | uint | FK → Cluster，not null | |
| namespace | varchar(100) | not null | |
| image_tag | varchar(200) | not null | 目标镜像 |
| replicas | int | not null | |
| status | varchar(20) | not null | 见状态机 |
| previous_image_tag | varchar(200) | | 回滚前镜像，供回滚 |
| is_rollback | bool | not null，default false | 是否回滚操作 |
| rollback_from_id | uint | FK → Deployment，null | 回滚来源部署 |
| started_at | timestamptz | null | |
| finished_at | timestamptz | null | |
| created_at | timestamptz | not null | |

### Go 模型（GORM）

```go
type Deployment struct {
	ID                 uint       `gorm:"primaryKey"`
	ServiceID          uint       `gorm:"not null;index"`
	BuildID            *uint      `gorm:"index"`
	TriggerUserID      uint       `gorm:"not null;index"`
	ClusterID          uint       `gorm:"not null;index"`
	Namespace          string     `gorm:"type:varchar(100);not null"`
	ImageTag           string     `gorm:"type:varchar(200);not null"`
	Replicas           int        `gorm:"not null"`
	Status             string     `gorm:"type:varchar(20);not null"`
	PreviousImageTag   string     `gorm:"type:varchar(200)"`
	IsRollback         bool       `gorm:"not null;default:false"`
	RollbackFromID     *uint      `gorm:"index"`
	StartedAt          *time.Time
	FinishedAt         *time.Time
	CreatedAt          time.Time

	Service       Service     `gorm:"foreignKey:ServiceID"`
	Build         *Build      `gorm:"foreignKey:BuildID"`
	TriggerUser   User        `gorm:"foreignKey:TriggerUserID"`
	Cluster       Cluster     `gorm:"foreignKey:ClusterID"`
	RollbackFrom  *Deployment `gorm:"foreignKey:RollbackFromID"`
}
```

### 关系

- **Belongs to**：`Service`、`Build`（可选）、`User`、`Cluster`、`Deployment`（rollback 来源）
- **Has many**：`Approval`；子部署可通过 `rollback_from_id` 指回父记录

### 校验规则

- `status` ∈ {`pending_approval`, `approved`, `deploying`, `success`, `failed`, `rolled_back`, `rejected`, `expired`}；`replicas` ≥ 1。
- `rollback_from_id` 仅在 `is_rollback == true` 时有意义（应用层一致）。

### 状态迁移

允许的状态：`pending_approval`、`approved`、`deploying`、`success`、`failed`、`rolled_back`、`rejected`、`expired`。

```
pending_approval ──► approved | rejected | expired

approved ──► deploying ──► success | failed

success ──► rolled_back   （例如：该次发布被后续回滚操作标记为已回滚；或通过业务规则更新本条记录）
```

说明：`rolled_back` 也可仅体现在**新创建的**回滚部署记录上（`is_rollback=true`），历史成功记录是否改为 `rolled_back` 由产品规则决定，上图为常见一种。

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| index | service_id, cluster_id, trigger_user_id | 列表与过滤 |
| index | build_id, rollback_from_id | 可选 |

---

## 9. Approval（发布审批）

**说明**：针对某次部署的审批流；审批人为服务 Owner（或业务规则扩展）。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| deployment_id | uint | FK → Deployment，not null | |
| requester_id | uint | FK → User，not null | 发起人 |
| approver_id | uint | FK → User，not null | 审批人（Owner） |
| status | varchar(10) | not null | pending / approved / rejected |
| comment | text | | 审批意见 |
| decided_at | timestamptz | null | 裁决时间 |
| created_at | timestamptz | not null | |

### Go 模型（GORM）

```go
type Approval struct {
	ID             uint       `gorm:"primaryKey"`
	DeploymentID   uint       `gorm:"not null;index"`
	RequesterID    uint       `gorm:"not null;index"`
	ApproverID     uint       `gorm:"not null;index"`
	Status         string     `gorm:"type:varchar(10);not null"`
	Comment        string     `gorm:"type:text"`
	DecidedAt      *time.Time
	CreatedAt      time.Time

	Deployment Deployment `gorm:"foreignKey:DeploymentID"`
	Requester  User       `gorm:"foreignKey:RequesterID"`
	Approver   User       `gorm:"foreignKey:ApproverID"`
}
```

### 关系

- **Belongs to**：`Deployment`、`User`（requester）、`User`（approver）

### 校验规则

- `status` ∈ {`pending`, `approved`, `rejected`}。
- `decided_at` 在 `approved`/`rejected` 时必填。

### 状态迁移

```
pending ──► approved
pending ──► rejected
```

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| index | deployment_id | 按部署查审批 |

---

## 10. ConfigTemplate（配置模板）

**说明**：服务级配置模板（ConfigMap/Secret 内容模板，Go template 语法）。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| service_id | uint | FK → Service，not null | |
| name | varchar(100) | not null | 模板名 |
| config_type | varchar(10) | not null | configmap / secret |
| template_content | text | not null | Go template 原文 |
| created_at | timestamptz | not null | |
| updated_at | timestamptz | not null | |

### Go 模型（GORM）

```go
type ConfigTemplate struct {
	ID              uint   `gorm:"primaryKey"`
	ServiceID       uint   `gorm:"not null;index:idx_config_template_service_name"`
	Name            string `gorm:"type:varchar(100);not null;index:idx_config_template_service_name"`
	ConfigType      string `gorm:"type:varchar(10);not null"`
	TemplateContent string `gorm:"type:text;not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	Service Service `gorm:"foreignKey:ServiceID"`
}
```

### 关系

- **Belongs to**：`Service`
- **Has many**：`ConfigEnvValue`、`ConfigVersion`

### 校验规则

- `config_type` ∈ {`configmap`, `secret`}。
- `template_content`：模板语法可解析。

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| unique | service_id, name | 联合唯一 `idx_config_template_service_name` |

---

## 11. ConfigEnvValue（按集群的环境变量密文）

**说明**：某模板在指定集群下的变量值（加密 JSON）。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| config_template_id | uint | FK → ConfigTemplate，not null | |
| cluster_id | uint | FK → Cluster，not null | |
| values_encrypted | text | not null | AES-256-GCM + Base64，明文为 key-value JSON |
| created_at | timestamptz | not null | |
| updated_at | timestamptz | not null | |

### Go 模型（GORM）

```go
type ConfigEnvValue struct {
	ID                 uint   `gorm:"primaryKey"`
	ConfigTemplateID   uint   `gorm:"not null;index:idx_config_env_tpl_cluster"`
	ClusterID          uint   `gorm:"not null;index:idx_config_env_tpl_cluster"`
	ValuesEncrypted    string `gorm:"type:text;not null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	ConfigTemplate ConfigTemplate `gorm:"foreignKey:ConfigTemplateID"`
	Cluster        Cluster        `gorm:"foreignKey:ClusterID"`
}
```

### 关系

- **Belongs to**：`ConfigTemplate`、`Cluster`

### 校验规则

- 解密后的 JSON 结构与模板变量一致（应用层）。

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| unique | config_template_id, cluster_id | `idx_config_env_tpl_cluster` |

---

## 12. ConfigVersion（配置渲染版本）

**说明**：某模板在某集群上每次渲染下发的版本记录。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| config_template_id | uint | FK → ConfigTemplate，not null | |
| cluster_id | uint | FK → Cluster，not null | |
| version | int | not null | 在 (template, cluster) 维度递增 |
| rendered_content | text | not null | 渲染后完整内容 |
| created_by_id | uint | FK → User，not null | |
| created_at | timestamptz | not null | |

### Go 模型（GORM）

```go
type ConfigVersion struct {
	ID               uint   `gorm:"primaryKey"`
	ConfigTemplateID uint   `gorm:"not null;uniqueIndex:idx_config_version_tpl_cluster_ver"`
	ClusterID        uint   `gorm:"not null;uniqueIndex:idx_config_version_tpl_cluster_ver"`
	Version          int    `gorm:"not null;uniqueIndex:idx_config_version_tpl_cluster_ver"`
	RenderedContent  string `gorm:"type:text;not null"`
	CreatedByID      uint   `gorm:"not null;index"`
	CreatedAt        time.Time

	ConfigTemplate ConfigTemplate `gorm:"foreignKey:ConfigTemplateID"`
	Cluster        Cluster        `gorm:"foreignKey:ClusterID"`
	CreatedBy      User           `gorm:"foreignKey:CreatedByID"`
}
```

### 关系

- **Belongs to**：`ConfigTemplate`、`Cluster`、`User`
- **Has many**：`ConfigDeployment`

### 校验规则

- `(config_template_id, cluster_id, version)` 唯一；`version` 在插入时由事务内 `MAX(version)+1` 或序列保证。

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| unique | config_template_id, cluster_id, version | 版本号在模板+集群下唯一 |

---

## 13. ConfigDeployment（配置下发记录）

**说明**：将某一 `ConfigVersion` 应用到集群中具体资源的一次操作记录。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| config_version_id | uint | FK → ConfigVersion，not null | |
| cluster_id | uint | FK → Cluster，not null | |
| namespace | varchar(100) | not null | 目标命名空间 |
| resource_name | varchar(200) | not null | ConfigMap/Secret 名称 |
| status | varchar(10) | not null | pending / success / failed |
| deployed_by_id | uint | FK → User，not null | |
| deployed_at | timestamptz | null | 实际完成时间 |
| created_at | timestamptz | not null | |

### Go 模型（GORM）

```go
type ConfigDeployment struct {
	ID               uint       `gorm:"primaryKey"`
	ConfigVersionID  uint       `gorm:"not null;index"`
	ClusterID        uint       `gorm:"not null;index"`
	Namespace        string     `gorm:"type:varchar(100);not null"`
	ResourceName     string     `gorm:"type:varchar(200);not null"`
	Status           string     `gorm:"type:varchar(10);not null"`
	DeployedByID     uint       `gorm:"not null;index"`
	DeployedAt       *time.Time
	CreatedAt        time.Time

	ConfigVersion ConfigVersion `gorm:"foreignKey:ConfigVersionID"`
	Cluster       Cluster       `gorm:"foreignKey:ClusterID"`
	DeployedBy    User          `gorm:"foreignKey:DeployedByID"`
}
```

### 关系

- **Belongs to**：`ConfigVersion`、`Cluster`、`User`

### 校验规则

- `status` ∈ {`pending`, `success`, `failed`}。
- `deployed_at` 在终态成功/失败时建议非空。

### 状态迁移

```
pending ──► success | failed
```

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| index | config_version_id, cluster_id, deployed_by_id | |

---

## 14. Notification（站内通知）

**说明**：用户维度的消息，可关联业务对象（审批、构建等）。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| user_id | uint | FK → User，not null | 接收人 |
| type | varchar(30) | not null | 见枚举 |
| title | varchar(200) | not null | 标题 |
| content | text | | 正文 |
| is_read | bool | not null，default false | 是否已读 |
| reference_type | varchar(30) | | 多态关联类型名 |
| reference_id | uint | | 多态关联 ID |
| created_at | timestamptz | not null | |

### Go 模型（GORM）

```go
type Notification struct {
	ID             uint   `gorm:"primaryKey"`
	UserID         uint   `gorm:"not null;index"`
	Type           string `gorm:"type:varchar(30);not null"`
	Title          string `gorm:"type:varchar(200);not null"`
	Content        string `gorm:"type:text"`
	IsRead         bool   `gorm:"not null;default:false"`
	ReferenceType  string `gorm:"type:varchar(30)"`
	ReferenceID    uint   `gorm:"default:0"`
	CreatedAt      time.Time

	User User `gorm:"foreignKey:UserID"`
}
```

### 关系

- **Belongs to**：`User`

### 校验规则

- `type` ∈ {`approval_request`, `build_complete`, `deploy_result`, `config_deployed`, ...}。
- `reference_type` / `reference_id` 成对出现或均为空（应用层）。

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| index | user_id | 用户收件箱 |
| index | user_id, is_read | 未读筛选（可选复合） |

---

## 15. AuditLog（审计日志）

**说明**：关键操作留痕；`detail` 为脱敏后的 JSON。

### 字段表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | uint | PK，自增 | 主键 |
| user_id | uint | FK → User，not null | 操作者 |
| action | varchar(50) | not null | create_service / trigger_build / deploy / rollback / approve / reject / update_config / ... |
| resource_type | varchar(30) | | 资源类型 |
| resource_id | uint | | 资源 ID |
| detail | jsonb | | 脱敏后的操作详情 |
| ip_address | varchar(45) | | IPv4/IPv6 |
| created_at | timestamptz | not null | |

### Go 模型（GORM）

```go
type AuditLog struct {
	ID             uint           `gorm:"primaryKey"`
	UserID         uint           `gorm:"not null;index:idx_audit_user"`
	Action         string         `gorm:"type:varchar(50);not null"`
	ResourceType   string         `gorm:"type:varchar(30);index:idx_audit_resource"`
	ResourceID     uint           `gorm:"index:idx_audit_resource"`
	Detail         datatypes.JSON `gorm:"type:jsonb"`
	IPAddress      string         `gorm:"type:varchar(45)"`
	CreatedAt      time.Time      `gorm:"index:idx_audit_created"`

	User User `gorm:"foreignKey:UserID"`
}
```

### 关系

- **Belongs to**：`User`

### 校验规则

- `detail` 禁止含明文密钥、token、完整个人信息（写入前脱敏）。

### 索引

| 索引 | 列 | 说明 |
|------|-----|------|
| index | created_at | `idx_audit_created`，按时间范围查询 |
| index | user_id | `idx_audit_user` |
| index | resource_type, resource_id | `idx_audit_resource`，按资源追溯 |

---

## 附录：敏感字段汇总

| 实体 | 字段 | 存储类型 | 说明 |
|------|------|----------|------|
| Cluster | kubeconfig_encrypted | text | AES-256-GCM + Base64 |
| GitRepo | credential_encrypted | text | 同上 |
| Registry | auth_config_encrypted | text | JSON 加密后 Base64 |
| ConfigEnvValue | values_encrypted | text | KV JSON 加密后 Base64 |

密钥材料须来自 KMS/环境变量/密钥管理服务，**禁止**硬编码或写入仓库。

---

## 附录：PostgreSQL 部分唯一索引示例（User OAuth）

```sql
CREATE UNIQUE INDEX idx_user_oauth_provider_id
ON users (oauth_provider, oauth_id)
WHERE oauth_provider IS NOT NULL AND oauth_id IS NOT NULL;
```

表名以实际 GORM 默认复数表名为准（如 `users`）。
