CREATE TABLE notifications (
    id          BIGSERIAL PRIMARY KEY,
    user_id     INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        SMALLINT NOT NULL CHECK (type BETWEEN 1 AND 11),
    channel     SMALLINT NOT NULL CHECK (channel BETWEEN 1 AND 5),
    status      SMALLINT NOT NULL DEFAULT 1 CHECK (status BETWEEN 1 AND 4),
    title       TEXT NOT NULL,
    body        TEXT NOT NULL,
    link        VARCHAR(512),
    metadata    JSONB,
    sent_at     TIMESTAMPTZ,
    read_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_user_time
ON notifications(user_id, created_at DESC);

CREATE INDEX idx_notifications_user_unread
ON notifications(user_id, created_at DESC)
WHERE status != 4;

CREATE INDEX idx_notifications_pending
ON notifications(created_at)
WHERE status = 1;

CREATE TABLE notification_settings (
    user_id           INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    email_enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    telegram_enabled  BOOLEAN NOT NULL DEFAULT TRUE,
    sms_enabled       BOOLEAN NOT NULL DEFAULT FALSE,
    push_enabled      BOOLEAN NOT NULL DEFAULT FALSE,
    muted_types       SMALLINT[] NOT NULL DEFAULT '{}',
    reminder_24h      BOOLEAN NOT NULL DEFAULT TRUE,
    reminder_2h       BOOLEAN NOT NULL DEFAULT TRUE,
    reminder_30m      BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trg_notification_settings_updated_at
BEFORE UPDATE ON notification_settings
FOR EACH ROW EXECUTE FUNCTION update_updated_at();