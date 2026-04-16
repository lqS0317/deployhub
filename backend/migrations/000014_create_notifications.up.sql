-- 用户站内通知
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    type VARCHAR(30) NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT DEFAULT '',
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    reference_type VARCHAR(30) DEFAULT '',
    reference_id INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications (user_id);
CREATE INDEX idx_notifications_user_read ON notifications (user_id, is_read);
