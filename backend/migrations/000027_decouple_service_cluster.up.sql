-- Service 与集群解耦：cluster_id 和 namespace 改为可空
ALTER TABLE services ALTER COLUMN cluster_id DROP NOT NULL;
ALTER TABLE services ALTER COLUMN namespace DROP NOT NULL;
ALTER TABLE services ALTER COLUMN namespace SET DEFAULT '';
