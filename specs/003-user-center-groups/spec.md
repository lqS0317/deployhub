# 003 - 用户中心：组管理与权限增强

## 1. Overview

为 DeployHub 引入「组（Group）」概念作为用户批量权限管理单元，激活已实现但未挂载的 ServiceRBAC 中间件，增强管理员用户管理能力。

## 2. Motivation

| 痛点 | 影响 |
|------|------|
| 无组概念，Service 权限逐人分配 | 团队 10+ 人时管理成本高 |
| ServiceRBAC 中间件已实现但未挂载 | Service 级权限未生效，任何已认证用户可操作任何 Service |
| 管理员无法创建用户 | 只能等用户自注册后再改角色 |
| 无法查看用户有效权限全景 | 排查权限问题困难 |

## 3. User Stories

- **S1**: Admin 创建组「后端团队」，添加 5 名成员，为该组分配 3 个 Service 的 developer 权限 → 5 人自动获得这些 Service 的 developer 权限
- **S2**: Admin 查看用户 alice 的权限全景 → 看到她在 ServiceA 有个人 owner 权限，在 ServiceB/C 通过「后端团队」获得 developer 权限
- **S3**: 普通用户尝试操作未授权的 Service → 返回 403
- **S4**: Admin 创建新用户 bob（username, email, password, role=member）→ bob 可以登录

## 4. Edge Cases

- 用户同时有个人权限和组权限 → 取最高值
- 用户属于多个组且对同一 Service 有不同权限 → 取最高值
- 删除组 → 级联删除 GroupMember + GroupServicePermission
- Admin 用户 → 跳过所有权限检查，视为 owner
- 组内添加已存在的成员 → 409 Conflict
- 从组移除不存在的成员 → 404

## 5. Functional Requirements

### 5.1 数据模型（3 个新表）

**groups**

| 字段 | 类型 | 约束 |
|------|------|------|
| id | uint | PK |
| name | varchar(100) | unique, not null |
| description | text | |
| created_by | uint | FK → users.id, not null |
| created_at | timestamp | |
| updated_at | timestamp | |

**group_members**

| 字段 | 类型 | 约束 |
|------|------|------|
| id | uint | PK |
| group_id | uint | FK → groups.id, not null |
| user_id | uint | FK → users.id, not null |
| created_at | timestamp | |

唯一索引: `(group_id, user_id)`

**group_service_permissions**

| 字段 | 类型 | 约束 |
|------|------|------|
| id | uint | PK |
| group_id | uint | FK → groups.id, not null |
| service_id | uint | FK → services.id, not null |
| role | varchar(10) | not null (viewer/developer/owner) |
| created_at | timestamp | |

唯一索引: `(group_id, service_id)`

### 5.2 核心权限计算

```
GetEffectiveRole(userID, serviceID) → (role string, sources []string)

1. User.Role == "admin" → return ("owner", ["admin"])
2. personal = ServiceMember.FindRole(serviceID, userID)
3. groupRoles = GroupServicePermission.FindByUserAndService(userID, serviceID)
4. effectiveRole = max(personal, max(groupRoles...))
5. return (effectiveRole, sources)  // sources 标注来源：个人 / 组名
```

缓存策略：Redis key `perm:{userID}:{serviceID}`, TTL 5 分钟，权限变更时清除 `perm:{userID}:*`。

### 5.3 API 契约

#### 组管理（Admin only）

| Method | Path | Body | Response |
|--------|------|------|----------|
| GET | /api/v1/groups | — | `{ items: Group[], total }` |
| POST | /api/v1/groups | `{ name, description }` | `201 Group` |
| GET | /api/v1/groups/:id | — | `Group` (含 member_count, permission_count) |
| PUT | /api/v1/groups/:id | `{ name?, description? }` | `Group` |
| DELETE | /api/v1/groups/:id | — | `204` |

#### 组成员（Admin only）

