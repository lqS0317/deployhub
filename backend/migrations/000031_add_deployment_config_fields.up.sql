-- 部署配置字段（从 Service 迁移）
ALTER TABLE deployments ADD COLUMN deploy_type VARCHAR(10) DEFAULT 'direct';
ALTER TABLE deployments ADD COLUMN workload_type VARCHAR(20) DEFAULT 'deployment';
ALTER TABLE deployments ADD COLUMN port INTEGER;
ALTER TABLE deployments ADD COLUMN cpu_request VARCHAR(20) DEFAULT '';
ALTER TABLE deployments ADD COLUMN mem_request VARCHAR(20) DEFAULT '';
ALTER TABLE deployments ADD COLUMN cpu_limit VARCHAR(20) DEFAULT '';
ALTER TABLE deployments ADD COLUMN mem_limit VARCHAR(20) DEFAULT '';
ALTER TABLE deployments ADD COLUMN health_check_path VARCHAR(200) DEFAULT '';
ALTER TABLE deployments ADD COLUMN helm_repo_id INTEGER REFERENCES git_repos(id);
ALTER TABLE deployments ADD COLUMN helm_chart_path VARCHAR(255) DEFAULT '';
ALTER TABLE deployments ADD COLUMN helm_release_name VARCHAR(100) DEFAULT '';
ALTER TABLE deployments ADD COLUMN helm_chart_branch VARCHAR(100) DEFAULT 'main';
