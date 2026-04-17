-- DeployHub 全量初始化 SQL（全新部署用，等效于执行 000001-000068 所有迁移的最终态）
-- 适用于 PostgreSQL 14+

-- ==================== 用户与权限 ====================

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    oauth_provider VARCHAR(20),
    oauth_id VARCHAR(100),
    role VARCHAR(10) NOT NULL DEFAULT 'member',
    avatar VARCHAR(500) DEFAULT '',
    nickname VARCHAR(100) DEFAULT '',
    phone_encrypted TEXT DEFAULT '',
    status VARCHAR(10) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT DEFAULT '',
    created_by INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS group_members (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, user_id)
);

CREATE TABLE IF NOT EXISTS group_service_permissions (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    service_id INTEGER,
    role VARCHAR(20) NOT NULL DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, service_id)
);

-- ==================== 基础设施 ====================

CREATE TABLE IF NOT EXISTS clusters (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) DEFAULT '',
    env VARCHAR(20) NOT NULL,
    api_server VARCHAR(500) DEFAULT '',
    kubeconfig_encrypted TEXT NOT NULL,
    status VARCHAR(10) NOT NULL DEFAULT 'active',
    helm_service_account VARCHAR(100) DEFAULT '',
    build_service_account VARCHAR(100) DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cluster_namespaces (
    id SERIAL PRIMARY KEY,
    cluster_id INTEGER NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    namespace VARCHAR(100) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(cluster_id, namespace)
);

CREATE TABLE IF NOT EXISTS git_repositories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    url VARCHAR(500) NOT NULL,
    provider VARCHAR(20) NOT NULL DEFAULT 'github',
    auth_type VARCHAR(20) NOT NULL DEFAULT 'token',
    credential_encrypted TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS registries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    url VARCHAR(500) NOT NULL,
    provider VARCHAR(20) NOT NULL DEFAULT 'docker',
    auth_config_encrypted TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== 服务管理 ====================

CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) DEFAULT '',
    description TEXT DEFAULT '',
    git_repo_id INTEGER NOT NULL REFERENCES git_repositories(id),
    git_branch VARCHAR(200) NOT NULL DEFAULT 'main',
    dockerfile_path VARCHAR(500) NOT NULL DEFAULT './Dockerfile',
    registry_id INTEGER,
    image_repo VARCHAR(500) DEFAULT '',
    cluster_id INTEGER,
    namespace VARCHAR(100) DEFAULT '',
    replicas INTEGER NOT NULL DEFAULT 1,
    cpu_request VARCHAR(20) DEFAULT '',
    mem_request VARCHAR(20) DEFAULT '',
    cpu_limit VARCHAR(20) DEFAULT '',
    mem_limit VARCHAR(20) DEFAULT '',
    port INTEGER DEFAULT 0,
    health_check_path VARCHAR(200) DEFAULT '',
    env_vars JSONB,
    volumes JSONB,
    owner_id INTEGER NOT NULL REFERENCES users(id),
    service_type VARCHAR(50) DEFAULT '',
    language VARCHAR(50) DEFAULT '',
    language_version VARCHAR(50) DEFAULT '',
    deploy_type VARCHAR(10) NOT NULL DEFAULT 'direct',
    workload_type VARCHAR(20) NOT NULL DEFAULT 'deployment',
    helm_repo_id INTEGER,
    helm_chart_path VARCHAR(255) DEFAULT '',
    helm_values_path VARCHAR(255) DEFAULT '',
    helm_release_name VARCHAR(100) DEFAULT '',
    helm_chart_branch VARCHAR(100) DEFAULT 'main',
    helm_env_file_path VARCHAR(255) DEFAULT '',
    default_port INTEGER DEFAULT 0,
    default_replicas INTEGER DEFAULT 1,
    default_cpu_request VARCHAR(20) DEFAULT '',
    default_mem_request VARCHAR(20) DEFAULT '',
    default_cpu_limit VARCHAR(20) DEFAULT '',
    default_mem_limit VARCHAR(20) DEFAULT '',
    default_command JSONB DEFAULT '[]',
    default_args JSONB DEFAULT '[]',
    default_workload_type VARCHAR(20) DEFAULT 'deployment',
    default_liveness_probe JSONB DEFAULT '{}',
    default_readiness_probe JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS service_members (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    role VARCHAR(20) NOT NULL DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, user_id)
);

