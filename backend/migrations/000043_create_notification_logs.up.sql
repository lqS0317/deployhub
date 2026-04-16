CREATE TABLE IF NOT EXISTS notification_logs (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    event_type VARCHAR(30) NOT NULL,
    title VARCHAR(255) NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    status VARCHAR(10) NOT NULL DEFAULT 'sent',
    error_msg TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_logs_service ON notification_logs(service_id);
CREATE INDEX idx_notification_logs_event ON notification_logs(event_type);
CREATE INDEX idx_notification_logs_created ON notification_logs(created_at);
