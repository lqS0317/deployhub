-- 服务表新增 env 文件路径
ALTER TABLE services ADD COLUMN helm_env_file_path VARCHAR(255) DEFAULT '';

-- 部署表新增镜像来源和外部镜像地址
ALTER TABLE deployments ADD COLUMN image_source VARCHAR(20) NOT NULL DEFAULT 'build';
ALTER TABLE deployments ADD COLUMN external_image VARCHAR(500) DEFAULT '';