| Method | Path | Body | Response |
|--------|------|------|----------|
| GET | /api/v1/groups/:id/members | — | `{ items: GroupMember[] }` |
| POST | /api/v1/groups/:id/members | `{ user_ids: uint[] }` | `201 GroupMember[]` |
| DELETE | /api/v1/groups/:id/members/:user_id | — | `204` |

#### 组 Service 权限（Admin only）

| Method | Path | Body | Response |
|--------|------|------|----------|
| GET | /api/v1/groups/:id/permissions | — | `{ items: GroupServicePermission[] }` |
| POST | /api/v1/groups/:id/permissions | `{ service_id, role }` | `201 GroupServicePermission` |
| PUT | /api/v1/groups/:id/permissions/:pid | `{ role }` | `GroupServicePermission` |
| DELETE | /api/v1/groups/:id/permissions/:pid | — | `204` |

#### 权限总览

| Method | Path | Body | Response |
|--------|------|------|----------|
| GET | /api/v1/users/:id/permissions | — | `{ items: EffectivePermission[] }` (Admin only) |
| GET | /api/v1/auth/my-permissions | — | `{ items: EffectivePermission[] }` |

**EffectivePermission** 结构:
```json
{
  "service_id": 1,
  "service_name": "payment-api",
  "role": "developer",
  "sources": [
    { "type": "personal", "name": "个人权限" },
    { "type": "group", "name": "后端团队", "group_id": 3 }
  ]
}
```

#### Admin 创建用户

| Method | Path | Body | Response |
|--------|------|------|----------|
| POST | /api/v1/users | `{ username, email, password, role }` | `201 User` |

### 5.4 ServiceRBAC 中间件挂载

挂载到以下路由组（已认证 + 需 Service 权限的路由）：

| 路由组 | 需要参数 | 最低权限 |
|--------|---------|---------|
| GET /services/:id | :id → serviceID | viewer |
| PUT /services/:id | :id → serviceID | developer |
| DELETE /services/:id | :id → serviceID | owner |
| POST /services/:id/members | :id → serviceID | owner |
| POST /builds（body.service_id） | body → serviceID | developer |
| POST /deployments（body.service_id） | body → serviceID | developer |
| GET /builds?service_id=X | query → serviceID | viewer |

中间件改造：`CheckPermission` 内部调用 `GetEffectiveRole` 替代直接查 `ServiceMember`。

### 5.5 前端变更

**系统设置 → 新增「组管理」Tab**
- 组列表页：名称、描述、成员数、Service 数、创建时间、操作（编辑/删除）
- 创建/编辑组对话框：名称 + 描述
- 组详情页（或展开行）：
  - 成员管理区：用户下拉选择 + 添加，列表 + 移除
  - Service 权限区：Service 下拉 + 角色选择 + 添加，列表 + 编辑角色/移除

**用户管理 Tab 增强**
- 新增「创建用户」按钮 + 对话框
- 每行增加「权限」链接 → 弹出权限查看弹窗

**权限查看弹窗**
- 表格展示：Service 名称、有效角色、来源（标签列表）

## 6. Success Criteria

- [ ] Admin 可以 CRUD 组、管理组成员、管理组 Service 权限
- [ ] Admin 可以创建用户
- [ ] GetEffectiveRole 正确合并个人 + 组权限并取最高值
- [ ] ServiceRBAC 中间件挂载并对非 Admin 用户生效
- [ ] 权限查看 API 返回完整权限源
- [ ] 组删除时级联清理关联数据
- [ ] 权限结果 Redis 缓存且变更时清除
- [ ] 所有新操作产生审计日志
- [ ] 全部后端测试通过

## 7. Assumptions

- Redis 连接已可用（现有配置）
- 现有 ServiceMember 数据保持不变，组权限是叠加层
- ServiceRBAC 挂载后，未分配权限的普通用户将无法操作 Service（破坏性变更需通知）
- 前端设置页 layout 已支持 Tab 扩展

## 8. Non-Goals

- 组的层级嵌套（子组）
- Service 批量权限模板
- 权限审批流程
- 组的 owner 概念（组只由 admin 管理）
- OAuth/SSO 组同步
