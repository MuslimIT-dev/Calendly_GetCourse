-- name: ListMastersByRating :many
SELECT
    m.id, m.user_id, m.slug, m.bio, m.specialization, m.years_of_experience,
    m.is_accepting_bookings, m.default_location_id,
    m.min_price, m.avg_rating, m.reviews_count,
    m.created_at, m.updated_at,

    u.name          AS user_name,
    u.email         AS user_email,
    u.avatar_url    AS user_avatar_url,
    u.timezone      AS user_timezone,
    u.email_verified AS user_email_verified,
    (SELECT array_agg(ur.role_id ORDER BY ur.role_id)
     FROM user_roles ur WHERE ur.user_id = u.id)::int[] AS user_role_ids,

    COALESCE((
        SELECT json_agg(json_build_object(
            'id', l.id, 'name', l.name, 'address', l.address,
            'timezone', l.timezone, 'is_online', l.is_online, 'meeting_url', l.meeting_url
        ) ORDER BY l.id)
        FROM locations l WHERE l.master_id = m.id AND l.is_active = TRUE
    ), '[]'::json) AS locations_json,

    COALESCE((
        SELECT json_agg(json_build_object(
            'id', s.id, 'name', s.name, 'slug', s.slug,
            'price', s.price, 'duration_mins', s.duration_mins
        ) ORDER BY s.price ASC)
        FROM (
            SELECT id, name, slug, price, duration_mins
            FROM services
            WHERE master_id = m.id AND is_active = TRUE
            ORDER BY price ASC
            LIMIT 3
        ) s
    ), '[]'::json) AS top_services_json,

    (SELECT COUNT(*) FROM services
     WHERE master_id = m.id AND is_active = TRUE) AS total_services

FROM masters m
JOIN users u ON u.id = m.user_id
WHERE m.is_accepting_bookings = TRUE
  AND (sqlc.narg('min_price')::int      IS NULL OR m.min_price >= sqlc.narg('min_price'))
  AND (sqlc.narg('max_price')::int      IS NULL OR m.min_price <= sqlc.narg('max_price'))
  AND (sqlc.narg('min_experience')::int IS NULL OR m.years_of_experience >= sqlc.narg('min_experience'))
  AND (sqlc.narg('specialization')::text IS NULL OR m.specialization = sqlc.narg('specialization'))
  AND (sqlc.narg('master_timezone')::text IS NULL OR u.timezone = sqlc.narg('master_timezone'))
  AND (sqlc.narg('service_id')::int IS NULL OR EXISTS (
        SELECT 1 FROM services s
        WHERE s.master_id = m.id AND s.id = sqlc.narg('service_id') AND s.is_active = TRUE
  ))
  AND (sqlc.narg('cursor_rating')::numeric IS NULL
       OR (m.avg_rating, m.id) < (sqlc.narg('cursor_rating'), sqlc.narg('cursor_id')))
ORDER BY m.avg_rating DESC, m.id DESC
LIMIT sqlc.arg('page_size');

-- name: ListMastersByExperience :many
SELECT
    m.id, m.user_id, m.slug, m.bio, m.specialization, m.years_of_experience,
    m.is_accepting_bookings, m.default_location_id,
    m.min_price, m.avg_rating, m.reviews_count,
    m.created_at, m.updated_at,

    u.name          AS user_name,
    u.email         AS user_email,
    u.avatar_url    AS user_avatar_url,
    u.timezone      AS user_timezone,
    u.email_verified AS user_email_verified,
    (SELECT array_agg(ur.role_id ORDER BY ur.role_id)
     FROM user_roles ur WHERE ur.user_id = u.id)::int[] AS user_role_ids,

    COALESCE((
        SELECT json_agg(json_build_object(
            'id', l.id, 'name', l.name, 'address', l.address,
            'timezone', l.timezone, 'is_online', l.is_online, 'meeting_url', l.meeting_url
        ) ORDER BY l.id)
        FROM locations l WHERE l.master_id = m.id AND l.is_active = TRUE
    ), '[]'::json) AS locations_json,

    COALESCE((
        SELECT json_agg(json_build_object(
            'id', s.id, 'name', s.name, 'slug', s.slug,
            'price', s.price, 'duration_mins', s.duration_mins
        ) ORDER BY s.price ASC)
        FROM (
            SELECT id, name, slug, price, duration_mins
            FROM services
            WHERE master_id = m.id AND is_active = TRUE
            ORDER BY price ASC
            LIMIT 3
        ) s
    ), '[]'::json) AS top_services_json,

    (SELECT COUNT(*) FROM services
     WHERE master_id = m.id AND is_active = TRUE) AS total_services

