CREATE TABLE group_service_permissions (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    service_id INTEGER NOT NULL REFERENCES services(id),
    role VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, service_id)
);

CREATE INDEX idx_group_service_perms_service ON group_service_permissions(service_id);
