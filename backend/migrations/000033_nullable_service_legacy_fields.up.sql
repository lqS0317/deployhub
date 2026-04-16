-- Service 旧字段改为可空（构建/部署配置已迁移到 Build/Deployment）
ALTER TABLE services ALTER COLUMN registry_id DROP NOT NULL;
ALTER TABLE services ALTER COLUMN image_repo DROP NOT NULL;
ALTER TABLE services ALTER COLUMN image_repo SET DEFAULT '';
ALTER TABLE services ALTER COLUMN port DROP NOT NULL;
ALTER TABLE services ALTER COLUMN port SET DEFAULT 0;
ALTER TABLE services ALTER COLUMN dockerfile_path DROP NOT NULL;
ALTER TABLE services ALTER COLUMN dockerfile_path SET DEFAULT '';
ALTER TABLE services ALTER COLUMN replicas SET DEFAULT 0;
