ALTER TABLE deployments ADD COLUMN pod_status VARCHAR(20) DEFAULT '';
ALTER TABLE deployments ADD COLUMN pod_message TEXT DEFAULT '';
