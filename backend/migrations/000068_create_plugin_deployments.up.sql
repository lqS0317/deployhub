CREATE TABLE IF NOT EXISTS plugin_deployments (
    id SERIAL PRIMARY KEY,
    plugin_id INTEGER NOT NULL REFERENCES route_plugins(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    namespace VARCHAR(100) NOT NULL DEFAULT 'default',
    status VARCHAR(20) NOT NULL DEFAULT 'deployed',
    yaml_snapshot TEXT DEFAULT '',
    error_msg TEXT DEFAULT '',
    deployed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_id, cluster_id, namespace)
);
