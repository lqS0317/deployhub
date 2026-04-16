-- 构建记录（Kaniko 等）
CREATE TABLE builds (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id),
    trigger_user_id INTEGER NOT NULL REFERENCES users(id),
    git_branch VARCHAR(200) NOT NULL,
    git_commit VARCHAR(40) DEFAULT '',
    image_tag VARCHAR(200) DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    build_cluster_id INTEGER NOT NULL REFERENCES clusters(id),
    kaniko_job_name VARCHAR(200) DEFAULT '',
    log TEXT DEFAULT '',
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_builds_service ON builds (service_id);
CREATE INDEX idx_builds_trigger_user ON builds (trigger_user_id);
CREATE INDEX idx_builds_cluster ON builds (build_cluster_id);
