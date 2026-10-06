-- name: CreateReview :one
INSERT INTO reviews (author_id, target_id, target, reference_id, rating, comment)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, author_id, target_id, target, reference_id,
          rating, comment, reply, replied_at, status,
          created_at, updated_at;

-- name: GetReviewByID :one
SELECT
    r.id, r.author_id, r.target_id, r.target, r.reference_id,
    r.rating, r.comment, r.reply, r.replied_at, r.status,
    r.created_at, r.updated_at,

    u.name       AS author_name,
    u.avatar_url AS author_avatar
FROM reviews r
JOIN users u ON u.id = r.author_id
WHERE r.id = $1;

-- name: GetReviewByReference :one
SELECT
    r.id, r.author_id, r.target_id, r.target, r.reference_id,
    r.rating, r.comment, r.reply, r.replied_at, r.status,
    r.created_at, r.updated_at,

    u.name       AS author_name,
    u.avatar_url AS author_avatar
FROM reviews r
JOIN users u ON u.id = r.author_id
WHERE r.author_id = $1 AND r.reference_id = $2;

-- name: ListTargetReviews :many
SELECT
    r.id, r.author_id, r.target_id, r.target, r.reference_id,
    r.rating, r.comment, r.reply, r.replied_at, r.status,
    r.created_at, r.updated_at,

    u.name       AS author_name,
    u.avatar_url AS author_avatar
FROM reviews r
JOIN users u ON u.id = r.author_id
WHERE r.target = sqlc.arg('target')
  AND r.target_id = sqlc.arg('target_id')
  AND r.status = 1
  AND (sqlc.narg('cursor_time')::timestamptz IS NULL
       OR (r.created_at, r.id) < (sqlc.narg('cursor_time'), sqlc.narg('cursor_id')))
ORDER BY r.created_at DESC, r.id DESC
LIMIT sqlc.arg('page_size');

-- name: GetReviewSummary :one
SELECT
    sqlc.arg('target')::smallint AS target,
    sqlc.arg('target_id')::int   AS target_id,
    COALESCE(AVG(rating) FILTER (WHERE status = 1), 0)::float AS avg_rating,
    COUNT(*) FILTER (WHERE status = 1)::int                   AS total_reviews,
    COUNT(*) FILTER (WHERE status = 1 AND rating = 1)::int    AS rating_1_count,
    COUNT(*) FILTER (WHERE status = 1 AND rating = 2)::int    AS rating_2_count,
    COUNT(*) FILTER (WHERE status = 1 AND rating = 3)::int    AS rating_3_count,
    COUNT(*) FILTER (WHERE status = 1 AND rating = 4)::int    AS rating_4_count,
    COUNT(*) FILTER (WHERE status = 1 AND rating = 5)::int    AS rating_5_count
FROM reviews
WHERE target = sqlc.arg('target') AND target_id = sqlc.arg('target_id');

-- name: ListMyReviews :many
SELECT
    r.id, r.author_id, r.target_id, r.target, r.reference_id,
    r.rating, r.comment, r.reply, r.replied_at, r.status,
    r.created_at, r.updated_at,

    u.name       AS author_name,
    u.avatar_url AS author_avatar
FROM reviews r
JOIN users u ON u.id = r.author_id
WHERE r.author_id = sqlc.arg('author_id')
  AND r.status <> 3
  AND (sqlc.narg('target_filter')::smallint IS NULL OR r.target = sqlc.narg('target_filter'))
  AND (sqlc.narg('cursor_time')::timestamptz IS NULL
       OR (r.created_at, r.id) < (sqlc.narg('cursor_time'), sqlc.narg('cursor_id')))
ORDER BY r.created_at DESC, r.id DESC
LIMIT sqlc.arg('page_size');

-- name: ListMyTargetReviews :many
SELECT
    r.id, r.author_id, r.target_id, r.target, r.reference_id,
    r.rating, r.comment, r.reply, r.replied_at, r.status,
    r.created_at, r.updated_at,

    u.name       AS author_name,
    u.avatar_url AS author_avatar
FROM reviews r
JOIN users u ON u.id = r.author_id
WHERE r.target = sqlc.arg('target')
  AND r.target_id = sqlc.arg('target_id')
  AND r.status <> 3
  AND (sqlc.narg('only_unanswered')::bool IS NULL
       OR (sqlc.narg('only_unanswered') = TRUE AND r.reply IS NULL))
  AND (sqlc.narg('cursor_time')::timestamptz IS NULL
       OR (r.created_at, r.id) < (sqlc.narg('cursor_time'), sqlc.narg('cursor_id')))
ORDER BY r.created_at DESC, r.id DESC
LIMIT sqlc.arg('page_size');

-- name: UpdateReview :one
UPDATE reviews
SET rating     = sqlc.narg('rating'),
    comment    = sqlc.narg('comment'),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND author_id = sqlc.arg('author_id') AND status = 1
RETURNING id, author_id, target_id, target, reference_id,
          rating, comment, reply, replied_at, status,
          created_at, updated_at;

-- name: ReplyToReview :one
UPDATE reviews
SET reply      = $2,
    replied_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND status = 1
RETURNING id, author_id, target_id, target, reference_id,
          rating, comment, reply, replied_at, status,
          created_at, updated_at;

-- name: DeleteReview :one
UPDATE reviews
SET status     = 3,
    updated_at = NOW()
WHERE id = $1 AND author_id = $2
RETURNING id, status;

-- name: HideReview :one
UPDATE reviews
SET status     = 2,
    updated_at = NOW()
WHERE id = $1 AND status = 1
RETURNING id, author_id, target_id, target, reference_id,
          rating, comment, reply, replied_at, status,
          created_at, updated_at;

-- name: UnhideReview :one
UPDATE reviews
SET status     = 1,
    updated_at = NOW()
WHERE id = $1 AND status = 2
RETURNING id, author_id, target_id, target, reference_id,
          rating, comment, reply, replied_at, status,
          created_at, updated_at;

-- name: ReviewBelongsToAuthor :one
SELECT EXISTS(
    SELECT 1 FROM reviews WHERE id = $1 AND author_id = $2
) AS belongs;

-- name: ReviewExistsForReference :one
SELECT EXISTS(
    SELECT 1 FROM reviews WHERE author_id = $1 AND reference_id = $2
) AS exists;

-- name: GetAverageRatingByTarget :one
SELECT
    COALESCE(AVG(rating) FILTER (WHERE status = 1), 0)::float AS avg_rating,
    COUNT(*) FILTER (WHERE status = 1)::int                   AS total_reviews
FROM reviews
WHERE target = $1 AND target_id = $2;