# Tasks: DeployHub — K8s 原生运维发布平台

**Input**: Design documents from `/specs/001-deployhub-platform/`

**Prerequisites**: plan.md, spec.md, data-model.md, contracts/api.md, research.md

**Tests**: TDD mandatory — write failing tests first, then implement.

**Layers**: 15 phases in strict order below. Each task targets 2–5 minutes. Code comments in Chinese.

---

## Phase 1: Project Scaffold

- [x] T001 Create root `docker-compose.yml` with PostgreSQL 15 + Redis 7 for local dev
- [x] T002 [P] Initialize Go module in `backend/go.mod` with Gin, GORM, client-go, gorilla/websocket, testify dependencies
- [x] T003 [P] Initialize Next.js 14 project in `frontend/` with Tailwind CSS, shadcn/ui, TanStack Query
- [x] T004 Create `backend/cmd/server/main.go` with Gin engine bootstrap, health check endpoint `/api/v1/health`
- [x] T005 Create `backend/internal/config/config.go` — load env vars (DATABASE_URL, REDIS_URL, JWT_SECRET, AES_KEY, SERVER_PORT)
- [x] T006 Create `backend/internal/pkg/response.go` — standardized JSON response helpers (Success, Error, Paginated)
- [x] T007 Create `backend/internal/pkg/errors.go` — error codes and app error type
- [x] T008 Create `backend/Makefile` with targets: dev, test, lint, migrate-up, migrate-down, migrate-create
- [x] T009 [P] Create `frontend/src/lib/api-client.ts` — axios instance with JWT interceptor, base URL config
- [x] T010 [P] Create `frontend/src/lib/ws-client.ts` — WebSocket client with auto-reconnect and heartbeat
- [x] T011 Create `frontend/src/app/layout.tsx` — root layout with TanStack QueryClientProvider
- [x] T012 Create `frontend/src/components/layout/sidebar.tsx` — sidebar navigation with 7 menu items
- [x] T013 Create `frontend/src/app/(dashboard)/layout.tsx` — dashboard layout with sidebar

**Checkpoint**: Backend starts with `/api/v1/health`, frontend shows sidebar layout

---

## Phase 2: Data Model & Migrations

For each entity: write test → create migration SQL (up+down) → create GORM model.

- [x] T014 Create `backend/internal/model/user.go` — User struct with GORM tags per data-model.md
- [x] T015 Create `backend/migrations/000001_create_users.up.sql` and `backend/migrations/000001_create_users.down.sql`
- [x] T016 [P] Create `backend/internal/model/cluster.go` — Cluster struct
- [x] T017 [P] Create `backend/migrations/000002_create_clusters.up.sql` and `backend/migrations/000002_create_clusters.down.sql`
- [x] T018 [P] Create `backend/internal/model/gitrepo.go` — GitRepo struct
- [x] T019 [P] Create `backend/migrations/000003_create_git_repos.up.sql` and `backend/migrations/000003_create_git_repos.down.sql`
- [x] T020 [P] Create `backend/internal/model/registry.go` — Registry struct
- [x] T021 [P] Create `backend/migrations/000004_create_registries.up.sql` and `backend/migrations/000004_create_registries.down.sql`
- [x] T022 Create `backend/internal/model/service.go` — Service struct with FK relationships
- [x] T023 Create `backend/migrations/000005_create_services.up.sql` and `backend/migrations/000005_create_services.down.sql`
- [x] T024 [P] Create `backend/internal/model/service_member.go` — ServiceMember struct
- [x] T025 [P] Create `backend/migrations/000006_create_service_members.up.sql` and `backend/migrations/000006_create_service_members.down.sql`
- [x] T026 [P] Create `backend/internal/model/build.go` — Build struct with status enum
- [x] T027 [P] Create `backend/migrations/000007_create_builds.up.sql` and `backend/migrations/000007_create_builds.down.sql`
- [x] T028 [P] Create `backend/internal/model/deployment.go` — Deployment struct with rollback fields
- [x] T029 [P] Create `backend/migrations/000008_create_deployments.up.sql` and `backend/migrations/000008_create_deployments.down.sql`
- [x] T030 [P] Create `backend/internal/model/approval.go` — Approval struct
- [x] T031 [P] Create `backend/migrations/000009_create_approvals.up.sql` and `backend/migrations/000009_create_approvals.down.sql`
- [x] T032 [P] Create `backend/internal/model/config_template.go` — ConfigTemplate struct
- [x] T033 [P] Create `backend/migrations/000010_create_config_templates.up.sql` and `backend/migrations/000010_create_config_templates.down.sql`
- [x] T034 [P] Create `backend/internal/model/config_env_value.go` — ConfigEnvValue struct
- [x] T035 [P] Create `backend/migrations/000011_create_config_env_values.up.sql` and `backend/migrations/000011_create_config_env_values.down.sql`
- [x] T036 [P] Create `backend/internal/model/config_version.go` — ConfigVersion struct
- [x] T037 [P] Create `backend/migrations/000012_create_config_versions.up.sql` and `backend/migrations/000012_create_config_versions.down.sql`
- [x] T038 [P] Create `backend/internal/model/config_deployment.go` — ConfigDeployment struct
- [x] T039 [P] Create `backend/migrations/000013_create_config_deployments.up.sql` and `backend/migrations/000013_create_config_deployments.down.sql`
- [x] T040 [P] Create `backend/internal/model/notification.go` — Notification struct
- [x] T041 [P] Create `backend/migrations/000014_create_notifications.up.sql` and `backend/migrations/000014_create_notifications.down.sql`
- [x] T042 [P] Create `backend/internal/model/audit_log.go` — AuditLog struct
- [x] T043 [P] Create `backend/migrations/000015_create_audit_logs.up.sql` and `backend/migrations/000015_create_audit_logs.down.sql`
- [x] T044 Create `backend/internal/model/db.go` — GORM database connection init with PostgreSQL DSN
- [x] T045 Run all migrations up via `backend/Makefile` and verify tables under `backend/migrations/` are applied in PostgreSQL

