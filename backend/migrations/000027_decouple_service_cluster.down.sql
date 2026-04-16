UPDATE services SET cluster_id = 0 WHERE cluster_id IS NULL;
UPDATE services SET namespace = 'default' WHERE namespace IS NULL OR namespace = '';
ALTER TABLE services ALTER COLUMN cluster_id SET NOT NULL;
ALTER TABLE services ALTER COLUMN namespace SET NOT NULL;