FROM masters m
JOIN users u ON u.id = m.user_id
WHERE m.is_accepting_bookings = TRUE
  AND (sqlc.narg('min_price')::int      IS NULL OR m.min_price >= sqlc.narg('min_price'))
  AND (sqlc.narg('max_price')::int      IS NULL OR m.min_price <= sqlc.narg('max_price'))
  AND (sqlc.narg('min_experience')::int IS NULL OR m.years_of_experience >= sqlc.narg('min_experience'))
  AND (sqlc.narg('specialization')::text IS NULL OR m.specialization = sqlc.narg('specialization'))
  AND (sqlc.narg('master_timezone')::text IS NULL OR u.timezone = sqlc.narg('master_timezone'))
  AND (sqlc.narg('service_id')::int IS NULL OR EXISTS (
        SELECT 1 FROM services s
        WHERE s.master_id = m.id AND s.id = sqlc.narg('service_id') AND s.is_active = TRUE
  ))
  AND (sqlc.narg('cursor_experience')::int IS NULL
       OR (m.years_of_experience, m.id) < (sqlc.narg('cursor_experience'), sqlc.narg('cursor_id')))
ORDER BY m.years_of_experience DESC, m.id DESC
LIMIT sqlc.arg('page_size');

-- name: ListMastersByPrice :many
SELECT
    m.id, m.user_id, m.slug, m.bio, m.specialization, m.years_of_experience,
    m.is_accepting_bookings, m.default_location_id,
    m.min_price, m.avg_rating, m.reviews_count,
    m.created_at, m.updated_at,

    u.name          AS user_name,
    u.email         AS user_email,
    u.avatar_url    AS user_avatar_url,
    u.timezone      AS user_timezone,
    u.email_verified AS user_email_verified,
    (SELECT array_agg(ur.role_id ORDER BY ur.role_id)
     FROM user_roles ur WHERE ur.user_id = u.id)::int[] AS user_role_ids,

    COALESCE((
        SELECT json_agg(json_build_object(
            'id', l.id, 'name', l.name, 'address', l.address,
            'timezone', l.timezone, 'is_online', l.is_online, 'meeting_url', l.meeting_url
        ) ORDER BY l.id)
        FROM locations l WHERE l.master_id = m.id AND l.is_active = TRUE
    ), '[]'::json) AS locations_json,

    COALESCE((
        SELECT json_agg(json_build_object(
            'id', s.id, 'name', s.name, 'slug', s.slug,
            'price', s.price, 'duration_mins', s.duration_mins
        ) ORDER BY s.price ASC)
        FROM (
            SELECT id, name, slug, price, duration_mins
            FROM services
            WHERE master_id = m.id AND is_active = TRUE
            ORDER BY price ASC
            LIMIT 3
        ) s
    ), '[]'::json) AS top_services_json,

    (SELECT COUNT(*) FROM services
     WHERE master_id = m.id AND is_active = TRUE) AS total_services

FROM masters m
JOIN users u ON u.id = m.user_id
WHERE m.is_accepting_bookings = TRUE
  AND (sqlc.narg('min_price')::int      IS NULL OR m.min_price >= sqlc.narg('min_price'))
  AND (sqlc.narg('max_price')::int      IS NULL OR m.min_price <= sqlc.narg('max_price'))
  AND (sqlc.narg('min_experience')::int IS NULL OR m.years_of_experience >= sqlc.narg('min_experience'))
  AND (sqlc.narg('specialization')::text IS NULL OR m.specialization = sqlc.narg('specialization'))
  AND (sqlc.narg('master_timezone')::text IS NULL OR u.timezone = sqlc.narg('master_timezone'))
  AND (sqlc.narg('service_id')::int IS NULL OR EXISTS (
        SELECT 1 FROM services s
        WHERE s.master_id = m.id AND s.id = sqlc.narg('service_id') AND s.is_active = TRUE
  ))
  AND (sqlc.narg('cursor_price')::int IS NULL
       OR (m.min_price, m.id) > (sqlc.narg('cursor_price'), sqlc.narg('cursor_id')))
ORDER BY m.min_price ASC, m.id ASC
LIMIT sqlc.arg('page_size');

-- name: GetMasterBySlug :one
SELECT
    m.id, m.user_id, m.slug, m.bio, m.specialization, m.years_of_experience,
    m.is_accepting_bookings, m.default_location_id,
    m.min_price, m.avg_rating, m.reviews_count,
    m.created_at, m.updated_at,

    u.name          AS user_name,
    u.email         AS user_email,
    u.avatar_url    AS user_avatar_url,
    u.timezone      AS user_timezone,
    u.email_verified AS user_email_verified,
    (SELECT array_agg(ur.role_id ORDER BY ur.role_id)
     FROM user_roles ur WHERE ur.user_id = u.id)::int[] AS user_role_ids,

    COALESCE((
        SELECT json_agg(json_build_object(
            'id', s.id, 'name', s.name, 'slug', s.slug, 'description', s.description,
            'price', s.price, 'duration_mins', s.duration_mins, 'is_active', s.is_active
        ) ORDER BY s.price ASC)
        FROM services s WHERE s.master_id = m.id AND s.is_active = TRUE
    ), '[]'::json) AS services_json

