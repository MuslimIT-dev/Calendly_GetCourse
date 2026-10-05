DROP TRIGGER IF EXISTS trg_notification_settings_updated_at ON notification_settings;
DROP TABLE IF EXISTS notification_settings CASCADE;

DROP INDEX IF EXISTS idx_notifications_pending;
DROP INDEX IF EXISTS idx_notifications_user_unread;
DROP INDEX IF EXISTS idx_notifications_user_time;
DROP TABLE IF EXISTS notifications CASCADE;