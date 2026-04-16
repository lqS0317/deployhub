DROP INDEX IF EXISTS idx_snr_service;
ALTER TABLE service_notification_rules ADD COLUMN event_type VARCHAR(30) NOT NULL DEFAULT '';
CREATE UNIQUE INDEX idx_snr_svc_ch_event ON service_notification_rules(service_id, channel_id, event_type);
