ALTER TABLE services ADD COLUMN default_env_vars JSONB DEFAULT '[]';
ALTER TABLE services ADD COLUMN default_secret_refs JSONB DEFAULT '[]';
ALTER TABLE services ADD COLUMN default_config_map_refs JSONB DEFAULT '[]';
