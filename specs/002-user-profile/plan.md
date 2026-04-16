# Plan: 用户资料管理与登录登出增强

## Summary

在现有 auth 模块上扩展用户资料管理能力。分 6 层实现：数据库迁移 → 后端 storage 服务 → 后端 auth 扩展 → 后端路由注册 → 前端 hooks → 前端 UI。

## Technical Context

### 现有代码（扩展点）
- `backend/internal/model/user.go` — User 模型，已有 avatar 字段
- `backend/internal/handler/auth.go` — AuthHandler，已有 Register/Login/GetMe
- `backend/internal/service/auth/auth.go` — AuthService，已有 Register/Login/GetUserByID
- `backend/internal/service/auth/password.go` — HashPassword/VerifyPassword
- `backend/internal/service/crypto/crypto.go` — CryptoService，AES-256-GCM
- `backend/internal/repository/user_repo.go` — UserRepository 接口
- `backend/internal/repository/user_repo_impl.go` — GORM 实现
- `frontend/src/hooks/use-auth.ts` — useLogin/useRegister/useMe/useLogout
- `frontend/src/components/layout/sidebar.tsx` — 侧边栏，无用户信息
- `backend/internal/config/config.go` — Config 结构体，需新增 S3 配置

### 新增文件
- `backend/migrations/000018_add_user_profile_fields.up.sql`
- `backend/migrations/000018_add_user_profile_fields.down.sql`
- `backend/internal/service/storage/storage.go` — S3 兼容对象存储服务
- `backend/internal/service/storage/storage_test.go`
- `frontend/src/hooks/use-profile.ts` — useUpdateProfile/useUploadAvatar/useChangePassword
- `frontend/src/components/layout/user-menu.tsx` — 侧边栏用户菜单组件
- `frontend/src/app/(dashboard)/profile/page.tsx` — 个人资料页面

## Implementation Layers

### Layer 1: 数据库迁移
- 创建 migration 000018，User 表新增 nickname 和 phone_encrypted 字段

### Layer 2: 配置扩展
- Config 结构体新增 S3 配置字段（Endpoint/AccessKey/SecretKey/Bucket/Region）
- .env.example 补充 S3 配置项

### Layer 3: Storage 服务
- 实现 StorageService 接口（Upload/GetURL）
- 基于 AWS SDK S3 兼容协议
- 文件类型白名单校验（jpg/png/webp）
- 文件大小限制（2MB）
- UUID 重命名

### Layer 4: 后端 Auth 扩展
- UserRepository 接口新增 UpdateProfile/UpdatePassword/UpdateAvatar 方法
- AuthService 新增 UpdateProfile（phone 加密）、ChangePassword（验旧密码）、UpdateAvatar
- 手机号脱敏工具函数 MaskPhone
- GetMe 返回完整字段（avatar/nickname/phone 脱敏）

### Layer 5: 后端路由注册
- AuthHandler 新增 UpdateProfile/ChangePassword/UploadAvatar/Logout 方法
- main.go 初始化 StorageService，注入 AuthHandler
- 注册新路由到 auth 路由组

### Layer 6: 前端
- 新增 use-profile.ts hooks
- 新增 UserMenu 组件（侧边栏底部）
- 修改 Sidebar 集成 UserMenu
- 新增个人资料页面

## Project Structure Changes

```
backend/
  migrations/
    000018_add_user_profile_fields.up.sql    [NEW]
    000018_add_user_profile_fields.down.sql   [NEW]
  internal/
    config/config.go                          [MODIFY - 新增 S3 配置]
    model/user.go                             [MODIFY - 新增字段]
    repository/user_repo.go                   [MODIFY - 新增方法]
    repository/user_repo_impl.go              [MODIFY - 新增方法]
    service/storage/storage.go                [NEW]
    service/storage/storage_test.go           [NEW]
    service/auth/auth.go                      [MODIFY - 新增方法]
    service/auth/auth_test.go                 [MODIFY - 新增测试]
    service/auth/mask.go                      [NEW - 脱敏工具]
    service/auth/mask_test.go                 [NEW]
    handler/auth.go                           [MODIFY - 新增 handler]
    handler/auth_test.go                      [MODIFY - 新增测试]
  cmd/server/main.go                          [MODIFY - 初始化 storage]

frontend/
  src/
    hooks/use-profile.ts                      [NEW]
    hooks/use-auth.ts                         [MODIFY - useLogout 增强]
    components/layout/user-menu.tsx            [NEW]
    components/layout/sidebar.tsx              [MODIFY - 集成 UserMenu]
    app/(dashboard)/profile/page.tsx           [NEW]
```

## Dependencies

后端新增依赖：
- `github.com/aws/aws-sdk-go-v2` — S3 兼容对象存储
- `github.com/google/uuid` — 文件 UUID 重命名

## Security Checklist

- [x] 手机号 AES-256-GCM 加密存储（复用 CryptoService）
- [x] 手机号 API 脱敏返回（保留前3后4）
- [x] 头像上传文件类型白名单
- [x] 头像上传大小限制 2MB
- [x] 头像文件 UUID 重命名（防路径遍历）
- [x] 修改密码验证旧密码
- [x] S3 配置从环境变量加载，不硬编码
- [x] 日志不输出手机号明文
