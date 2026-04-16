ALTER TABLE services ADD COLUMN default_volumes JSONB DEFAULT '[]';
ALTER TABLE services ADD COLUMN default_volume_claim_templates JSONB DEFAULT '[]';
ALTER TABLE services ADD COLUMN default_service_account_name VARCHAR(100) DEFAULT '';
