ALTER TABLE deployments DROP COLUMN IF EXISTS external_image;
ALTER TABLE deployments DROP COLUMN IF EXISTS image_source;
ALTER TABLE services DROP COLUMN IF EXISTS helm_env_file_path;
