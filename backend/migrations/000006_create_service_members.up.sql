-- 服务成员与角色（用户删除时级联清理）
CREATE TABLE service_members (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_service_member_unique ON service_members (service_id, user_id);
CREATE INDEX idx_service_members_service ON service_members (service_id);
CREATE INDEX idx_service_members_user ON service_members (user_id);
