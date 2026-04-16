-- 配置下发到集群的记录
CREATE TABLE config_deployments (
    id SERIAL PRIMARY KEY,
    config_version_id INTEGER NOT NULL REFERENCES config_versions(id),
    cluster_id INTEGER NOT NULL REFERENCES clusters(id),
    namespace VARCHAR(100) NOT NULL,
    resource_name VARCHAR(200) NOT NULL,
    status VARCHAR(10) NOT NULL DEFAULT 'pending',
    deployed_by_id INTEGER NOT NULL REFERENCES users(id),
    deployed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_config_deployments_version ON config_deployments (config_version_id);
CREATE INDEX idx_config_deployments_cluster ON config_deployments (cluster_id);
CREATE INDEX idx_config_deployments_deployer ON config_deployments (deployed_by_id);
