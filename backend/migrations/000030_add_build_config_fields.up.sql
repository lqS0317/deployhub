-- 构建配置字段（从 Service 迁移）
ALTER TABLE builds ADD COLUMN name VARCHAR(200) DEFAULT '';
ALTER TABLE builds ADD COLUMN dockerfile_path VARCHAR(500) DEFAULT './Dockerfile';
ALTER TABLE builds ADD COLUMN registry_id INTEGER REFERENCES registries(id);
ALTER TABLE builds ADD COLUMN image_repo VARCHAR(500) DEFAULT '';
ALTER TABLE builds ADD COLUMN build_context VARCHAR(500) DEFAULT '.';
