CREATE TABLE IF NOT EXISTS config_releases (
    id SERIAL PRIMARY KEY,
    config_entry_id INTEGER NOT NULL REFERENCES config_entries(id) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1,
    snapshot JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'published',
    comment VARCHAR(500) DEFAULT '',
    created_by_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_config_releases_lookup ON config_releases(config_entry_id, version DESC);
