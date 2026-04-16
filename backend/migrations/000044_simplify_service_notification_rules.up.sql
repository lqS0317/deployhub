-- 去掉 event_type 列，简化为一个服务一条记录一个渠道
ALTER TABLE service_notification_rules DROP CONSTRAINT IF EXISTS idx_snr_svc_ch_event;
ALTER TABLE service_notification_rules DROP CONSTRAINT IF EXISTS service_notification_rules_service_id_channel_id_event_type_key;
ALTER TABLE service_notification_rules DROP COLUMN IF EXISTS event_type;
-- 一个服务只绑定一个渠道
CREATE UNIQUE INDEX IF NOT EXISTS idx_snr_service ON service_notification_rules(service_id);
