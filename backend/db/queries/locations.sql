-- name: ListLocationsByMaster :many
SELECT
    id, master_id, name, address, timezone, is_online, meeting_url, is_active,
    created_at, updated_at
FROM locations
WHERE master_id = sqlc.arg('master_id')
  AND (sqlc.narg('only_active')::bool IS NULL OR is_active = sqlc.narg('only_active'))
ORDER BY id;

-- name: GetLocationByID :one
SELECT
    id, master_id, name, address, timezone, is_online, meeting_url, is_active,
    created_at, updated_at
FROM locations
WHERE id = $1;

-- name: CreateLocation :one
INSERT INTO locations (master_id, name, address, timezone, is_online, meeting_url)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, master_id, name, address, timezone, is_online, meeting_url, is_active,
          created_at, updated_at;

-- name: UpdateLocation :one
UPDATE locations
SET name        = COALESCE(sqlc.narg('name'), name),
    address     = COALESCE(sqlc.narg('address'), address),
    timezone    = COALESCE(sqlc.narg('timezone'), timezone),
    is_online   = COALESCE(sqlc.narg('is_online'), is_online),
    meeting_url = COALESCE(sqlc.narg('meeting_url'), meeting_url),
    is_active   = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at  = NOW()
WHERE id = sqlc.arg('id')
RETURNING id, master_id, name, address, timezone, is_online, meeting_url, is_active,
          created_at, updated_at;

-- name: DeleteLocation :exec
DELETE FROM locations WHERE id = $1;

-- name: LocationBelongsToMaster :one
SELECT EXISTS(
    SELECT 1 FROM locations WHERE id = $1 AND master_id = $2
) AS belongs;

-- name: CountActiveLocationsByMaster :one
SELECT COUNT(*) FROM locations
WHERE master_id = $1 AND is_active = TRUE;