# 003 - 任务列表

## Phase 1: DB Migrations (3 tasks)

- [ ] **T001** 创建 `000019_create_groups.up.sql` — groups 表 (id, name unique, description, created_by FK, timestamps)
- [ ] **T002** 创建 `000020_create_group_members.up.sql` — group_members 表 (id, group_id FK, user_id FK, created_at, unique(group_id,user_id))
- [ ] **T003** 创建 `000021_create_group_service_permissions.up.sql` — group_service_permissions 表 (id, group_id FK, service_id FK, role, created_at, unique(group_id,service_id))
- [ ] **T004** 创建对应 3 个 down migration 文件

## Phase 2: GORM Models (3 tasks)

- [ ] **T005** 新建 `model/group.go` — Group 结构体：ID, Name, Description, CreatedBy, CreatedAt, UpdatedAt, Creator *User
- [ ] **T006** 新建 `model/group_member.go` — GroupMember 结构体：ID, GroupID, UserID, CreatedAt, Group *Group, User *User
- [ ] **T007** 新建 `model/group_service_permission.go` — GroupServicePermission 结构体：ID, GroupID, ServiceID, Role, CreatedAt, Group *Group, Service *Service

## Phase 3: Repository Layer (6 tasks)

- [ ] **T008** 新建 `repository/group_repo.go` — GroupRepository 接口：Create, FindByID, FindByName, Update, Delete, List
- [ ] **T009** 新建 `repository/group_repo_impl.go` — GORM 实现 GroupRepository，List 含分页
- [ ] **T010** 新建 `repository/group_member_repo.go` — GroupMemberRepository 接口：Create, Delete, ListByGroup, ListByUser, Exists
- [ ] **T011** 新建 `repository/group_member_repo_impl.go` — GORM 实现，Create 时处理唯一索引冲突
- [ ] **T012** 新建 `repository/group_permission_repo.go` — GroupPermissionRepository 接口：Create, FindByID, Update, Delete, ListByGroup, FindByGroupAndService, FindRolesByUserAndService
- [ ] **T013** 新建 `repository/group_permission_repo_impl.go` — GORM 实现，FindRolesByUserAndService 通过 JOIN group_members 查询

## Phase 4: Group Service (5 tasks)

- [ ] **T014** 新建 `service/group/group.go` — GroupService 结构体 + 构造函数，注入 3 个 repo
- [ ] **T015** 实现组 CRUD：Create(name, desc, createdBy), GetByID, Update, List
- [ ] **T016** 实现组 Delete — 级联删除 group_members + group_service_permissions（事务）
- [ ] **T017** 实现成员管理：AddMembers(groupID, userIDs), RemoveMember(groupID, userID), ListMembers(groupID)
- [ ] **T018** 实现权限管理：AddPermission(groupID, serviceID, role), UpdatePermission(id, role), RemovePermission(id), ListPermissions(groupID)

## Phase 5: Group Service Tests (3 tasks)

- [ ] **T019** 新建 `service/group/group_test.go` — mock 3 个 repo
- [ ] **T020** 测试组 CRUD + 级联删除
- [ ] **T021** 测试成员管理 + 权限管理（重复添加返回错误等边界情况）

## Phase 6: GetEffectiveRole (5 tasks)

- [ ] **T022** 新建 `service/svc/effective_role.go` — EffectiveRoleService 结构体，注入 memberRepo + groupMemberRepo + groupPermRepo + userRepo
- [ ] **T023** 实现 GetEffectiveRole(userID, serviceID) → (role, sources): admin 直接返回 owner，合并个人+组权限取 max
- [ ] **T024** 实现 GetAllEffectivePermissions(userID) → []EffectivePermission：遍历所有 Service 生成权限全景
- [ ] **T025** 实现 Redis 缓存：GetEffectiveRole 先查 Redis，miss 时计算并写入 (TTL 5min)；InvalidateCache(userID) 删除 perm:{userID}:*
- [ ] **T026** 新建 `service/svc/effective_role_test.go` — 测试: admin绕过、个人权限、组权限、合并取max、无权限

## Phase 7: RBAC 中间件改造 (4 tasks)

