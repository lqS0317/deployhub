-- 各集群下配置值（加密存储；模板删除时级联）
CREATE TABLE config_env_values (
    id SERIAL PRIMARY KEY,
    config_template_id INTEGER NOT NULL REFERENCES config_templates(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL REFERENCES clusters(id),
    values_encrypted TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_config_env_tpl_cluster ON config_env_values (config_template_id, cluster_id);
