-- name: ReplaceServiceLocations :exec
DELETE FROM service_locations WHERE service_id = $1;

-- name: InsertServiceLocations :exec
INSERT INTO service_locations (service_id, location_id, duration_mins, price, currency, is_active)
SELECT $1, location_id, duration_mins, price, currency, is_active
FROM unnest(
    sqlc.arg('location_ids')::int[],
    sqlc.arg('durations')::int[],
    sqlc.arg('prices')::int[],
    sqlc.arg('currencies')::varchar(3)[],
    sqlc.arg('actives')::bool[]
) AS t(location_id, duration_mins, price, currency, is_active);

-- name: ValidateLocationsBelongToMaster :one
SELECT (
    SELECT COUNT(*) FROM locations
    WHERE id = ANY(sqlc.arg('location_ids')::int[]) AND master_id = $1
) = cardinality(sqlc.arg('location_ids')::int[]) AS all_belong;

-- name: GetLocationsByService :many
SELECT location_id, duration_mins, price, currency, is_active
FROM service_locations WHERE service_id = $1;