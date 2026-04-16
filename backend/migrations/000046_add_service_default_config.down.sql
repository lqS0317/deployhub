ALTER TABLE services DROP COLUMN IF EXISTS default_readiness_probe;
ALTER TABLE services DROP COLUMN IF EXISTS default_liveness_probe;
ALTER TABLE services DROP COLUMN IF EXISTS default_service_account_name;
ALTER TABLE services DROP COLUMN IF EXISTS default_config_map_refs;
ALTER TABLE services DROP COLUMN IF EXISTS default_secret_refs;
ALTER TABLE services DROP COLUMN IF EXISTS default_volumes;
ALTER TABLE services DROP COLUMN IF EXISTS default_env_vars;
