-- 配置模板（按服务维度）
CREATE TABLE config_templates (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id),
    name VARCHAR(100) NOT NULL,
    config_type VARCHAR(10) NOT NULL,
    template_content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_config_template_service_name ON config_templates (service_id, name);
