-- 服务表新增 Helm 部署相关字段
ALTER TABLE services ADD COLUMN deploy_type VARCHAR(10) NOT NULL DEFAULT 'direct';
ALTER TABLE services ADD COLUMN helm_repo_id INTEGER REFERENCES git_repos(id);
ALTER TABLE services ADD COLUMN helm_chart_path VARCHAR(255) DEFAULT '';
ALTER TABLE services ADD COLUMN helm_values_path VARCHAR(255) DEFAULT '';
ALTER TABLE services ADD COLUMN helm_release_name VARCHAR(100) DEFAULT '';
ALTER TABLE services ADD COLUMN helm_chart_branch VARCHAR(100) DEFAULT 'main';
