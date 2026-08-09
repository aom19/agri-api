-- Audit log: stores every mutation for traceability
CREATE TABLE IF NOT EXISTS audit_log (
    id          BIGSERIAL PRIMARY KEY,
    entity_type VARCHAR(50)  NOT NULL,
    entity_id   VARCHAR(50)  NOT NULL,
    action      VARCHAR(30)  NOT NULL,
    actor_id    BIGINT       REFERENCES users(id),
    changes     JSONB,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_actor  ON audit_log(actor_id);
CREATE INDEX idx_audit_log_time   ON audit_log(created_at DESC);

-- Notifications: actionable alerts for admin/manager
CREATE TABLE IF NOT EXISTS notifications (
    id          BIGSERIAL    PRIMARY KEY,
    type        VARCHAR(50)  NOT NULL,
    title       VARCHAR(255) NOT NULL,
    message     TEXT         NOT NULL DEFAULT '',
    entity_type VARCHAR(50),
    entity_id   VARCHAR(50),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_time ON notifications(created_at DESC);

-- Per-user notification delivery and read state
CREATE TABLE IF NOT EXISTS user_notifications (
    id              BIGSERIAL   PRIMARY KEY,
    notification_id BIGINT      NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    user_id         BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_notifications_user    ON user_notifications(user_id, read_at);
CREATE INDEX idx_user_notifications_notif   ON user_notifications(notification_id);

-- Permissions for viewing notifications
INSERT INTO permissions (name, description) VALUES
    ('notifications:read', 'Vizualizare notificări proprii'),
    ('audit:read', 'Vizualizare jurnal de audit')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code IN ('admin', 'manager', 'operator')
  AND p.name = 'notifications:read'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'admin'
  AND p.name = 'audit:read'
ON CONFLICT DO NOTHING;
