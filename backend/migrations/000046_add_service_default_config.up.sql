ALTER TABLE services ADD COLUMN default_env_vars JSONB DEFAULT '[]';
ALTER TABLE services ADD COLUMN default_volumes JSONB DEFAULT '[]';
ALTER TABLE services ADD COLUMN default_secret_refs JSONB DEFAULT '[]';
ALTER TABLE services ADD COLUMN default_config_map_refs JSONB DEFAULT '[]';
ALTER TABLE services ADD COLUMN default_service_account_name VARCHAR(100) DEFAULT '';
ALTER TABLE services ADD COLUMN default_liveness_probe JSONB DEFAULT '{}';
ALTER TABLE services ADD COLUMN default_readiness_probe JSONB DEFAULT '{}';
