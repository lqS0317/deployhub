CREATE TABLE IF NOT EXISTS service_config_refs (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    config_entry_name VARCHAR(100) NOT NULL,
    mount_path VARCHAR(255) DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, config_entry_name)
);
