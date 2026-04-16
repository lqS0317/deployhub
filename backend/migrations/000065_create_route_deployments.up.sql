CREATE TABLE IF NOT EXISTS route_deployments (
    id SERIAL PRIMARY KEY,
    route_entry_id INTEGER NOT NULL REFERENCES route_entries(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    namespace VARCHAR(100) NOT NULL DEFAULT 'default',
    status VARCHAR(20) NOT NULL DEFAULT 'deployed',
    config_snapshot JSONB DEFAULT '{}',
    rendered_yaml TEXT DEFAULT '',
    error_msg TEXT DEFAULT '',
    deployed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(route_entry_id, cluster_id, namespace)
);
