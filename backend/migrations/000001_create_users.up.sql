-- 用户表：本地账号与 OAuth 绑定
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255),
    oauth_provider VARCHAR(20),
    oauth_id VARCHAR(100),
    role VARCHAR(10) NOT NULL DEFAULT 'member',
    avatar VARCHAR(500) DEFAULT '',
    status VARCHAR(10) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_users_username ON users (username);
CREATE UNIQUE INDEX idx_users_email ON users (email);
-- OAuth 用户部分唯一索引（仅当 provider 与 oauth_id 均非空时生效）
CREATE UNIQUE INDEX idx_user_oauth_provider_id ON users (oauth_provider, oauth_id)
    WHERE oauth_provider IS NOT NULL AND oauth_id IS NOT NULL;
