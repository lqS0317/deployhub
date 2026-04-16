-- Git 仓库与凭据（凭据加密存储）
CREATE TABLE git_repos (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    url VARCHAR(500) NOT NULL,
    provider VARCHAR(20) NOT NULL,
    auth_type VARCHAR(10) NOT NULL,
    credential_encrypted TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_git_repos_name ON git_repos (name);
