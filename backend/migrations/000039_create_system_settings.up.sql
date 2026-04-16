CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    description VARCHAR(255) DEFAULT '',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 初始化默认值
INSERT INTO system_settings (key, value, description) VALUES
    ('helm_job_namespace', 'deployhub-jobs', 'Helm Runner Job 运行的命名空间'),
    ('env_values_map', '', '集群环境到 Helm values 文件后缀的映射，格式: qanet:qa,testnet:testnet,mainnet:mainnet')
ON CONFLICT (key) DO NOTHING;
