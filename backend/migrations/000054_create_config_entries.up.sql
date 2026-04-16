-- 清理旧配置中心表
DROP TABLE IF EXISTS config_permissions;
DROP TABLE IF EXISTS config_releases;
DROP TABLE IF EXISTS config_items;
DROP TABLE IF EXISTS service_config_envs;

-- 配置条目（一个服务一个环境下可有多个条目）
CREATE TABLE IF NOT EXISTS config_entries (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    config_type VARCHAR(20) NOT NULL DEFAULT 'configmap',
    format VARCHAR(20) NOT NULL DEFAULT 'properties',
    draft_content TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, cluster_id, name)
);
