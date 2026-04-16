-- Kubernetes 集群配置（kubeconfig 加密存储）
CREATE TABLE clusters (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100) DEFAULT '',
    env VARCHAR(10) NOT NULL,
    api_server VARCHAR(500) DEFAULT '',
    kubeconfig_encrypted TEXT NOT NULL,
    status VARCHAR(10) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_clusters_name ON clusters (name);
