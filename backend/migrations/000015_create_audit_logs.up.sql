-- 审计日志（操作与资源引用）
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(30) DEFAULT '',
    resource_id INTEGER DEFAULT 0,
    detail JSONB DEFAULT '{}',
    ip_address VARCHAR(45) DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_user ON audit_logs (user_id);
CREATE INDEX idx_audit_resource ON audit_logs (resource_type, resource_id);
CREATE INDEX idx_audit_created ON audit_logs (created_at);
