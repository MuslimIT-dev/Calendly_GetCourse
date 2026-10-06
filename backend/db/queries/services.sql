-- name: GetMyServices :many
SELECT
    s.id, s.master_id, s.default_service_id, s.slug, s.name, s.description,
    s.price, s.currency, s.duration_mins, s.is_active,
    s.created_at, s.updated_at,
    COALESCE((
        SELECT json_agg(json_build_object(
            'location_id', sl.location_id,
            'duration_mins', COALESCE(NULLIF(sl.duration_mins, 0), s.duration_mins),
            'price', COALESCE(NULLIF(sl.price, 0), s.price),
            'is_active', sl.is_active
        ) ORDER BY sl.location_id)
        FROM service_locations sl
        WHERE sl.service_id = s.id
    ), '[]'::json) AS locations_json
FROM services s
WHERE s.master_id = $1 AND s.is_active = TRUE
ORDER BY s.id;

-- name: GetServiceByID :one
SELECT
    s.id, s.master_id, s.default_service_id, s.slug, s.name, s.description,
    s.price, s.currency, s.duration_mins, s.is_active,
    s.created_at, s.updated_at,
    COALESCE((
        SELECT json_agg(json_build_object(
            'location_id', sl.location_id,
            'duration_mins', COALESCE(NULLIF(sl.duration_mins, 0), s.duration_mins),
            'price', COALESCE(NULLIF(sl.price, 0), s.price),
            'is_active', sl.is_active
        ) ORDER BY sl.location_id)
        FROM service_locations sl
        WHERE sl.service_id = s.id
    ), '[]'::json) AS locations_json
FROM services s
WHERE s.id = $1;

-- name: CreateService :one
INSERT INTO services (master_id, default_service_id, slug, name, description, price, currency, duration_mins)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, master_id, default_service_id, slug, name, description,
          price, currency, duration_mins, is_active, created_at, updated_at;

-- name: UpdateService :one
UPDATE services
SET description   = COALESCE(sqlc.narg('description'), description),
    duration_mins = COALESCE(sqlc.narg('duration_mins'), duration_mins),
    price         = COALESCE(sqlc.narg('price'), price),
    updated_at    = NOW()
WHERE id = sqlc.arg('id') AND master_id = sqlc.arg('master_id')
RETURNING id, master_id, default_service_id, slug, name, description,
          price, currency, duration_mins, is_active, created_at, updated_at;

-- name: DeleteService :exec
UPDATE services SET is_active = FALSE, updated_at = NOW()
WHERE id = $1 AND master_id = $2;

-- name: ServiceSlugExists :one
SELECT EXISTS(
    SELECT 1 FROM services
    WHERE master_id = $1 AND slug = $2 AND id <> COALESCE(sqlc.narg('exclude_id'), 0)
) AS exists;

-- name: GetDefaultServiceForCopy :one
SELECT id, category_id, slug, name, description, price, currency, duration_mins
FROM default_services WHERE id = $1 AND is_active = TRUE;

-- name: UpdateServicesMinPrice :exec
UPDATE masters SET min_price = (
    SELECT COALESCE(MIN(price), 0) FROM services
    WHERE master_id = $1 AND is_active = TRUE
), updated_at = NOW()
WHERE id = $1;