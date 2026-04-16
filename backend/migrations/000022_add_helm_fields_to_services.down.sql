ALTER TABLE services DROP COLUMN IF EXISTS helm_chart_branch;
ALTER TABLE services DROP COLUMN IF EXISTS helm_release_name;
ALTER TABLE services DROP COLUMN IF EXISTS helm_values_path;
ALTER TABLE services DROP COLUMN IF EXISTS helm_chart_path;
ALTER TABLE services DROP COLUMN IF EXISTS helm_repo_id;
ALTER TABLE services DROP COLUMN IF EXISTS deploy_type;