**Checkpoint**: All 15 tables created in PostgreSQL, GORM models compile

---

## Phase 3: Auth Module

- [x] T046 Write test for crypto service in `backend/internal/service/crypto/crypto_test.go` — test AES-256-GCM encrypt/decrypt roundtrip
- [x] T047 Implement `backend/internal/service/crypto/crypto.go` — AES-256-GCM encrypt/decrypt using Go crypto/aes+cipher
- [x] T048 Write test for password hashing in `backend/internal/service/auth/password_test.go` — BCrypt hash and verify
- [x] T049 Implement `backend/internal/service/auth/password.go` — BCrypt hash and compare
- [x] T050 Define UserRepository interface in `backend/internal/repository/user_repo.go`
- [x] T051 Implement UserRepository with GORM in `backend/internal/repository/user_repo_impl.go`
- [x] T052 Write test for JWT token generation/validation in `backend/internal/service/auth/jwt_test.go`
- [x] T053 Implement `backend/internal/service/auth/jwt.go` — generate and validate JWT tokens
- [x] T054 Write test for auth service (register, login) in `backend/internal/service/auth/auth_test.go`
- [x] T055 Implement `backend/internal/service/auth/auth.go` — register (BCrypt), login (verify+JWT), OAuth callback
- [x] T056 Write httptest for auth handlers in `backend/internal/handler/auth_test.go` — POST /register, /login, GET /me
- [x] T057 Implement `backend/internal/handler/auth.go` — register, login, OAuth callback, get me, update me
- [x] T058 Create `backend/internal/middleware/jwt.go` — JWT auth middleware extracting user from token
- [x] T059 Write test for JWT middleware in `backend/internal/middleware/jwt_test.go`
- [x] T060 Create `backend/internal/middleware/admin.go` — admin-only middleware checking user.role
- [x] T061 Register auth routes in `backend/cmd/server/main.go` — `/api/v1/auth/*` group

**Checkpoint**: Can register, login, get JWT, access `/me` with token

---

## Phase 4: Cluster Management

- [x] T062 Define ClusterRepository interface in `backend/internal/repository/cluster_repo.go`
- [x] T063 Implement ClusterRepository with GORM in `backend/internal/repository/cluster_repo_impl.go`
- [x] T064 Write test for cluster service in `backend/internal/service/cluster/cluster_test.go` — CRUD with encrypted kubeconfig
- [x] T065 Implement `backend/internal/service/cluster/cluster.go` — Create (encrypt kubeconfig), List, Get, Update, Delete
- [x] T066 Implement `backend/internal/service/cluster/clientset.go` — build client-go clientset from decrypted kubeconfig, clientset cache/pool
- [x] T067 Write test for cluster clientset in `backend/internal/service/cluster/clientset_test.go` — verify clientset creation from kubeconfig
- [x] T068 Implement `backend/internal/service/cluster/test_connection.go` — test K8s API reachability, return server version
- [x] T069 Write httptest for cluster handlers in `backend/internal/handler/cluster_test.go`
- [x] T070 Implement `backend/internal/handler/cluster.go` — CRUD + test connection endpoints
- [x] T071 Register cluster routes in `backend/cmd/server/main.go` — `/api/v1/clusters/*`

