-- 服务元信息字段
ALTER TABLE services ADD COLUMN service_type VARCHAR(50) DEFAULT '';
ALTER TABLE services ADD COLUMN language VARCHAR(50) DEFAULT '';
ALTER TABLE services ADD COLUMN language_version VARCHAR(50) DEFAULT '';
