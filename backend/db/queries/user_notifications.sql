-- name: GetNotificationSettings :one
SELECT
    user_id, email_enabled, telegram_enabled, sms_enabled, push_enabled,
    muted_types, reminder_24h, reminder_2h, reminder_30m, updated_at
FROM notification_settings
WHERE user_id = $1;

-- name: CreateDefaultNotificationSettings :one
INSERT INTO notification_settings (user_id)
VALUES ($1)
ON CONFLICT (user_id) DO NOTHING
RETURNING
    user_id, email_enabled, telegram_enabled, sms_enabled, push_enabled,
    muted_types, reminder_24h, reminder_2h, reminder_30m, updated_at;

-- name: UpdateNotificationSettings :one
INSERT INTO notification_settings (
    user_id, email_enabled, telegram_enabled, sms_enabled, push_enabled,
    muted_types, reminder_24h, reminder_2h, reminder_30m
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (user_id) DO UPDATE SET
    email_enabled    = EXCLUDED.email_enabled,
    telegram_enabled = EXCLUDED.telegram_enabled,
    sms_enabled      = EXCLUDED.sms_enabled,
    push_enabled     = EXCLUDED.push_enabled,
    muted_types      = EXCLUDED.muted_types,
    reminder_24h     = EXCLUDED.reminder_24h,
    reminder_2h      = EXCLUDED.reminder_2h,
    reminder_30m     = EXCLUDED.reminder_30m,
    updated_at       = NOW()
RETURNING
    user_id, email_enabled, telegram_enabled, sms_enabled, push_enabled,
    muted_types, reminder_24h, reminder_2h, reminder_30m, updated_at;

-- name: IsTypeMuted :one
SELECT EXISTS(
    SELECT 1 FROM notification_settings
    WHERE user_id = $1 AND $2 = ANY(muted_types)
) AS muted;

-- name: GetChannelEnabled :one
SELECT
    email_enabled,
    telegram_enabled,
    sms_enabled,
    push_enabled
FROM notification_settings
WHERE user_id = $1;