**Checkpoint**: Can add cluster with kubeconfig (encrypted), test connection

---

## Phase 5: Git Repo & Registry Management

- [x] T072 [P] Define GitRepoRepository interface in `backend/internal/repository/gitrepo_repo.go`
- [x] T073 [P] Define RegistryRepository interface in `backend/internal/repository/registry_repo.go`
- [x] T074 [P] Implement GitRepoRepository in `backend/internal/repository/gitrepo_repo_impl.go`
- [x] T075 [P] Implement RegistryRepository in `backend/internal/repository/registry_repo_impl.go`
- [x] T076 Write test for gitrepo service in `backend/internal/service/gitrepo/gitrepo_test.go` — CRUD + encrypt credential
- [x] T077 Implement `backend/internal/service/gitrepo/gitrepo.go` — CRUD + test connection + list branches
- [x] T078 [P] Write test for registry service in `backend/internal/service/registry/registry_test.go`
- [x] T079 [P] Implement `backend/internal/service/registry/registry.go` — CRUD with encrypted auth_config
- [x] T080 Write httptest for gitrepo handlers in `backend/internal/handler/gitrepo_test.go`
- [x] T081 Implement `backend/internal/handler/gitrepo.go` — CRUD + test + branches endpoints
- [x] T082 [P] Write httptest for registry handlers in `backend/internal/handler/registry_test.go`
- [x] T083 [P] Implement `backend/internal/handler/registry.go` — CRUD endpoints
- [x] T084 Register gitrepo and registry routes in `backend/cmd/server/main.go`

**Checkpoint**: Can manage Git repos and Registries via API

---

## Phase 6: Service Management

- [x] T085 Define ServiceRepository interface in `backend/internal/repository/service_repo.go`
- [x] T086 Implement ServiceRepository in `backend/internal/repository/service_repo_impl.go`
- [x] T087 Define ServiceMemberRepository interface in `backend/internal/repository/service_member_repo.go`
- [x] T088 Implement ServiceMemberRepository in `backend/internal/repository/service_member_repo_impl.go`
- [x] T089 Write test for service management in `backend/internal/service/svc/svc_test.go` — Create, List, Get, Update, Delete with owner auto-assignment
- [x] T090 Implement `backend/internal/service/svc/svc.go` — CRUD + auto-create owner ServiceMember on create
- [x] T091 Write test for RBAC check in `backend/internal/service/svc/rbac_test.go` — owner/developer/viewer permission boundaries
- [x] T092 Implement `backend/internal/service/svc/rbac.go` — CheckPermission(userID, serviceID, requiredRole) method
- [x] T093 Create `backend/internal/middleware/rbac.go` — service-level RBAC middleware using svc.CheckPermission
- [x] T094 Write test for member management in `backend/internal/service/svc/member_test.go` — add, update role, remove
- [x] T095 Implement `backend/internal/service/svc/member.go` — AddMember, UpdateRole, RemoveMember
- [x] T096 Write httptest for service handlers in `backend/internal/handler/service_test.go`
- [x] T097 Implement `backend/internal/handler/service.go` — CRUD + member management endpoints
- [x] T098 Register service routes in `backend/cmd/server/main.go` — `/api/v1/services/*`

**Checkpoint**: Can create services with RBAC, manage members

---

## Phase 7: Service Import

- [x] T099 Write test for K8s YAML parser in `backend/internal/service/svc/import_test.go` — parse Deployment+Service YAML to Service fields
- [x] T100 Implement `backend/internal/service/svc/import_yaml.go` — parse K8s Deployment YAML extracting name, image, replicas, ports, resources
- [x] T101 Write test for simplified format parser in `backend/internal/service/svc/import_simple_test.go`
- [x] T102 Implement `backend/internal/service/svc/import_simple.go` — parse simplified YAML/JSON format
- [x] T103 Implement `backend/internal/service/svc/import.go` — Preview (parse + validate + detect duplicates) and Confirm (batch create)
- [x] T104 Write httptest for import handlers in `backend/internal/handler/service_test.go` (import section)
- [x] T105 Add import endpoints to `backend/internal/handler/service.go` — POST `/services/import/preview`, POST `/services/import`

