# Feature 019: 配置中心多条目模式 + 部署零覆盖

## 概述

配置中心改为多配置条目(ConfigEntry)模式，服务通过 ServiceConfigRef 显式引用配置条目，Deployment 模型清理所有配置字段。

## 三大核心变更

### 1. 配置条目(ConfigEntry)替代 ServiceConfigEnv
每个服务每个环境下可创建多个配置条目，每个条目有独立的名称、类型(env/configmap/secret)、格式和发布版本。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | |
| service_id | uint FK(services) CASCADE | |
| cluster_id | uint | 环境 |
| name | varchar(100) | 条目名称 |
| config_type | varchar(20) | env / configmap / secret |
| format | varchar(20) | properties / yaml / json（env 固定 properties） |
| draft_content | text | yaml/json 草稿 |
| created_at, updated_at | timestamp | |

UNIQUE(service_id, cluster_id, name)

### 2. 服务配置引用(ServiceConfigRef)
服务显式声明引用哪些配置条目（按名称跨环境匹配）。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | |
| service_id | uint FK(services) CASCADE | |
| config_entry_name | varchar(100) | 配置条目名称（跨环境匹配） |
| mount_path | varchar(255) | 挂载路径（configmap/secret 用，env 为空） |
| created_at | timestamp | |

UNIQUE(service_id, config_entry_name)

### 3. ConfigItem/ConfigRelease 归属 ConfigEntry
- ConfigItem: config_entry_id FK(config_entries) — 归属具体条目
- ConfigRelease: config_entry_id FK(config_entries) — 归属具体条目
- ConfigPermission: 保持 service_id + cluster_id 不变

### 4. Deployment 字段清理
删除约 20 个配置相关字段，保留核心部署跟踪字段。

### 5. Service 字段清理
删除 DefaultEnvVars, DefaultSecretRefs, DefaultConfigMapRefs（已移到配置中心），保留运行时参数。

## API

```
# 配置条目
GET    /services/:id/config-entries?cluster_id=X
POST   /services/:id/config-entries
PUT    /config-entries/:id
DELETE /config-entries/:id

# 配置项（properties）
GET    /config-entries/:id/items
POST   /config-entries/:id/items
PUT    /config-entries/:id/items/:item_id
DELETE /config-entries/:id/items/:item_id

# 草稿（yaml/json）
GET    /config-entries/:id/draft
PUT    /config-entries/:id/draft

# 发布/回滚
POST   /config-entries/:id/release
POST   /config-entries/:id/rollback
GET    /config-entries/:id/releases

# 服务配置引用
GET    /services/:id/config-refs
POST   /services/:id/config-refs
PUT    /services/:id/config-refs/:ref_id
DELETE /services/:id/config-refs/:ref_id

# 权限（保持 service + cluster 粒度）
GET    /services/:id/configs/:cid/permissions
POST   /services/:id/configs/:cid/permissions
DELETE /services/:id/configs/:cid/permissions/:pid
```

## 部署流程

Direct 部署:
1. 从 Service 读取: DefaultReplicas, DefaultPort, DefaultCPU/Mem, DefaultProbes, DefaultCommand/Args, DefaultVolumes, DefaultVCTs, DefaultSA, DefaultWorkloadType
2. 从 ServiceConfigRef 获取引用列表
3. 按 (config_entry_name + cluster_id) 查找 ConfigEntry → 最新 published ConfigRelease
4. 按 config_type 生成:
   - env → envFrom (ConfigMap)
   - configmap → K8s ConfigMap + volumeMount
   - secret → K8s Secret + volumeMount
5. 合并生成完整 Deployment/StatefulSet YAML
6. Apply

部署对话框简化: 选服务 → 选集群+命名空间 → 选镜像来源 → 确认

## 前端

配置中心页面:
- 左侧: 环境列表
- 右侧: 配置条目列表（名称/类型/最新版本/操作）
- 点击条目 → 编辑面板（KV 表格或 YAML 编辑器 + 发布/回滚/历史）
- 新建条目弹窗

服务编辑页面:
- 新增"配置引用"分区
- 表格: 配置名称 | 挂载路径 | 操作

部署对话框:
- 移除 DirectConfigForm 和 YAML 编辑器
- 简化为基础字段

## 非目标

- Helm 配置管理
- 跨服务共享配置
- 运行时参数按环境差异化
- 配置热推送

## 风险

- Migration 复杂度（外键重建）
- Deployment 字段清理影响范围大
- ConfigEntry 名称匹配：环境缺失配置应报错
