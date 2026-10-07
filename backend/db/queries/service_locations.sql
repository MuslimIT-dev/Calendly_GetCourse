-- name: ReplaceServiceLocations :exec
DELETE FROM service_locations WHERE service_id = $1;

-- name: InsertServiceLocations :exec
INSERT INTO service_locations (service_id, location_id, duration_mins, price, currency, is_active)
SELECT
    sqlc.arg('service_id')::int,
    x.location_id,
    x.duration_mins,
    x.price,
    x.currency,
    x.is_active
FROM jsonb_to_recordset(sqlc.arg('data')::jsonb)
     AS x(location_id INT, duration_mins INT, price INT, currency VARCHAR(3), is_active BOOL);

-- name: ValidateLocationsBelongToMaster :one
SELECT (
    SELECT COUNT(*) FROM locations
    WHERE id = ANY(sqlc.arg('location_ids')::int[]) AND master_id = $1
) = cardinality(sqlc.arg('location_ids')::int[]) AS all_belong;

-- name: GetLocationsByService :many
SELECT location_id, duration_mins, price, currency, is_active
FROM service_locations WHERE service_id = $1;