**Checkpoint**: Can upload YAML, preview parsed results, confirm batch import

---

## Phase 8: Build Engine

- [x] T106 Define BuildRepository interface in `backend/internal/repository/build_repo.go`
- [x] T107 Implement BuildRepository in `backend/internal/repository/build_repo_impl.go`
- [x] T108 Write test for build service in `backend/internal/service/build/build_test.go` — create build, status transitions
- [x] T109 Implement `backend/internal/service/build/build.go` — CreateBuild (permission check, record creation)
- [x] T110 Implement `backend/internal/service/build/kaniko.go` — generate Kaniko Job spec (git clone init container + kaniko main container + registry push)
- [x] T111 Write test for Kaniko Job spec generation in `backend/internal/service/build/kaniko_test.go`
- [x] T112 Implement `backend/internal/service/build/watcher.go` — Watch K8s Job status via client-go, update Build record on completion/failure
- [x] T113 Implement `backend/internal/service/build/cleanup.go` — delete completed/failed Kaniko Job and Pod resources
- [x] T114 Create `backend/internal/ws/hub.go` — WebSocket hub managing connections per build/deployment
- [x] T115 Create `backend/internal/ws/client.go` — WebSocket client connection with read/write pumps, heartbeat
- [x] T116 Implement `backend/internal/handler/ws.go` — WebSocket upgrade handler for `/ws/builds/:id/log`
- [x] T117 Implement `backend/internal/service/build/log_streamer.go` — stream Pod logs via client-go to WebSocket hub
- [x] T118 Write httptest for build handlers in `backend/internal/handler/build_test.go`
- [x] T119 Implement `backend/internal/handler/build.go` — trigger build, list, detail, cancel, get log
- [x] T120 Register build routes and WS route in `backend/cmd/server/main.go`

**Checkpoint**: Can trigger build, watch Kaniko Job, stream logs via WebSocket

---

## Phase 9: Deploy Engine

- [x] T121 Define DeploymentRepository interface in `backend/internal/repository/deployment_repo.go`
- [x] T122 Implement DeploymentRepository in `backend/internal/repository/deployment_repo_impl.go`
- [x] T123 Write test for deploy service in `backend/internal/service/deploy/deploy_test.go` — create deployment, concurrent deployment lock
- [x] T124 Implement `backend/internal/service/deploy/deploy.go` — CreateDeployment (concurrent check, record prev image tag)
- [x] T125 Implement `backend/internal/service/deploy/executor.go` — patch K8s Deployment via client-go (update image, replicas)
- [x] T126 Write test for deployment executor in `backend/internal/service/deploy/executor_test.go`
- [x] T127 Implement `backend/internal/service/deploy/watcher.go` — Watch Deployment rollout status (readyReplicas), push progress via WS hub
- [x] T128 Add WebSocket handler for `/ws/deployments/:id/progress` in `backend/internal/handler/ws.go`
- [x] T129 Implement `backend/internal/service/deploy/rollback.go` — create rollback deployment (is_rollback=true, rollback_from_id, use target image_tag)
- [x] T130 Write httptest for deploy handlers in `backend/internal/handler/deploy_test.go`
- [x] T131 Implement `backend/internal/handler/deploy.go` — create deployment, list, detail, rollback
- [x] T132 Register deployment routes in `backend/cmd/server/main.go`

**Checkpoint**: Can deploy via Rolling Update, watch progress via WebSocket, rollback

---

## Phase 10: Approval Engine

- [x] T133 Define ApprovalRepository interface in `backend/internal/repository/approval_repo.go`
- [x] T134 Implement ApprovalRepository in `backend/internal/repository/approval_repo_impl.go`
- [x] T135 Write test for approval service in `backend/internal/service/approval/approval_test.go` — rules (owner skip, developer needs approval, admin bypass)
- [x] T136 Implement `backend/internal/service/approval/approval.go` — CreateApproval, Approve, Reject with rule engine
- [x] T137 Implement `backend/internal/service/approval/rules.go` — approval rule evaluation (owner auto-approve, developer requires owner, admin emergency)
- [x] T138 Wire approval into deploy flow in `backend/internal/service/deploy/deploy.go` and `backend/internal/service/approval/approval.go` — after CreateDeployment, check if approval needed; on Approve trigger executor
- [x] T139 Write httptest for approval handlers in `backend/internal/handler/approval_test.go`
- [x] T140 Implement `backend/internal/handler/approval.go` — list, detail, approve, reject
- [x] T141 Register approval routes in `backend/cmd/server/main.go`

