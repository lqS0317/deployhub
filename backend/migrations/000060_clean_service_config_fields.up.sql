ALTER TABLE services DROP COLUMN IF EXISTS default_env_vars;
ALTER TABLE services DROP COLUMN IF EXISTS default_secret_refs;
ALTER TABLE services DROP COLUMN IF EXISTS default_config_map_refs;
