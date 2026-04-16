CREATE TABLE IF NOT EXISTS config_items (
    id SERIAL PRIMARY KEY,
    namespace_id INTEGER NOT NULL REFERENCES config_namespaces(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    comment VARCHAR(500) DEFAULT '',
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_config_items_unique_key ON config_items(namespace_id, cluster_id, key) WHERE NOT is_deleted;
CREATE INDEX idx_config_items_ns_cluster ON config_items(namespace_id, cluster_id);