**Checkpoint**: Full deploy+approval flow works end-to-end

---

## Phase 11: Config Center

- [x] T142 [P] Define ConfigTemplateRepository in `backend/internal/repository/config_template_repo.go`
- [x] T143 [P] Define ConfigEnvValueRepository in `backend/internal/repository/config_env_value_repo.go`
- [x] T144 [P] Define ConfigVersionRepository in `backend/internal/repository/config_version_repo.go`
- [x] T145 [P] Define ConfigDeploymentRepository in `backend/internal/repository/config_deployment_repo.go`
- [x] T146 [P] Implement `backend/internal/repository/config_template_repo_impl.go`, `backend/internal/repository/config_env_value_repo_impl.go`, `backend/internal/repository/config_version_repo_impl.go`, `backend/internal/repository/config_deployment_repo_impl.go`
- [x] T147 Write test for config template rendering in `backend/internal/service/config/render_test.go` — Go template with variables
- [x] T148 Implement `backend/internal/service/config/render.go` — render Go template with env values, validation, sandboxing
- [x] T149 Write test for config service in `backend/internal/service/config/config_test.go` — template CRUD, env values, versioning
- [x] T150 Implement `backend/internal/service/config/config.go` — template CRUD, env value management, render preview, create version
- [x] T151 Implement `backend/internal/service/config/diff.go` — text diff between two ConfigVersion rendered contents
- [x] T152 Implement `backend/internal/service/config/sync.go` — deploy rendered config to K8s ConfigMap/Secret via client-go
- [x] T153 Write httptest for config handlers in `backend/internal/handler/config_test.go`
- [x] T154 Implement `backend/internal/handler/config.go` — all config endpoints (template CRUD, env-values, render, versions, diff, deploy)
- [x] T155 Register config routes in `backend/cmd/server/main.go`

**Checkpoint**: Can create templates, fill env vars, render, diff versions, deploy to K8s

---

## Phase 12: Notification Engine

- [x] T156 Define NotificationRepository in `backend/internal/repository/notification_repo.go`
- [x] T157 Implement NotificationRepository in `backend/internal/repository/notification_repo_impl.go`
- [x] T158 Write test for notification service in `backend/internal/service/notification/notification_test.go`
- [x] T159 Implement `backend/internal/service/notification/notification.go` — create notification, list, mark read, unread count
- [x] T160 Implement `backend/internal/service/notification/webhook.go` — send to Feishu/DingTalk/Slack webhook (HTTP POST with channel-specific payload format)
- [x] T161 Write test for webhook sender in `backend/internal/service/notification/webhook_test.go` — mock HTTP for each channel type
- [x] T162 Wire notifications in `backend/internal/service/build/build.go`, `backend/internal/service/deploy/deploy.go`, `backend/internal/service/approval/approval.go` — on build complete/fail, deploy status change, approval create/approve/reject
- [x] T163 Write httptest for notification handlers in `backend/internal/handler/notification_test.go`
- [x] T164 Implement `backend/internal/handler/notification.go` — list, mark read, read all, unread count
- [x] T165 Add NotificationChannel model + migration + CRUD for admin in `backend/internal/model/notification_channel.go`, `backend/migrations/000016_create_notification_channels.up.sql`, `backend/migrations/000016_create_notification_channels.down.sql`, `backend/internal/handler/notification_channel.go`
- [x] T166 Register notification and channel routes in `backend/cmd/server/main.go`

**Checkpoint**: Notifications created on events, webhook delivery works

---

## Phase 13: Audit Logging

