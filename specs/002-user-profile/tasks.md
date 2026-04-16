# Tasks: 002-user-profile 用户资料管理与登录登出增强

## Phase 1: 数据库迁移（Layer 1）

- [ ] **T001** 创建 `backend/migrations/000018_add_user_profile_fields.up.sql`：ALTER TABLE users ADD COLUMN nickname VARCHAR(100) DEFAULT '', ADD COLUMN phone_encrypted TEXT DEFAULT ''
- [ ] **T002** 创建 `backend/migrations/000018_add_user_profile_fields.down.sql`：ALTER TABLE users DROP COLUMN nickname, DROP COLUMN phone_encrypted

## Phase 2: User 模型更新（Layer 2）

- [ ] **T003** 修改 `backend/internal/model/user.go`：添加 Nickname 和 PhoneEncrypted 字段（含 GORM tags 和 json tags）

## Phase 3: 配置扩展（Layer 2）

- [ ] **T004** 修改 `backend/internal/config/config.go`：Config 结构体新增 S3Endpoint/S3AccessKey/S3SecretKey/S3Bucket/S3Region 字段，Load() 从环境变量读取
- [ ] **T005** 修改 `backend/.env.example`：添加 S3_ENDPOINT/S3_ACCESS_KEY/S3_SECRET_KEY/S3_BUCKET/S3_REGION 示例

## Phase 4: Storage 服务（Layer 3）

- [ ] **T006** 创建 `backend/internal/service/storage/storage.go`：定义 StorageService 接口（Upload(file, contentType) → url, error）
- [ ] **T007** 在 storage.go 中实现 S3StorageService：基于 aws-sdk-go-v2 S3 兼容客户端
- [ ] **T008** 在 S3StorageService.Upload 中实现：文件类型白名单校验（jpg/png/webp）、大小限制 2MB、UUID 重命名
- [ ] **T009** 实现 NewS3StorageService(endpoint, accessKey, secretKey, bucket, region) 构造函数
- [ ] **T010** 创建 `backend/internal/service/storage/storage_test.go`：测试文件类型校验拒绝非法类型、大小超限拒绝

## Phase 5: 手机号脱敏工具（Layer 4）

- [ ] **T011** 创建 `backend/internal/service/auth/mask_test.go`：测试 MaskPhone 函数——正常11位("13812341234"→"138****1234")、短号码原样返回、空字符串返回空
- [ ] **T012** 创建 `backend/internal/service/auth/mask.go`：实现 MaskPhone(phone string) string，保留前3后4，中间替换为 ****

## Phase 6: Repository 层扩展（Layer 4）

- [ ] **T013** 修改 `backend/internal/repository/user_repo.go`：接口新增 UpdateProfile(id uint, nickname, phoneEncrypted string) error
- [ ] **T014** 修改 `backend/internal/repository/user_repo.go`：接口新增 UpdateAvatar(id uint, avatarURL string) error
- [ ] **T015** 修改 `backend/internal/repository/user_repo.go`：接口新增 UpdatePassword(id uint, passwordHash string) error
- [ ] **T016** 修改 `backend/internal/repository/user_repo_impl.go`：实现 UpdateProfile — db.Model(&User{}).Where("id = ?", id).Updates(map)
- [ ] **T017** 修改 `backend/internal/repository/user_repo_impl.go`：实现 UpdateAvatar — db.Model(&User{}).Where("id = ?", id).Update("avatar", url)
- [ ] **T018** 修改 `backend/internal/repository/user_repo_impl.go`：实现 UpdatePassword — db.Model(&User{}).Where("id = ?", id).Update("password_hash", hash)

## Phase 7: Auth Service 扩展（Layer 4）

- [ ] **T019** 修改 `backend/internal/service/auth/auth_test.go`：新增 TestUpdateProfile 测试——nickname 和 phone 加密存储
- [ ] **T020** 修改 `backend/internal/service/auth/auth.go`：AuthService 结构体新增 cryptoSvc 字段，修改构造函数签名
- [ ] **T021** 修改 `backend/internal/service/auth/auth.go`：实现 UpdateProfile(userID, nickname, phone) — phone 通过 cryptoSvc.Encrypt 加密后调 repo.UpdateProfile
- [ ] **T022** 修改 `backend/internal/service/auth/auth_test.go`：新增 TestChangePassword 测试——旧密码正确时成功、旧密码错误时失败
- [ ] **T023** 修改 `backend/internal/service/auth/auth.go`：实现 ChangePassword(userID, oldPassword, newPassword) — 验证旧密码，hash 新密码，调 repo.UpdatePassword
- [ ] **T024** 修改 `backend/internal/service/auth/auth.go`：实现 UpdateAvatar(userID, avatarURL) — 调 repo.UpdateAvatar
- [ ] **T025** 修改 `backend/internal/service/auth/auth.go`：实现 GetFullProfile(userID) — 获取用户，phone 解密+脱敏返回

