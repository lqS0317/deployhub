CREATE TABLE IF NOT EXISTS config_releases (
    id SERIAL PRIMARY KEY,
    namespace_id INTEGER NOT NULL REFERENCES config_namespaces(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    snapshot JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'published',
    comment VARCHAR(500) DEFAULT '',
    created_by_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_config_releases_lookup ON config_releases(namespace_id, cluster_id, version DESC);