-- ==================== 构建 ====================

CREATE TABLE IF NOT EXISTS builds (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    trigger_user_id INTEGER NOT NULL REFERENCES users(id),
    git_branch VARCHAR(200) NOT NULL,
    git_commit VARCHAR(40) DEFAULT '',
    image_tag VARCHAR(200) DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    build_cluster_id INTEGER NOT NULL,
    kaniko_job_name VARCHAR(200) DEFAULT '',
    name VARCHAR(200) DEFAULT '',
    dockerfile_path VARCHAR(500) DEFAULT './Dockerfile',
    registry_id INTEGER,
    image_repo VARCHAR(500) DEFAULT '',
    build_context VARCHAR(500) DEFAULT '.',
    log TEXT DEFAULT '',
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== 部署 ====================

CREATE TABLE IF NOT EXISTS deployments (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    build_id INTEGER,
    trigger_user_id INTEGER NOT NULL REFERENCES users(id),
    cluster_id INTEGER NOT NULL,
    namespace VARCHAR(100) NOT NULL,
    image_tag VARCHAR(200) NOT NULL,
    status VARCHAR(20) NOT NULL,
    previous_image_tag VARCHAR(200) DEFAULT '',
    is_rollback BOOLEAN NOT NULL DEFAULT false,
    rollback_from_id INTEGER,
    helm_revision INTEGER,
    image_source VARCHAR(20) NOT NULL DEFAULT 'build',
    external_image VARCHAR(500) DEFAULT '',
    fail_reason TEXT DEFAULT '',
    preview_yaml TEXT,
    preview_summary JSONB,
    deploy_type VARCHAR(10) DEFAULT 'direct',
    workload_type VARCHAR(20) DEFAULT 'deployment',
    helm_repo_id INTEGER,
    helm_chart_path VARCHAR(255) DEFAULT '',
    helm_release_name VARCHAR(100) DEFAULT '',
    helm_chart_branch VARCHAR(100) DEFAULT 'main',
    helm_service_account VARCHAR(100) DEFAULT '',
    deploy_command TEXT DEFAULT '',
    pod_status VARCHAR(20) DEFAULT '',
    pod_message TEXT DEFAULT '',
    health_check_path VARCHAR(200) DEFAULT '',
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS approvals (
    id SERIAL PRIMARY KEY,
    deployment_id INTEGER NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    requester_id INTEGER NOT NULL REFERENCES users(id),
    approver_id INTEGER NOT NULL REFERENCES users(id),
    status VARCHAR(10) NOT NULL DEFAULT 'pending',
    comment TEXT DEFAULT '',
    decided_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== 配置中心 ====================

CREATE TABLE IF NOT EXISTS config_entries (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    config_type VARCHAR(20) NOT NULL DEFAULT 'configmap',
    format VARCHAR(20) NOT NULL DEFAULT 'properties',
    mount_path VARCHAR(255) DEFAULT '',
    draft_content TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, cluster_id, name)
);

CREATE TABLE IF NOT EXISTS config_items (
    id SERIAL PRIMARY KEY,
    config_entry_id INTEGER NOT NULL REFERENCES config_entries(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    comment VARCHAR(500) DEFAULT '',
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_config_items_unique_key ON config_items(config_entry_id, key) WHERE NOT is_deleted;

CREATE TABLE IF NOT EXISTS config_releases (
    id SERIAL PRIMARY KEY,
    config_entry_id INTEGER NOT NULL REFERENCES config_entries(id) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1,
    snapshot JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'published',
    comment VARCHAR(500) DEFAULT '',
    created_by_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_config_releases_lookup ON config_releases(config_entry_id, version DESC);

CREATE TABLE IF NOT EXISTS config_permissions (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL DEFAULT 0,
    user_id INTEGER NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, cluster_id, user_id)
);

-- ==================== 旧配置表（兼容，后续可删） ====================

CREATE TABLE IF NOT EXISTS config_templates (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    config_type VARCHAR(10) NOT NULL DEFAULT 'configmap',
    template_content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, name)
);

CREATE TABLE IF NOT EXISTS config_env_values (
    id SERIAL PRIMARY KEY,
    config_template_id INTEGER NOT NULL REFERENCES config_templates(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    values_encrypted TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(config_template_id, cluster_id)
);

CREATE TABLE IF NOT EXISTS config_versions (
    id SERIAL PRIMARY KEY,
    config_template_id INTEGER NOT NULL REFERENCES config_templates(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    rendered_content TEXT NOT NULL DEFAULT '',
    created_by_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS config_deployments (
    id SERIAL PRIMARY KEY,
    config_version_id INTEGER NOT NULL REFERENCES config_versions(id),
    cluster_id INTEGER NOT NULL,
    namespace VARCHAR(100) NOT NULL DEFAULT 'default',
    resource_name VARCHAR(200) NOT NULL DEFAULT '',
    status VARCHAR(10) NOT NULL DEFAULT 'pending',
    deployed_by_id INTEGER NOT NULL DEFAULT 0,
    deployed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ==================== Helm Values ====================

CREATE TABLE IF NOT EXISTS helm_values (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    version INTEGER NOT NULL DEFAULT 1,
    updated_by INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_id, cluster_id)
);

-- ==================== 通知 ====================

CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    type VARCHAR(30) NOT NULL DEFAULT 'deployment',
    title VARCHAR(200) NOT NULL,
    content TEXT DEFAULT '',
    is_read BOOLEAN NOT NULL DEFAULT false,
    reference_type VARCHAR(30) DEFAULT '',
    reference_id INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notification_channels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL,
    webhook_url VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notification_rules (
    id SERIAL PRIMARY KEY,
    channel_id INTEGER NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    event_type VARCHAR(30) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(channel_id, event_type)
);

CREATE TABLE IF NOT EXISTS service_notification_rules (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    channel_id INTEGER NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_snr_service ON service_notification_rules(service_id);

CREATE TABLE IF NOT EXISTS notification_logs (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    event_type VARCHAR(30) NOT NULL,
    title VARCHAR(255) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    status VARCHAR(10) NOT NULL DEFAULT 'sent',
    error_msg TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_notification_logs_service ON notification_logs(service_id);
CREATE INDEX idx_notification_logs_event ON notification_logs(event_type);
CREATE INDEX idx_notification_logs_created ON notification_logs(created_at);

-- ==================== 审计日志 ====================

CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id INTEGER,
    detail JSONB DEFAULT '{}',
    ip_address VARCHAR(45) DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);

-- ==================== 系统设置 ====================

CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    description VARCHAR(255) DEFAULT '',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO system_settings (key, value, description) VALUES
    ('helm_job_namespace', 'deployhub-jobs', 'Helm Runner Job 运行的命名空间'),
    ('env_values_map', '', '集群环境到 Helm values 文件后缀的映射')
ON CONFLICT (key) DO NOTHING;

-- ==================== 路由中心 ====================

CREATE TABLE IF NOT EXISTS route_entries (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    resource_type VARCHAR(20) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    created_by_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, resource_type)
);

CREATE TABLE IF NOT EXISTS route_deployments (
    id SERIAL PRIMARY KEY,
    route_entry_id INTEGER NOT NULL REFERENCES route_entries(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    namespace VARCHAR(100) NOT NULL DEFAULT 'default',
    status VARCHAR(20) NOT NULL DEFAULT 'deployed',
    config_snapshot JSONB DEFAULT '{}',
    rendered_yaml TEXT DEFAULT '',
    error_msg TEXT DEFAULT '',
    deployed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(route_entry_id, cluster_id, namespace)
);

CREATE TABLE IF NOT EXISTS route_permissions (
    id SERIAL PRIMARY KEY,
    cluster_id INTEGER NOT NULL DEFAULT 0,
    user_id INTEGER NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(cluster_id, user_id)
);

-- ==================== 插件中心 ====================

CREATE TABLE IF NOT EXISTS route_plugins (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    yaml_content TEXT DEFAULT '',
    created_by_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS plugin_deployments (
    id SERIAL PRIMARY KEY,
    plugin_id INTEGER NOT NULL REFERENCES route_plugins(id) ON DELETE CASCADE,
    cluster_id INTEGER NOT NULL,
    namespace VARCHAR(100) NOT NULL DEFAULT 'default',
    status VARCHAR(20) NOT NULL DEFAULT 'deployed',
    yaml_snapshot TEXT DEFAULT '',
    error_msg TEXT DEFAULT '',
    deployed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(plugin_id, cluster_id, namespace)
);
