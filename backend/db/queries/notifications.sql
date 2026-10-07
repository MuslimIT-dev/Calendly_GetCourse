-- name: CreateNotification :one
INSERT INTO notifications (
    user_id, type, channel, status, title, body, link, metadata, sent_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, type, channel, status, title, body, link,
          metadata, sent_at, read_at, created_at;

-- name: GetNotificationByID :one
SELECT
    id, user_id, type, channel, status, title, body, link,
    metadata, sent_at, read_at, created_at
FROM notifications
WHERE id = $1;

-- name: ListMyNotifications :many
SELECT
    n.id, n.user_id, n.type, n.channel, n.status, n.title, n.body, n.link,
    n.metadata, n.sent_at, n.read_at, n.created_at,
    (SELECT COUNT(*)::int FROM notifications n2
     WHERE n2.user_id = sqlc.arg('user_id') AND n2.status != 4) AS unread_count
FROM notifications n
WHERE n.user_id = sqlc.arg('user_id')
  AND (sqlc.narg('status_filter')::smallint IS NULL OR n.status = sqlc.narg('status_filter'))
  AND (sqlc.narg('cursor_time')::timestamptz IS NULL
       OR (n.created_at, n.id) < (sqlc.narg('cursor_time'), sqlc.narg('cursor_id')))
ORDER BY n.created_at DESC, n.id DESC
LIMIT sqlc.arg('page_size');

-- name: MarkNotificationAsRead :execrows
UPDATE notifications
SET status  = 4,
    read_at = NOW()
WHERE id = $1 AND user_id = $2 AND status != 4;

-- name: MarkAllNotificationsAsRead :execrows
UPDATE notifications
SET status  = 4,
    read_at = NOW()
WHERE user_id = $1 AND status != 4;

-- name: DeleteNotification :execrows
DELETE FROM notifications
WHERE id = $1 AND user_id = $2;

-- name: GetUnreadCount :one
SELECT COUNT(*)::int AS count
FROM notifications
WHERE user_id = $1 AND status != 4;

-- name: GetPendingNotifications :many
SELECT
    id, user_id, type, channel, status, title, body, link,
    metadata, sent_at, read_at, created_at
FROM notifications
WHERE status = 1
ORDER BY created_at ASC
LIMIT $1;

-- name: MarkNotificationSent :exec
UPDATE notifications
SET status  = 2,
    sent_at = NOW()
WHERE id = $1 AND status = 1;

-- name: MarkNotificationFailed :exec
UPDATE notifications
SET status = 3
WHERE id = $1 AND status = 1;

-- name: ListUserNotificationsByType :many
SELECT
    id, user_id, type, channel, status, title, body, link,
    metadata, sent_at, read_at, created_at
FROM notifications
WHERE user_id = $1 AND type = $2
ORDER BY created_at DESC
LIMIT $3;