- [ ] **T027** 修改 `service/svc/rbac.go` — RBACService 注入 EffectiveRoleService，CheckPermission 委托 GetEffectiveRole
- [ ] **T028** 修改 `middleware/rbac.go` — 确保 ServiceRBAC 中间件从 URL param / body / query 获取 serviceID
- [ ] **T029** 修改 `cmd/server/main.go` — 实例化 EffectiveRoleService + 新 RBACService，不再丢弃 rbacSvc
- [ ] **T030** 修改 `cmd/server/main.go` — 挂载 ServiceRBAC 到 service/build/deploy 路由（按 spec 权限级别）

## Phase 8: Group Handler (6 tasks)

- [ ] **T031** 新建 `handler/group.go` — GroupHandler 结构体 + 构造函数
- [ ] **T032** 实现组 CRUD 端点：List, Create, Get, Update, Delete
- [ ] **T033** 实现组成员端点：ListMembers, AddMembers, RemoveMember
- [ ] **T034** 实现组权限端点：ListPermissions, AddPermission, UpdatePermission, RemovePermission
- [ ] **T035** 实现 RegisterGroupRoutes — AdminOnly 中间件，注册所有路由
- [ ] **T036** 修改 `cmd/server/main.go` — 初始化 GroupHandler 并注册路由

## Phase 9: Permission Handler (3 tasks)

- [ ] **T037** 新建 `handler/permission.go` — PermissionHandler，注入 EffectiveRoleService
- [ ] **T038** 实现 GetUserPermissions (Admin only) + GetMyPermissions 端点
- [ ] **T039** 修改 `cmd/server/main.go` — 注册 GET /users/:id/permissions, GET /auth/my-permissions

## Phase 10: Admin Create User (3 tasks)

- [ ] **T040** 修改 `service/auth/auth.go` — 新增 AdminCreateUser(username, email, password, role) 方法
- [ ] **T041** 修改 `handler/user_admin.go` — 添加 CreateUser handler (POST /users)
- [ ] **T042** AdminCreateUser 测试 — 校验用户名/邮箱唯一、密码 BCrypt 哈希

## Phase 11: Handler Tests (3 tasks)

- [ ] **T043** 新建 `handler/group_test.go` — 测试组 CRUD + 成员 + 权限端点
- [ ] **T044** 新建 `handler/permission_test.go` — 测试权限总览端点
- [ ] **T045** 运行 `go test ./...` 全量通过

## Phase 12: Frontend Types + Hooks (3 tasks)

- [ ] **T046** 修改 `types/index.ts` — 新增 Group, GroupMember, GroupServicePermission, EffectivePermission 接口
- [ ] **T047** 新建 `hooks/use-groups.ts` — 组 CRUD + 成员 + 权限的 hooks
- [ ] **T048** 新建 `hooks/use-permissions.ts` — useUserPermissions, useMyPermissions

## Phase 13: Frontend 组管理 UI (4 tasks)

- [ ] **T049** 修改 `settings/layout.tsx` — settingsTabs 添加 `{ href: "/settings/groups", label: "组管理" }`
- [ ] **T050** 新建 `settings/groups/page.tsx` — 组列表页：名称、描述、成员数、Service 数、创建/编辑/删除
- [ ] **T051** 新建 `components/group/group-form-dialog.tsx` — 创建/编辑组对话框 (name + description)
- [ ] **T052** 新建 `components/group/group-detail-panel.tsx` — 组详情展开面板：成员管理区 + Service 权限区

## Phase 14: Frontend 用户管理增强 (3 tasks)

- [ ] **T053** 新建 `components/user/create-user-dialog.tsx` — Admin 创建用户对话框 (username, email, password, role)
- [ ] **T054** 新建 `components/user/permission-dialog.tsx` — 权限查看弹窗 (EffectivePermission 表格，标注来源)
- [ ] **T055** 修改 `settings/users/page.tsx` — 添加「创建用户」按钮 + 每行「权限」链接

## Phase 15: Frontend Build + Verification (2 tasks)

- [ ] **T056** 运行 `next build` 前端编译通过
- [ ] **T057** 运行 `go test ./...` 后端全量通过

---

**总计: 57 tasks, 15 phases**
