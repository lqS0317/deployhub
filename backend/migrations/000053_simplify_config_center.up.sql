-- 删除旧的配置中心表（Feature 018 原始版本）
DROP TABLE IF EXISTS config_permissions;
DROP TABLE IF EXISTS config_releases;
DROP TABLE IF EXISTS config_items;
DROP TABLE IF EXISTS config_namespaces;

-- 新建服务环境配置元数据表
CREATE TABLE IF NOT EXISTS service_config_envs (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    format VARCHAR(20) NOT NULL DEFAULT 'properties',
    config_type VARCHAR(20) NOT NULL DEFAULT 'configmap',
    draft_content TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, cluster_id)
);

-- 重建配置项表（service_id 维度）
CREATE TABLE IF NOT EXISTS config_items (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    comment VARCHAR(500) DEFAULT '',
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_config_items_unique_key ON config_items(service_id, cluster_id, key) WHERE NOT is_deleted;
CREATE INDEX idx_config_items_svc_cluster ON config_items(service_id, cluster_id);

-- 重建发布记录表（service_id 维度）
CREATE TABLE IF NOT EXISTS config_releases (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    snapshot JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'published',
    comment VARCHAR(500) DEFAULT '',
    created_by_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_config_releases_lookup ON config_releases(service_id, cluster_id, version DESC);

-- 重建权限表（service_id 维度）
CREATE TABLE IF NOT EXISTS config_permissions (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL DEFAULT 0,
    user_id INTEGER NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, cluster_id, user_id)
);
