# Feature 018: 配置中心改造 — Apollo 风格配置管理

## 概述

将 DeployHub 配置中心从 Go template 渲染模式改造为类 Apollo 配置管理平台。采用「服务 → 配置集(Namespace) → 多环境(Cluster) → Key-Value/YAML/JSON 配置项」的层级结构，支持 properties 表格编辑、YAML/JSON 代码编辑器，发布/回滚版本化管理。

## 动机

1. 多环境管理清晰（左侧环境列表一键切换）
2. Key-Value 表格编辑门槛低，YAML/JSON 满足复杂场景
3. 发布/回滚版本化，每次变更可追溯
4. 细粒度权限控制（配置集+环境级别）
5. 配置与部署解耦，配置变更不直接影响运行中服务

## 数据模型

### ConfigNamespace（配置集）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | |
| service_id | uint FK(services) | 归属服务 |
| name | varchar(100) | 配置集名称（同服务下唯一） |
| format | varchar(20) | properties / yaml / json |
| config_type | varchar(20) | configmap / secret |
| description | text | 描述 |
| draft_content | text | YAML/JSON 格式的未发布草稿内容 |
| created_at | timestamp | |
| updated_at | timestamp | |

唯一索引: (service_id, name)

### ConfigItem（配置项，仅 properties 格式）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | |
| namespace_id | uint FK(config_namespaces) | |
| cluster_id | uint FK(clusters) | 环境 |
| key | varchar(255) | 配置键 |
| value | text | 值（secret 类型 AES-256-GCM 加密） |
| comment | varchar(500) | 备注 |
| is_deleted | bool | 软删除标记 |
| created_at | timestamp | |
| updated_at | timestamp | |

唯一索引: (namespace_id, cluster_id, key) WHERE is_deleted = false

### ConfigRelease（发布记录）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | |
| namespace_id | uint FK | |
| cluster_id | uint FK | |
| version | int | 自增版本号 |
| snapshot | jsonb | 全量配置快照 |
| status | varchar(20) | published / rolled_back |
| comment | varchar(500) | 发布说明 |
| created_by_id | uint FK(users) | |
| created_at | timestamp | |

索引: (namespace_id, cluster_id, version DESC)

### ConfigPermission（配置权限）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | |
| namespace_id | uint FK | |
| cluster_id | uint FK | 0 = 所有环境 |
| user_id | uint FK(users) | |
| role | varchar(20) | viewer / editor / publisher |
| created_at | timestamp | |

唯一索引: (namespace_id, cluster_id, user_id)

## 工作流

### 编辑
- properties 格式: 通过 ConfigItem CRUD 操作
- yaml/json 格式: 保存到 ConfigNamespace.draft_content

### 发布
1. 收集当前配置（properties 从 ConfigItem 聚合，yaml/json 从 draft_content 读取）
2. 生成全量 JSON 快照
3. 写入 ConfigRelease（version 自增，status=published）
4. properties 格式清除 is_deleted=true 的记录

### 回滚
1. 选择历史版本
2. 恢复快照内容（properties 写回 ConfigItem，yaml/json 写回 draft_content）
3. 生成新 ConfigRelease（status=rolled_back，关联原版本）

### 部署集成
- 部署时查询服务所有配置集的最新 published 版本
- 每个配置集生成一个 K8s ConfigMap 或 Secret（基于 config_type）
- ConfigMap/Secret 名称: `{service-name}-{namespace-name}`
- Secret 值在写入 K8s 前解密

## 权限规则

| 角色 | 查看 | 编辑 | 发布/回滚 |
|------|------|------|-----------|
| viewer | ✓ (Secret 脱敏) | ✗ | ✗ |
| editor | ✓ | ✓ | ✗ |
| publisher | ✓ | ✓ | ✓ |
| Admin | ✓ | ✓ | ✓ (全局) |
| 服务 Owner | ✓ | ✓ | ✓ (该服务) |

## API 路由

```
# 配置集
GET    /services/:id/config-namespaces
POST   /services/:id/config-namespaces
PUT    /config-namespaces/:id
DELETE /config-namespaces/:id

# 配置项（properties 格式）
GET    /config-namespaces/:id/clusters/:cid/items
POST   /config-namespaces/:id/clusters/:cid/items
PUT    /config-namespaces/:id/clusters/:cid/items/:item_id
DELETE /config-namespaces/:id/clusters/:cid/items/:item_id

# 草稿（yaml/json 格式）
GET    /config-namespaces/:id/clusters/:cid/draft
PUT    /config-namespaces/:id/clusters/:cid/draft

# 发布与回滚
POST   /config-namespaces/:id/clusters/:cid/release
POST   /config-namespaces/:id/clusters/:cid/rollback
GET    /config-namespaces/:id/clusters/:cid/releases

# 权限
GET    /config-namespaces/:id/permissions
POST   /config-namespaces/:id/permissions
DELETE /config-namespaces/:id/permissions/:perm_id
```

## 前端页面（参照 Apollo）

### 布局
- 左侧面板: 环境列表（Cluster tabs）+ 配置集列表（可点击切换）
- 右侧主区域: 三个 Tab
  - 配置项: properties 表格 / YAML/JSON 编辑器
  - 更改历史: 未发布的修改列表
  - 发布历史: 版本列表 + 快照查看 + 回滚按钮
- 顶部操作栏: 发布按钮（显示待发布数）、回滚、授权

### 交互
- 发布弹窗: 显示变更 diff（新增/修改/删除）+ 发布说明输入
- 回滚弹窗: 版本列表 + 快照预览 + 确认回滚
- Secret 值: 前端以 `******` 显示，点击"显示"后解密展示（需 publisher 权限）

## 非目标

- 配置热推送（需客户端 SDK）
- 灰度发布（部分实例用新配置）
- 配置模板继承
- 外部配置中心集成

## 风险

- 现有 4 个 Config 模型数据迁移
- 部署流程中旧 Config 引用同步更新
- properties Key 命名冲突检测
- YAML/JSON 前端语法校验