FROM masters m
JOIN users u ON u.id = m.user_id
WHERE m.slug = $1;

-- name: GetMasterByUserID :one
SELECT
    m.id, m.user_id, m.slug, m.bio, m.specialization, m.years_of_experience,
    m.is_accepting_bookings, m.default_location_id,
    m.min_price, m.avg_rating, m.reviews_count,
    m.created_at, m.updated_at,

    u.name          AS user_name,
    u.email         AS user_email,
    u.avatar_url    AS user_avatar_url,
    u.timezone      AS user_timezone,
    u.email_verified AS user_email_verified,
    (SELECT array_agg(ur.role_id ORDER BY ur.role_id)
     FROM user_roles ur WHERE ur.user_id = u.id)::int[] AS user_role_ids,

    COALESCE((
        SELECT json_agg(json_build_object(
            'id', s.id, 'name', s.name, 'slug', s.slug, 'description', s.description,
            'price', s.price, 'duration_mins', s.duration_mins, 'is_active', s.is_active
        ) ORDER BY s.price ASC)
        FROM services s WHERE s.master_id = m.id AND s.is_active = TRUE
    ), '[]'::json) AS services_json

FROM masters m
JOIN users u ON u.id = m.user_id
WHERE m.user_id = $1;

-- name: CreateMaster :one
INSERT INTO masters (user_id, slug, bio, specialization, years_of_experience, is_accepting_bookings)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, slug, bio, specialization, years_of_experience,
          is_accepting_bookings, default_location_id, min_price, avg_rating, reviews_count,
          created_at, updated_at;

-- name: UpdateMaster :one
UPDATE masters
SET bio                  = COALESCE(sqlc.narg('bio'), bio),
    specialization       = COALESCE(sqlc.narg('specialization'), specialization),
    years_of_experience  = COALESCE(sqlc.narg('years_of_experience'), years_of_experience),
    is_accepting_bookings = COALESCE(sqlc.narg('is_accepting_bookings'), is_accepting_bookings),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING id, user_id, slug, bio, specialization, years_of_experience,
          is_accepting_bookings, default_location_id, min_price, avg_rating, reviews_count,
          created_at, updated_at;

-- name: UpdateMasterSlug :one
UPDATE masters SET slug = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, slug, bio, specialization, years_of_experience,
          is_accepting_bookings, default_location_id, min_price, avg_rating, reviews_count,
          created_at, updated_at;

-- name: UpdateMasterMinPrice :exec
UPDATE masters SET min_price = (
    SELECT COALESCE(MIN(price), 0) FROM services
    WHERE master_id = $1 AND is_active = TRUE
), updated_at = NOW()
WHERE id = $1;

-- name: UpdateMasterRating :exec
UPDATE masters SET
    avg_rating = (SELECT COALESCE(AVG(r.rating), 0)
                  FROM reviews r
                  WHERE r.target_id = masters.id AND r.target = 1 AND r.status = 1),
    reviews_count = (SELECT COUNT(*)
                     FROM reviews r
                     WHERE r.target_id = masters.id AND r.target = 1 AND r.status = 1),
    updated_at = NOW()
WHERE masters.id = $1;

-- name: SetDefaultLocation :exec
UPDATE masters SET default_location_id = $2, updated_at = NOW() WHERE id = $1;

-- name: GetMasterIDByUserID :one
SELECT id FROM masters WHERE user_id = $1;

-- name: MasterSlugExists :one
SELECT EXISTS(SELECT 1 FROM masters WHERE slug = $1) AS exists;

-- name: ReplaceMasterLanguages :exec
DELETE FROM languages WHERE master_id = $1;

-- name: InsertMasterLanguages :exec
INSERT INTO languages (master_id, language, proficiency)
SELECT $1, x.language, x.proficiency::proficiency_level
FROM jsonb_to_recordset(sqlc.arg('data')::jsonb)
     AS x(language TEXT, proficiency TEXT);

-- name: GetMasterLanguages :many
SELECT language, proficiency FROM languages WHERE master_id = $1;

-- name: ReplaceMasterCertificates :exec
DELETE FROM certificates WHERE master_id = $1;

-- name: InsertMasterCertificates :exec
INSERT INTO certificates (master_id, name, organization, year, file_url)
SELECT
    sqlc.arg('master_id')::int,
    x.name,
    x.organization,
    x.year,
    x.file_url
FROM jsonb_to_recordset(sqlc.arg('data')::jsonb)
     AS x(name TEXT, organization TEXT, year INT, file_url TEXT);

-- name: GetMasterCertificates :many
SELECT id, name, organization, year, file_url FROM certificates WHERE master_id = $1;