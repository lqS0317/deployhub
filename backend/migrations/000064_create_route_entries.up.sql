CREATE TABLE IF NOT EXISTS route_entries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    resource_type VARCHAR(20) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    created_by_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, resource_type)
);
