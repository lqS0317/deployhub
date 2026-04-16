CREATE TABLE IF NOT EXISTS config_permissions (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL DEFAULT 0,
    user_id INTEGER NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, cluster_id, user_id)
);
