-- 可部署服务：关联 Git、镜像仓库与目标集群
CREATE TABLE services (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(200) DEFAULT '',
    description TEXT DEFAULT '',
    git_repo_id INTEGER NOT NULL REFERENCES git_repos(id),
    git_branch VARCHAR(200) NOT NULL DEFAULT 'main',
    dockerfile_path VARCHAR(500) NOT NULL DEFAULT './Dockerfile',
    registry_id INTEGER NOT NULL REFERENCES registries(id),
    image_repo VARCHAR(500) NOT NULL,
    cluster_id INTEGER NOT NULL REFERENCES clusters(id),
    namespace VARCHAR(100) NOT NULL DEFAULT 'default',
    replicas INTEGER NOT NULL DEFAULT 1,
    cpu_request VARCHAR(20) DEFAULT '',
    mem_request VARCHAR(20) DEFAULT '',
    cpu_limit VARCHAR(20) DEFAULT '',
    mem_limit VARCHAR(20) DEFAULT '',
    port INTEGER NOT NULL,
    health_check_path VARCHAR(200) DEFAULT '',
    env_vars JSONB DEFAULT '{}',
    volumes JSONB DEFAULT '[]',
    owner_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_services_name ON services (name);
CREATE INDEX idx_services_cluster_namespace ON services (cluster_id, namespace);
CREATE INDEX idx_services_git_repo ON services (git_repo_id);
CREATE INDEX idx_services_registry ON services (registry_id);
CREATE INDEX idx_services_owner ON services (owner_id);
