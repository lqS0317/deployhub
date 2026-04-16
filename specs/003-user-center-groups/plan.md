# 003 - 实现计划

## Layer 1: DB Migrations

### 受影响文件
- **新建** `backend/migrations/000019_create_groups.up.sql`
- **新建** `backend/migrations/000019_create_groups.down.sql`
- **新建** `backend/migrations/000020_create_group_members.up.sql`
- **新建** `backend/migrations/000020_create_group_members.down.sql`
- **新建** `backend/migrations/000021_create_group_service_permissions.up.sql`
- **新建** `backend/migrations/000021_create_group_service_permissions.down.sql`

### 实现要点
- groups: name unique, created_by FK users
- group_members: 唯一索引 (group_id, user_id)
- group_service_permissions: 唯一索引 (group_id, service_id), role varchar(10)

---

## Layer 2: GORM Models

### 受影响文件
- **新建** `backend/internal/model/group.go`
- **新建** `backend/internal/model/group_member.go`
- **新建** `backend/internal/model/group_service_permission.go`

### 实现要点
- Group: ID, Name, Description, CreatedBy, CreatedAt, UpdatedAt, Creator *User (foreignKey)
- GroupMember: ID, GroupID, UserID, CreatedAt, Group *Group, User *User
- GroupServicePermission: ID, GroupID, ServiceID, Role, CreatedAt, Group *Group, Service *Service

---

## Layer 3: Repository Layer

### 受影响文件
- **新建** `backend/internal/repository/group_repo.go` (接口)
- **新建** `backend/internal/repository/group_repo_impl.go` (实现)
- **新建** `backend/internal/repository/group_member_repo.go` (接口)
- **新建** `backend/internal/repository/group_member_repo_impl.go` (实现)
- **新建** `backend/internal/repository/group_permission_repo.go` (接口)
- **新建** `backend/internal/repository/group_permission_repo_impl.go` (实现)

### 接口设计

**GroupRepository:**
- Create, FindByID, FindByName, Update, Delete, List(page, pageSize)

**GroupMemberRepository:**
- Create, Delete(groupID, userID), ListByGroup(groupID), ListByUser(userID), Exists(groupID, userID)

**GroupPermissionRepository:**
- Create, FindByID, Update, Delete, ListByGroup(groupID), FindByGroupAndService(groupID, serviceID), FindByUserAndService(userID, serviceID) (JOIN group_members)

---

## Layer 4: Group Service

### 受影响文件
- **新建** `backend/internal/service/group/group.go`

### 实现要点
- GroupService: groupRepo, memberRepo, permRepo
- CRUD 组：Create(name, desc, createdBy), Update(id, name, desc), Delete(id) 级联删除, List, GetByID
- 成员管理：AddMembers(groupID, userIDs), RemoveMember(groupID, userID), ListMembers(groupID)
- 权限管理：AddPermission(groupID, serviceID, role), UpdatePermission(id, role), RemovePermission(id), ListPermissions(groupID)
- 删除组时级联删除 group_members + group_service_permissions

---

## Layer 5: GetEffectiveRole

### 受影响文件
- **新建** `backend/internal/service/svc/effective_role.go`
- **新建** `backend/internal/service/svc/effective_role_test.go`
- **修改** `backend/internal/service/svc/rbac.go` — 注入 EffectiveRoleService

### 实现要点
- EffectiveRoleService: memberRepo, groupMemberRepo, groupPermRepo, userRepo, redisClient
- GetEffectiveRole(userID, serviceID) → (role, sources[]): 合并个人 + 所有组权限取最高值
- GetAllEffectivePermissions(userID) → []EffectivePermission: 用户权限全景
- Redis 缓存: key `perm:{userID}:{serviceID}`, TTL 5min
- InvalidateCache(userID): 删除 `perm:{userID}:*`
- roleLevel map: viewer=1, developer=2, owner=3
- admin 用户直接返回 owner