- [x] T167 Define AuditLogRepository in `backend/internal/repository/audit_log_repo.go`
- [x] T168 Implement AuditLogRepository in `backend/internal/repository/audit_log_repo_impl.go`
- [x] T169 Write test for audit service in `backend/internal/service/audit/audit_test.go`
- [x] T170 Implement `backend/internal/service/audit/audit.go` — CreateLog (sanitize detail), List with filters
- [x] T171 Create `backend/internal/middleware/audit.go` — middleware that logs write operations (POST/PUT/DELETE) automatically
- [x] T172 Write test for audit middleware in `backend/internal/middleware/audit_test.go`
- [x] T173 Write httptest for audit handlers in `backend/internal/handler/audit_test.go`
- [x] T174 Implement `backend/internal/handler/audit.go` — GET `/audit-logs` with filters (user_id, action, resource_type, date range)
- [x] T175 Register audit middleware on all write routes and audit-logs route in `backend/cmd/server/main.go`
- [x] T176 Implement admin user management in `backend/internal/handler/user_admin.go` — GET `/users`, PUT `/users/:id/role`, PUT `/users/:id/status`

**Checkpoint**: All write operations produce audit logs, admin can query and manage users

---

## Phase 14: Frontend Pages

### Auth Pages

- [x] T177 Create `frontend/src/app/(auth)/login/page.tsx` — login form with username/password
- [x] T178 [P] Create `frontend/src/app/(auth)/register/page.tsx` — registration form
- [x] T179 Create `frontend/src/hooks/use-auth.ts` — TanStack Query hooks for login, register, getMe

### Service Management (首页)

- [x] T180 Create `frontend/src/types/index.ts` — TypeScript types for all entities (User, Service, Build, Deployment, etc.)
- [x] T181 Create `frontend/src/hooks/use-services.ts` — TanStack Query hooks for service CRUD
- [x] T182 Create `frontend/src/app/(dashboard)/services/page.tsx` — service list with filters (cluster, search), action buttons
- [x] T183 Create `frontend/src/components/service/service-create-dialog.tsx` — create service form dialog
- [x] T184 Create `frontend/src/app/(dashboard)/services/[id]/page.tsx` — service detail page

### Build Center

- [x] T185 Create `frontend/src/hooks/use-builds.ts` — TanStack Query hooks for builds + WebSocket log stream
- [x] T186 Create `frontend/src/app/(dashboard)/builds/page.tsx` — build list with status badges and progress bars
- [x] T187 Create `frontend/src/components/build/build-log-viewer.tsx` — real-time log viewer component using `frontend/src/lib/ws-client.ts`
- [x] T188 Create `frontend/src/components/build/trigger-build-dialog.tsx` — select service + branch, trigger build

### Deploy Management

- [x] T189 Create `frontend/src/hooks/use-deployments.ts` — TanStack Query hooks for deployments + WebSocket progress
- [x] T190 Create `frontend/src/app/(dashboard)/deployments/page.tsx` — deployment list with status timeline
- [x] T191 Create `frontend/src/app/(dashboard)/deployments/[id]/page.tsx` — deployment detail with Pod status cards, progress bar, timeline
- [x] T192 Create `frontend/src/components/deploy/deploy-dialog.tsx` — select service + build + cluster, deploy
- [x] T193 Create `frontend/src/components/deploy/rollback-dialog.tsx` — select target version, confirm rollback

### Config Center

- [x] T194 Create `frontend/src/hooks/use-configs.ts` — TanStack Query hooks for config templates
- [x] T195 Create `frontend/src/app/(dashboard)/configs/page.tsx` — config template list by service
- [x] T196 Create `frontend/src/components/config/template-editor.tsx` — textarea with syntax highlighting for Go template
- [x] T197 Create `frontend/src/components/config/env-values-editor.tsx` — per-cluster variable key-value editor
- [x] T198 Create `frontend/src/components/config/version-diff.tsx` — side-by-side diff viewer component
- [x] T199 Create `frontend/src/components/config/deploy-config-dialog.tsx` — select cluster + namespace, deploy config

### Approval Center

- [x] T200 Create `frontend/src/hooks/use-approvals.ts` — TanStack Query hooks for approvals
- [x] T201 Create `frontend/src/app/(dashboard)/approvals/page.tsx` — pending approvals list with approve/reject actions

### Notification Center

- [x] T202 Create `frontend/src/hooks/use-notifications.ts` — TanStack Query hooks for notifications
- [x] T203 Create `frontend/src/app/(dashboard)/notifications/page.tsx` — notification list with unread badge, mark read

### System Settings

