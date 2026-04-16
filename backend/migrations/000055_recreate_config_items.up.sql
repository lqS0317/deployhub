CREATE TABLE IF NOT EXISTS config_items (
    id SERIAL PRIMARY KEY,
    config_entry_id INTEGER NOT NULL REFERENCES config_entries(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    comment VARCHAR(500) DEFAULT '',
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_config_items_unique_key ON config_items(config_entry_id, key) WHERE NOT is_deleted;