---

## Layer 6: ServiceRBAC 中间件改造 + 挂载

### 受影响文件
- **修改** `backend/internal/middleware/rbac.go` — CheckPermission 调用 GetEffectiveRole
- **修改** `backend/internal/service/svc/rbac.go` — 注入 EffectiveRoleService，CheckPermission 委托
- **修改** `backend/cmd/server/main.go` — 实例化 EffectiveRoleService，激活 rbacSvc，挂载中间件

### 挂载位置
- services/:id — GET(viewer), PUT(developer), DELETE(owner)
- services/:id/members — POST(owner)
- builds POST — developer (从 body 取 service_id)
- deployments POST — developer (从 body 取 service_id)

---

## Layer 7: Group Handler

### 受影响文件
- **新建** `backend/internal/handler/group.go`
- **修改** `backend/cmd/server/main.go` — 注册路由

### 端点
- 组 CRUD: GET/POST /groups, GET/PUT/DELETE /groups/:id
- 组成员: GET/POST /groups/:id/members, DELETE /groups/:id/members/:user_id
- 组权限: GET/POST /groups/:id/permissions, PUT/DELETE /groups/:id/permissions/:pid
- 所有路由 AdminOnly

---

## Layer 8: Permission Handler

### 受影响文件
- **新建** `backend/internal/handler/permission.go`
- **修改** `backend/cmd/server/main.go` — 注册路由
- **修改** `backend/internal/handler/auth.go` — 添加 GetMyPermissions 端点

### 端点
- GET /users/:id/permissions — Admin 查看某用户权限全景
- GET /auth/my-permissions — 当前用户查看自己的权限全景

---

## Layer 9: Admin Create User

### 受影响文件
- **修改** `backend/internal/handler/user_admin.go` — 添加 CreateUser handler
- **修改** `backend/internal/service/auth/auth.go` — 复用 Register 逻辑或新增 AdminCreateUser

### 实现要点
- POST /users — Admin 创建用户 (username, email, password, role)
- BCrypt 哈希密码
- 校验用户名/邮箱唯一

---

## Layer 10: Frontend Hooks

### 受影响文件
- **新建** `frontend/src/hooks/use-groups.ts`
- **新建** `frontend/src/hooks/use-permissions.ts`

### 实现要点
- useGroups, useGroup(id), useCreateGroup, useUpdateGroup, useDeleteGroup
- useGroupMembers(id), useAddGroupMembers, useRemoveGroupMember
- useGroupPermissions(id), useAddGroupPermission, useUpdateGroupPermission, useRemoveGroupPermission
- useUserPermissions(userId), useMyPermissions

---

## Layer 11: Frontend UI

### 受影响文件
- **修改** `frontend/src/app/(dashboard)/settings/layout.tsx` — 添加「组管理」Tab
- **新建** `frontend/src/app/(dashboard)/settings/groups/page.tsx` — 组列表页
- **新建** `frontend/src/components/group/group-form-dialog.tsx` — 创建/编辑组
- **新建** `frontend/src/components/group/group-detail-panel.tsx` — 组详情（成员+权限管理）
- **修改** `frontend/src/app/(dashboard)/settings/users/page.tsx` — 创建用户按钮 + 权限查看
- **新建** `frontend/src/components/user/create-user-dialog.tsx` — 创建用户对话框
- **新建** `frontend/src/components/user/permission-dialog.tsx` — 权限查看弹窗
- **修改** `frontend/src/types/index.ts` — 新增 Group, GroupMember, GroupServicePermission, EffectivePermission 类型

---

## Layer 12: Tests & Verification

### 受影响文件
- **新建** `backend/internal/service/svc/effective_role_test.go`
- **新建** `backend/internal/service/group/group_test.go`
- **新建** `backend/internal/handler/group_test.go`
- 运行 `go test ./...` 全量通过
- 运行 `next build` 前端编译通过