- [x] T204 Create `frontend/src/app/(dashboard)/settings/page.tsx` — settings layout with tabs
- [x] T205 Create `frontend/src/app/(dashboard)/settings/clusters/page.tsx` — cluster management (add, edit, test, delete)
- [x] T206 [P] Create `frontend/src/app/(dashboard)/settings/git-repos/page.tsx` — git repo management
- [x] T207 [P] Create `frontend/src/app/(dashboard)/settings/registries/page.tsx` — registry management
- [x] T208 Create `frontend/src/app/(dashboard)/settings/users/page.tsx` — user management (role, status)
- [x] T209 Create `frontend/src/app/(dashboard)/settings/notifications/page.tsx` — notification channel management

**Checkpoint**: All 7 frontend page modules functional with API integration

---

## Phase 15: Integration Testing & Polish

- [ ] T210 Write end-to-end test in `backend/test/e2e/setup_flow_test.go` — register → login → create cluster → create git repo → create registry → create service
- [ ] T211 Write end-to-end test in `backend/test/e2e/cicd_flow_test.go` — create service → trigger build → verify build status → deploy → approve → verify deployment
- [ ] T212 Write end-to-end test in `backend/test/e2e/rollback_flow_test.go` — deploy v1 → deploy v2 → rollback to v1 → verify rollback record
- [ ] T213 Write end-to-end test in `backend/test/e2e/config_flow_test.go` — create config template → fill env vars → render → deploy to cluster
- [x] T214 Add request logging middleware in `backend/internal/middleware/logger.go` — structured JSON logs
- [x] T215 Add CORS middleware in `backend/internal/middleware/cors.go` — allow frontend origin
- [x] T216 Create `backend/Dockerfile` — multi-stage build (builder + runtime)
- [x] T217 Create `frontend/Dockerfile` — multi-stage build (deps + build + runtime)
- [x] T218 Run full lint pass on backend (`go vet`, golangci-lint) and fix issues in `backend/`
- [x] T219 Run full lint pass on frontend (eslint, `tsc --noEmit`) and fix issues in `frontend/`
- [ ] T220 Verify all acceptance scenarios from `specs/001-deployhub-platform/spec.md` pass manually

**Checkpoint**: Full platform operational, all tests passing

---

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1 (Scaffold): No dependencies
- Phase 2 (Data Model): Depends on Phase 1
- Phase 3 (Auth): Depends on Phase 2
- Phase 4 (Cluster): Depends on Phase 3 (needs auth middleware)
- Phase 5 (Git+Registry): Depends on Phase 3 (needs auth middleware)
- Phase 6 (Service): Depends on Phase 4+5 (needs cluster, git, registry)
- Phase 7 (Import): Depends on Phase 6
- Phase 8 (Build): Depends on Phase 6 + Phase 4 (needs service + cluster clientset)
- Phase 9 (Deploy): Depends on Phase 8 (needs build artifacts) + Phase 4 (clientset)
- Phase 10 (Approval): Depends on Phase 9
- Phase 11 (Config): Depends on Phase 6 + Phase 4 (needs service + cluster)
- Phase 12 (Notification): Depends on Phase 10 (wire into approval/build/deploy)
- Phase 13 (Audit): Depends on Phase 3 (needs auth middleware for user context)
- Phase 14 (Frontend): Depends on Phase 3-13 (needs backend APIs)
- Phase 15 (Integration): Depends on Phase 1-14

### Parallel Opportunities

- Phase 4 and Phase 5 can run in parallel (both only need auth)
- Phase 11 (Config) can run in parallel with Phase 8-10 (Build/Deploy/Approval)
- Phase 13 (Audit) can start as soon as Phase 3 is done
- Within Phase 2, all entity models marked `[P]` can be created in parallel
- Within Phase 14, different page modules can be developed in parallel

---

## Implementation Strategy

### MVP (Phase 1-10)

1. Scaffold + Data Model + Auth → Foundation
2. Cluster + Git + Registry + Service → Infrastructure management
3. Build + Deploy + Approval → Core CI/CD loop
4. Validate full flow: create service → build → deploy → approve

### Incremental Additions

5. Config Center (Phase 11) → Configuration management
6. Notification + Audit (Phase 12-13) → Observability
7. Frontend (Phase 14) → User interface
8. Integration Testing (Phase 15) → Quality assurance

---

## Notes

- `[P]` tasks = different files, no dependencies
- TDD: Write failing test BEFORE implementation for every service and handler
- Code comments in Chinese
- Each task should be completable in 2–5 minutes
- Commit after each task or logical group
- Constitution: All sensitive data encrypted (AES-256-GCM), RBAC enforced, audit logged
