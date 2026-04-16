CREATE TABLE IF NOT EXISTS route_permissions (
    id SERIAL PRIMARY KEY,
    cluster_id INTEGER NOT NULL DEFAULT 0,
    user_id INTEGER NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(cluster_id, user_id)
);
