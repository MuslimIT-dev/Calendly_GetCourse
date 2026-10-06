-- name: GetExceptionsByLocationAndRange :many
SELECT
    id, location_id, date, type, is_full_day,
    custom_start, custom_end, rest_start, rest_end,
    reason, created_at, updated_at
FROM exceptions
WHERE location_id = $1
  AND date BETWEEN sqlc.arg('from_date')::date AND sqlc.arg('to_date')::date
ORDER BY date;

-- name: GetExceptionByLocationAndDate :one
SELECT
    id, location_id, date, type, is_full_day,
    custom_start, custom_end, rest_start, rest_end,
    reason, created_at, updated_at
FROM exceptions
WHERE location_id = $1 AND date = $2;

-- name: CreateException :one
INSERT INTO exceptions (
    location_id, date, type, is_full_day,
    custom_start, custom_end, rest_start, rest_end, reason
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, location_id, date, type, is_full_day,
          custom_start, custom_end, rest_start, rest_end,
          reason, created_at, updated_at;

-- name: DeleteException :exec
DELETE FROM exceptions WHERE id = $1;

-- name: DeleteExceptionByLocationAndDate :exec
DELETE FROM exceptions WHERE location_id = $1 AND date = $2;

-- name: ExceptionBelongsToLocation :one
SELECT EXISTS(
    SELECT 1 FROM exceptions WHERE id = $1 AND location_id = $2
) AS belongs;