## Phase 8: Auth Handler 扩展（Layer 5）

- [ ] **T026** 修改 `backend/internal/handler/auth.go`：AuthHandler 结构体新增 storageSvc 字段，修改构造函数签名
- [ ] **T027** 修改 `backend/internal/handler/auth.go`：增强 GetMe — 调用 GetFullProfile 返回完整字段（avatar/nickname/phone 脱敏）
- [ ] **T028** 修改 `backend/internal/handler/auth.go`：新增 UpdateProfile handler — 绑定 nickname/phone，调 authSvc.UpdateProfile
- [ ] **T029** 修改 `backend/internal/handler/auth.go`：新增 ChangePassword handler — 绑定 old_password/new_password，调 authSvc.ChangePassword
- [ ] **T030** 修改 `backend/internal/handler/auth.go`：新增 UploadAvatar handler — 解析 multipart file，调 storageSvc.Upload，再调 authSvc.UpdateAvatar
- [ ] **T031** 修改 `backend/internal/handler/auth.go`：新增 Logout handler — 返回 200 message

## Phase 9: 路由注册与 main.go 接线（Layer 5）

- [ ] **T032** 修改 `backend/cmd/server/main.go`：初始化 StorageService（从 config 读取 S3 配置）
- [ ] **T033** 修改 `backend/cmd/server/main.go`：AuthService 构造函数传入 cryptoSvc
- [ ] **T034** 修改 `backend/cmd/server/main.go`：AuthHandler 构造函数传入 storageSvc
- [ ] **T035** 修改 `backend/cmd/server/main.go`：在 auth 路由组注册新路由 PUT /profile、PUT /password、POST /avatar、POST /logout

## Phase 10: 前端 Hooks（Layer 6）

- [ ] **T036** 创建 `frontend/src/hooks/use-profile.ts`：实现 useUpdateProfile hook — PUT /auth/profile
- [ ] **T037** 在 use-profile.ts 中实现 useUploadAvatar hook — POST /auth/avatar（FormData）
- [ ] **T038** 在 use-profile.ts 中实现 useChangePassword hook — PUT /auth/password
- [ ] **T039** 修改 `frontend/src/hooks/use-auth.ts`：useLogout 增加调用 POST /auth/logout 再清 token

## Phase 11: 前端 UI — 侧边栏用户菜单（Layer 7）

- [ ] **T040** 创建 `frontend/src/components/layout/user-menu.tsx`：UserMenu 组件——使用 useMe 获取用户信息，显示头像（默认占位）+ 用户名
- [ ] **T041** UserMenu 组件添加下拉菜单：点击展开，显示"个人资料"链接和"退出登录"按钮
- [ ] **T042** 修改 `frontend/src/components/layout/sidebar.tsx`：底部导入并渲染 UserMenu 组件

## Phase 12: 前端 UI — 个人资料页面（Layer 7）

- [ ] **T043** 创建 `frontend/src/app/(dashboard)/profile/page.tsx`：页面框架——标题"个人资料"，三个区块（头像、基本信息、修改密码）
- [ ] **T044** 创建 `frontend/src/components/user/avatar-upload.tsx`：AvatarUpload 组件——显示当前头像，点击/拖拽触发文件选择，调用 useUploadAvatar 上传
- [ ] **T045** 在 profile/page.tsx 中集成 AvatarUpload 组件
- [ ] **T046** 在 profile/page.tsx 中实现基本信息编辑区块：昵称输入框、手机号输入框、保存按钮，调用 useUpdateProfile
- [ ] **T047** 在 profile/page.tsx 中实现修改密码区块：旧密码、新密码、确认新密码输入框，调用 useChangePassword
- [ ] **T048** 在 profile/page.tsx 中添加成功提示：保存资料/修改密码/上传头像成功后弹出绿色 Toast

## Phase 13: 验证与收尾

- [ ] **T049** 修改 `backend/.env`：添加 S3 配置项（可留空，storage 服务在未配置时 Upload 返回错误）
- [ ] **T050** 运行 go vet ./... 确认后端无编译错误
- [ ] **T051** 运行 npx tsc --noEmit 确认前端无类型错误
