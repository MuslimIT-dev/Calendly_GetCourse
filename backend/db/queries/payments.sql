-- name: CreatePayment :one
INSERT INTO payments (
    user_id, amount, currency, status, provider, purpose,
    reference_id, external_id, payment_url, expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, user_id, amount, currency, status, provider, purpose,
          reference_id, external_id, payment_url,
          failure_reason, receipt_url,
          paid_at, refunded_at, expires_at,
          created_at, updated_at;

-- name: GetPaymentByID :one
SELECT
    p.id, p.user_id, p.amount, p.currency, p.status, p.provider, p.purpose,
    p.reference_id, p.external_id, p.payment_url,
    p.failure_reason, p.receipt_url,
    p.paid_at, p.refunded_at, p.expires_at,
    p.created_at, p.updated_at,

    u.name       AS user_name,
    u.email      AS user_email,
    u.avatar_url AS user_avatar_url
FROM payments p
JOIN users u ON u.id = p.user_id
WHERE p.id = $1;

-- name: GetPaymentByExternalID :one
SELECT
    p.id, p.user_id, p.amount, p.currency, p.status, p.provider, p.purpose,
    p.reference_id, p.external_id, p.payment_url,
    p.failure_reason, p.receipt_url,
    p.paid_at, p.refunded_at, p.expires_at,
    p.created_at, p.updated_at,

    u.name       AS user_name,
    u.email      AS user_email,
    u.avatar_url AS user_avatar_url
FROM payments p
JOIN users u ON u.id = p.user_id
WHERE p.provider = $1 AND p.external_id = $2;

-- name: GetPaymentByReference :one
SELECT
    p.id, p.user_id, p.amount, p.currency, p.status, p.provider, p.purpose,
    p.reference_id, p.external_id, p.payment_url,
    p.failure_reason, p.receipt_url,
    p.paid_at, p.refunded_at, p.expires_at,
    p.created_at, p.updated_at,

    u.name       AS user_name,
    u.email      AS user_email,
    u.avatar_url AS user_avatar_url
FROM payments p
JOIN users u ON u.id = p.user_id
WHERE p.purpose = $1 AND p.reference_id = $2
ORDER BY p.created_at DESC
LIMIT 1;

-- name: ListMyPayments :many
SELECT
    p.id, p.user_id, p.amount, p.currency, p.status, p.provider, p.purpose,
    p.reference_id, p.external_id, p.payment_url,
    p.failure_reason, p.receipt_url,
    p.paid_at, p.refunded_at, p.expires_at,
    p.created_at, p.updated_at,

    u.name       AS user_name,
    u.email      AS user_email,
    u.avatar_url AS user_avatar_url
FROM payments p
JOIN users u ON u.id = p.user_id
WHERE p.user_id = sqlc.arg('user_id')
  AND (sqlc.narg('status_filter')::smallint  IS NULL OR p.status  = sqlc.narg('status_filter'))
  AND (sqlc.narg('purpose_filter')::smallint IS NULL OR p.purpose = sqlc.narg('purpose_filter'))
  AND (sqlc.narg('cursor_time')::timestamptz IS NULL
       OR (p.created_at, p.id) < (sqlc.narg('cursor_time'), sqlc.narg('cursor_id')))
ORDER BY p.created_at DESC, p.id DESC
LIMIT sqlc.arg('page_size');

-- name: ListAllPayments :many
SELECT
    p.id, p.user_id, p.amount, p.currency, p.status, p.provider, p.purpose,
    p.reference_id, p.external_id, p.payment_url,
    p.failure_reason, p.receipt_url,
    p.paid_at, p.refunded_at, p.expires_at,
    p.created_at, p.updated_at,

    u.name       AS user_name,
    u.email      AS user_email,
    u.avatar_url AS user_avatar_url
FROM payments p
JOIN users u ON u.id = p.user_id
WHERE (sqlc.narg('status_filter')::smallint   IS NULL OR p.status = sqlc.narg('status_filter'))
  AND (sqlc.narg('user_id')::int              IS NULL OR p.user_id = sqlc.narg('user_id'))
  AND (sqlc.narg('from_time')::timestamptz    IS NULL OR p.created_at >= sqlc.narg('from_time'))
  AND (sqlc.narg('to_time')::timestamptz      IS NULL OR p.created_at <  sqlc.narg('to_time'))
  AND (sqlc.narg('cursor_time')::timestamptz  IS NULL
       OR (p.created_at, p.id) < (sqlc.narg('cursor_time'), sqlc.narg('cursor_id')))
ORDER BY p.created_at DESC, p.id DESC
LIMIT sqlc.arg('page_size');

-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, amount, currency, status, provider, purpose,
          reference_id, external_id, payment_url,
          failure_reason, receipt_url,
          paid_at, refunded_at, expires_at,
          created_at, updated_at;

-- name: MarkPaymentPaid :one
UPDATE payments
SET status       = 2,
    paid_at      = NOW(),
    external_id  = COALESCE(sqlc.narg('external_id'), external_id),
    receipt_url  = COALESCE(sqlc.narg('receipt_url'), receipt_url),
    payment_url  = COALESCE(sqlc.narg('payment_url'), payment_url),
    updated_at   = NOW()
WHERE id = $1 AND status = 1
RETURNING id, user_id, amount, currency, status, provider, purpose,
          reference_id, external_id, payment_url,
          failure_reason, receipt_url,
          paid_at, refunded_at, expires_at,
          created_at, updated_at;

-- name: MarkPaymentFailed :one
UPDATE payments
SET status         = 3,
    failure_reason = $2,
    updated_at     = NOW()
WHERE id = $1 AND status = 1
RETURNING id, user_id, amount, currency, status, provider, purpose,
          reference_id, external_id, payment_url,
          failure_reason, receipt_url,
          paid_at, refunded_at, expires_at,
          created_at, updated_at;

-- name: MarkPaymentRefunded :one
UPDATE payments
SET status       = 4,
    refunded_at  = NOW(),
    updated_at   = NOW()
WHERE id = $1 AND status = 2
RETURNING id, user_id, amount, currency, status, provider, purpose,
          reference_id, external_id, payment_url,
          failure_reason, receipt_url,
          paid_at, refunded_at, expires_at,
          created_at, updated_at;

-- name: CancelPayment :one
UPDATE payments
SET status     = 3,
    updated_at = NOW()
WHERE id = $1 AND status = 1
RETURNING id, user_id, amount, currency, status, provider, purpose,
          reference_id, external_id, payment_url,
          failure_reason, receipt_url,
          paid_at, refunded_at, expires_at,
          created_at, updated_at;

-- name: ExpirePendingPayments :execrows
UPDATE payments
SET status = 5, updated_at = NOW()
WHERE status = 1 AND expires_at < NOW();

-- name: PaymentBelongsToUser :one
SELECT EXISTS(
    SELECT 1 FROM payments WHERE id = $1 AND user_id = $2
) AS belongs;

-- name: PaymentStatusByID :one
SELECT status FROM payments WHERE id = $1;

-- name: GetPaidPaymentByReference :one
SELECT
    id, user_id, amount, currency, status, provider, purpose,
    reference_id, external_id, payment_url,
    failure_reason, receipt_url,
    paid_at, refunded_at, expires_at,
    created_at, updated_at
FROM payments
WHERE purpose = $1
  AND reference_id = $2
  AND status = 2
ORDER BY paid_at DESC
LIMIT 1;