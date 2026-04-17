-- =====================================================================
-- 线上数据库补丁：同步 init.sql 与 GORM Model 的差异
-- 执行前请先备份数据库，执行顺序不可调换
-- =====================================================================

-- 1. users: password_hash 改为可空，role 默认值改为 'member'
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'member';

-- 2. groups: 新增 created_by 字段
ALTER TABLE groups ADD COLUMN IF NOT EXISTS created_by INTEGER NOT NULL DEFAULT 0;

-- 3. registries: 新增 provider 和 is_default 字段
ALTER TABLE registries ADD COLUMN IF NOT EXISTS provider VARCHAR(20) NOT NULL DEFAULT 'docker';
ALTER TABLE registries ADD COLUMN IF NOT EXISTS is_default BOOLEAN NOT NULL DEFAULT FALSE;

-- 4. deployments: 新增 health_check_path 字段
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS health_check_path VARCHAR(200) DEFAULT '';

-- 5. approvals: 新增 requester_id 和 decided_at，删除 updated_at
ALTER TABLE approvals ADD COLUMN IF NOT EXISTS requester_id INTEGER;
ALTER TABLE approvals ADD COLUMN IF NOT EXISTS decided_at TIMESTAMP;
-- 回填 requester_id（如果有历史数据，可设为 approver_id 或部署触发者）
-- UPDATE approvals SET requester_id = approver_id WHERE requester_id IS NULL;
-- 回填后再设置 NOT NULL：
-- ALTER TABLE approvals ALTER COLUMN requester_id SET NOT NULL;

-- 6. config_templates: 新增 config_type，删除 format
ALTER TABLE config_templates ADD COLUMN IF NOT EXISTS config_type VARCHAR(10) NOT NULL DEFAULT 'configmap';
-- format 列可保留不删，不影响运行

-- 7. config_env_values: 列名重命名 template_id -> config_template_id, variables -> values_encrypted
ALTER TABLE config_env_values RENAME COLUMN template_id TO config_template_id;
ALTER TABLE config_env_values RENAME COLUMN variables TO values_encrypted;
ALTER TABLE config_env_values ALTER COLUMN values_encrypted TYPE TEXT USING values_encrypted::TEXT;
ALTER TABLE config_env_values ALTER COLUMN values_encrypted SET NOT NULL;
ALTER TABLE config_env_values ALTER COLUMN values_encrypted SET DEFAULT '';

-- 8. config_versions: 列名重命名 template_id -> config_template_id
ALTER TABLE config_versions RENAME COLUMN template_id TO config_template_id;

-- 9. config_deployments: 列名重命名 + 新增字段
ALTER TABLE config_deployments RENAME COLUMN version_id TO config_version_id;
ALTER TABLE config_deployments ADD COLUMN IF NOT EXISTS resource_name VARCHAR(200) NOT NULL DEFAULT '';
ALTER TABLE config_deployments ADD COLUMN IF NOT EXISTS deployed_by_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE config_deployments ADD COLUMN IF NOT EXISTS deployed_at TIMESTAMP;
-- resource_type 列可保留不删，不影响运行

-- 10. helm_values: 新增 updated_by 字段
ALTER TABLE helm_values ADD COLUMN IF NOT EXISTS updated_by INTEGER;

-- 11. notifications: 新增 reference_type 和 reference_id 字段
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS reference_type VARCHAR(30) DEFAULT '';
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS reference_id INTEGER DEFAULT 0;

-- 12. services: 删除 default_volume_claim_templates（Model 中无此字段，可保留不删）
-- 如确认不需要可执行：
-- ALTER TABLE services DROP COLUMN IF EXISTS default_volume_claim_templates;
