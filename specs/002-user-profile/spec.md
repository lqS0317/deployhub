# Spec: 用户资料管理与登录登出增强

## Summary

为 DeployHub 补全用户体验基础能力：侧边栏显示当前登录用户信息及登出入口、个人资料编辑（头像/昵称/手机号）、修改密码。在现有 auth 模块上扩展，新增 storage 服务支持头像上传到 S3 兼容对象存储。

## User Stories

### P0 — 必须实现

1. **US-01** 作为已登录用户，我希望在侧边栏看到自己的头像和用户名，以确认当前登录身份
2. **US-02** 作为已登录用户，我希望能一键登出，以安全退出系统
3. **US-03** 作为已登录用户，我希望能编辑昵称和手机号，以完善个人资料
4. **US-04** 作为已登录用户，我希望能上传头像，以个性化自己的账户
5. **US-05** 作为已登录用户，我希望能修改密码（验证旧密码），以保障账户安全

### P1 — 应该实现

6. **US-06** 作为 API 调用方，我希望 GET /auth/me 返回完整用户信息（含脱敏手机号），以支持前端显示
7. **US-07** 作为系统，手机号必须 AES-256-GCM 加密存储，API 返回时脱敏（138****1234）

## Edge Cases

1. 上传非图片文件 → 返回 400，提示仅支持 jpg/png/webp
2. 上传超过 2MB 文件 → 返回 400，提示文件过大
3. 修改密码时旧密码错误 → 返回 401
4. 手机号为空 → 允许（nullable），返回空字符串
5. 对象存储未配置 → 头像上传返回 500，提示对象存储未配置
6. 昵称超长（>100字符）→ 返回 400

## Functional Requirements

| ID | 要求 | 优先级 |
|----|------|--------|
| FR-01 | User 表新增 nickname(varchar(100)) 和 phone_encrypted(text) 字段 | P0 |
| FR-02 | GET /auth/me 返回 id, username, email, role, status, avatar, nickname, phone(脱敏), created_at | P0 |
| FR-03 | PUT /auth/profile 更新 nickname 和 phone（phone 加密存储） | P0 |
| FR-04 | PUT /auth/password 修改密码，必须验证旧密码 | P0 |
| FR-05 | POST /auth/avatar 上传头像，类型白名单 jpg/png/webp，大小 ≤2MB，UUID 重命名 | P0 |
| FR-06 | storage 服务封装 S3 兼容上传，配置从环境变量加载 | P0 |
| FR-07 | 侧边栏底部显示用户头像+用户名+下拉菜单（个人资料/登出） | P0 |
| FR-08 | 个人资料页面/弹窗：头像上传、昵称编辑、手机号编辑、修改密码 | P0 |
| FR-09 | 前端 useMe hook 集成到侧边栏 | P0 |
| FR-10 | 新增 useUpdateProfile、useUploadAvatar、useChangePassword hooks | P0 |
| FR-11 | 手机号脱敏规则：保留前3后4，中间 **** | P1 |

## Data Model Changes

### User 表变更（migration 000018）

新增字段：
- `nickname VARCHAR(100) DEFAULT ''` — 用户昵称
- `phone_encrypted TEXT DEFAULT ''` — 手机号（AES-256-GCM 加密存储）

已有字段沿用：
- `avatar VARCHAR(500)` — 改为存储对象存储 URL

## API Contracts

### 修改：GET /api/v1/auth/me
```json
{
  "id": 1,
  "username": "sunyu",
  "email": "sunyu@example.com",
  "role": "admin",
  "status": "active",
  "avatar": "https://s3.example.com/avatars/uuid.jpg",
  "nickname": "孙宇",
  "phone": "138****1234",
  "created_at": "2026-04-07T00:00:00Z"
}
```

### 新增：PUT /api/v1/auth/profile
Request: `{ "nickname": "孙宇", "phone": "13812341234" }`
Response: `{ "message": "资料已更新" }`

### 新增：PUT /api/v1/auth/password
Request: `{ "old_password": "xxx", "new_password": "yyy" }`
Response: `{ "message": "密码已修改" }`

### 新增：POST /api/v1/auth/avatar
Request: `multipart/form-data`, field `avatar`
Response: `{ "avatar_url": "https://s3.example.com/avatars/uuid.jpg" }`

### 新增：POST /api/v1/auth/logout
Request: 无 body
Response: `{ "message": "已登出" }`

## Success Criteria

1. 侧边栏显示当前用户头像和用户名
2. 点击登出后 token 清除，跳转登录页
3. 能成功上传头像并在侧边栏和资料页显示
4. 能修改昵称和手机号，手机号在 DB 中加密存储
5. 能修改密码，旧密码错误时返回错误提示
6. GET /auth/me 返回完整信息，手机号脱敏

## Assumptions

1. 用户已有 S3 兼容对象存储服务可用（AWS S3 / MinIO / 阿里云 OSS）
2. 手机号仅存储字符串，不做国际格式校验
3. 头像不做裁剪，原图上传
4. 登出采用前端无状态方案，不做服务端 token 黑名单

## Non-Goals

- OAuth 登录流程
- 手机号短信验证码
- 头像裁剪
- 服务端 token 黑名单
- 邮箱修改
