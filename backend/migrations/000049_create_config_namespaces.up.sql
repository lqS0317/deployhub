CREATE TABLE IF NOT EXISTS config_namespaces (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    format VARCHAR(20) NOT NULL DEFAULT 'properties',
    config_type VARCHAR(20) NOT NULL DEFAULT 'configmap',
    description TEXT DEFAULT '',
    draft_content TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, name)
);
