-- name: ListCategories :many
SELECT id, name, is_active, created_at, updated_at
FROM categories
WHERE is_active = TRUE
ORDER BY name;

-- name: GetCategory :one
SELECT id, name, is_active, created_at, updated_at
FROM categories WHERE id = $1;

-- name: CreateCategory :one
INSERT INTO categories (name) VALUES ($1)
RETURNING id, name, is_active, created_at, updated_at;

-- name: UpdateCategory :one
UPDATE categories SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, name, is_active, created_at, updated_at;

-- name: DeleteCategory :exec
UPDATE categories SET is_active = FALSE, updated_at = NOW() WHERE id = $1;

-- name: CategoryHasServices :one
SELECT EXISTS(SELECT 1 FROM default_services WHERE category_id = $1 AND is_active = TRUE) AS exists;

-- name: ListDefaultServices :many
SELECT id, category_id, slug, name, description, price, currency, duration_mins, is_active,
       created_at, updated_at
FROM default_services
WHERE is_active = TRUE
  AND (sqlc.narg('category_id')::int IS NULL OR category_id = sqlc.narg('category_id'))
ORDER BY name;

-- name: GetDefaultService :one
SELECT id, category_id, slug, name, description, price, currency, duration_mins, is_active,
       created_at, updated_at
FROM default_services WHERE id = $1;

-- name: CreateDefaultService :one
INSERT INTO default_services (category_id, slug, name, description, price, currency, duration_mins)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, category_id, slug, name, description, price, currency, duration_mins, is_active,
          created_at, updated_at;

-- name: UpdateDefaultService :one
UPDATE default_services
SET slug          = COALESCE(sqlc.narg('slug'), slug),
    name          = COALESCE(sqlc.narg('name'), name),
    description   = COALESCE(sqlc.narg('description'), description),
    duration_mins = COALESCE(sqlc.narg('duration_mins'), duration_mins),
    price         = COALESCE(sqlc.narg('price'), price),
    updated_at    = NOW()
WHERE id = sqlc.arg('id')
RETURNING id, category_id, slug, name, description, price, currency, duration_mins, is_active,
          created_at, updated_at;

-- name: DeleteDefaultService :exec
UPDATE default_services SET is_active = FALSE, updated_at = NOW() WHERE id = $1;

-- name: DefaultServiceSlugExists :one
SELECT EXISTS(SELECT 1 FROM default_services WHERE slug = $1 AND id <> COALESCE(sqlc.narg('exclude_id'), 0)) AS exists;