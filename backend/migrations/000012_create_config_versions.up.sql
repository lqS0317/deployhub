-- 配置渲染版本历史
CREATE TABLE config_versions (
    id SERIAL PRIMARY KEY,
    config_template_id INTEGER NOT NULL REFERENCES config_templates(id),
    cluster_id INTEGER NOT NULL REFERENCES clusters(id),
    version INTEGER NOT NULL,
    rendered_content TEXT NOT NULL,
    created_by_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_config_version_tpl_cluster_ver ON config_versions (config_template_id, cluster_id, version);
CREATE INDEX idx_config_versions_created_by ON config_versions (created_by_id);
