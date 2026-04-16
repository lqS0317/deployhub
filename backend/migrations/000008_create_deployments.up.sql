-- 部署记录（含回滚关联）
CREATE TABLE deployments (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id),
    build_id INTEGER REFERENCES builds(id),
    trigger_user_id INTEGER NOT NULL REFERENCES users(id),
    cluster_id INTEGER NOT NULL REFERENCES clusters(id),
    namespace VARCHAR(100) NOT NULL,
    image_tag VARCHAR(200) NOT NULL,
    replicas INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending_approval',
    previous_image_tag VARCHAR(200) DEFAULT '',
    is_rollback BOOLEAN NOT NULL DEFAULT FALSE,
    rollback_from_id INTEGER REFERENCES deployments(id),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deployments_service ON deployments (service_id);
CREATE INDEX idx_deployments_cluster ON deployments (cluster_id);
CREATE INDEX idx_deployments_trigger_user ON deployments (trigger_user_id);
CREATE INDEX idx_deployments_build ON deployments (build_id);
CREATE INDEX idx_deployments_rollback ON deployments (rollback_from_id);
