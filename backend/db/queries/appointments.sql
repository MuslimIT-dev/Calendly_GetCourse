-- name: CreateAppointment :one
INSERT INTO appointments (
    client_id, created_by, master_id, service_id, location_id,
    start_time, end_time, timezone,
    price, currency,
    is_online, meeting_url,
    client_notes, status, payment_status
)
VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8,
    $9, $10,
    $11, $12,
    $13, $14, $15
)
RETURNING id, client_id, created_by, master_id, service_id, location_id,
          start_time, end_time, timezone, status, price, currency, payment_status,
          is_online, meeting_url, client_notes, master_notes,
          cancelled_at, cancelled_by, cancel_reason,
          created_at, updated_at;

-- name: GetAppointmentByID :one
SELECT
    a.id, a.client_id, a.created_by, a.master_id, a.service_id, a.location_id,
    a.start_time, a.end_time, a.timezone, a.status, a.price, a.currency, a.payment_status,
    a.is_online, a.meeting_url, a.client_notes, a.master_notes,
    a.cancelled_at, a.cancelled_by, a.cancel_reason,
    a.created_at, a.updated_at,

    m.slug           AS master_slug,
    mu.name          AS master_name,
    s.name           AS service_name,
    s.slug           AS service_slug,
    l.name           AS location_name,
    l.address        AS location_address
FROM appointments a
JOIN masters   m  ON m.id = a.master_id
JOIN users     mu ON mu.id = m.user_id
JOIN services  s  ON s.id = a.service_id
JOIN locations l  ON l.id = a.location_id
WHERE a.id = $1;

-- name: ListMyAppointmentsByStatus :many
SELECT
    a.id, a.client_id, a.created_by, a.master_id, a.service_id, a.location_id,
    a.start_time, a.end_time, a.timezone, a.status, a.price, a.currency, a.payment_status,
    a.is_online, a.meeting_url, a.client_notes, a.master_notes,
    a.cancelled_at, a.cancelled_by, a.cancel_reason,
    a.created_at, a.updated_at,

    m.slug           AS master_slug,
    mu.name          AS master_name,
    s.name           AS service_name,
    s.slug           AS service_slug,
    l.name           AS location_name,
    l.address        AS location_address
FROM appointments a
JOIN masters   m  ON m.id = a.master_id
JOIN users     mu ON mu.id = m.user_id
JOIN services  s  ON s.id = a.service_id
JOIN locations l  ON l.id = a.location_id
WHERE a.client_id = sqlc.arg('client_id')
  AND (sqlc.narg('status_filter')::smallint IS NULL OR a.status = sqlc.narg('status_filter'))
  AND (sqlc.narg('cursor_time')::timestamptz IS NULL
       OR (a.start_time, a.id) < (sqlc.narg('cursor_time'), sqlc.narg('cursor_id')))
ORDER BY a.start_time DESC, a.id DESC
LIMIT sqlc.arg('page_size');

-- name: ListMyAppointmentsUpcoming :many
SELECT
    a.id, a.client_id, a.created_by, a.master_id, a.service_id, a.location_id,
    a.start_time, a.end_time, a.timezone, a.status, a.price, a.currency, a.payment_status,
    a.is_online, a.meeting_url, a.client_notes, a.master_notes,
    a.cancelled_at, a.cancelled_by, a.cancel_reason,
    a.created_at, a.updated_at,

    m.slug           AS master_slug,
    mu.name          AS master_name,
    s.name           AS service_name,
    s.slug           AS service_slug,
    l.name           AS location_name,
    l.address        AS location_address
FROM appointments a
JOIN masters   m  ON m.id = a.master_id
JOIN users     mu ON mu.id = m.user_id
JOIN services  s  ON s.id = a.service_id
JOIN locations l  ON l.id = a.location_id
WHERE a.client_id = sqlc.arg('client_id')
  AND a.status IN (1, 2)
  AND a.start_time >= NOW()
ORDER BY a.start_time ASC
LIMIT sqlc.arg('page_size');

-- name: ListMasterAppointmentsByRange :many
SELECT
    a.id, a.client_id, a.created_by, a.master_id, a.service_id, a.location_id,
    a.start_time, a.end_time, a.timezone, a.status, a.price, a.currency, a.payment_status,
    a.is_online, a.meeting_url, a.client_notes, a.master_notes,
    a.cancelled_at, a.cancelled_by, a.cancel_reason,
    a.created_at, a.updated_at,

    m.slug           AS master_slug,
    mu.name          AS master_name,
    s.name           AS service_name,
    s.slug           AS service_slug,
    l.name           AS location_name,
    l.address        AS location_address
FROM appointments a
JOIN masters   m  ON m.id = a.master_id
JOIN users     mu ON mu.id = m.user_id
JOIN services  s  ON s.id = a.service_id
JOIN locations l  ON l.id = a.location_id
WHERE a.master_id = sqlc.arg('master_id')
  AND (sqlc.narg('location_id')::int IS NULL OR a.location_id = sqlc.narg('location_id'))
  AND (sqlc.narg('status_filter')::smallint IS NULL OR a.status = sqlc.narg('status_filter'))
  AND (sqlc.narg('from_time')::timestamptz IS NULL OR a.start_time >= sqlc.narg('from_time'))
  AND (sqlc.narg('to_time')::timestamptz   IS NULL OR a.start_time <  sqlc.narg('to_time'))
  AND (sqlc.narg('cursor_time')::timestamptz IS NULL
       OR (a.start_time, a.id) < (sqlc.narg('cursor_time'), sqlc.narg('cursor_id')))
ORDER BY a.start_time DESC, a.id DESC
LIMIT sqlc.arg('page_size');

-- name: UpdateAppointmentStatus :one
UPDATE appointments
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, status, updated_at;

-- name: CancelAppointment :one
UPDATE appointments
SET status        = $2,
    cancelled_at  = NOW(),
    cancelled_by  = $3,
    cancel_reason = $4,
    updated_at    = NOW()
WHERE id = $1
RETURNING id, client_id, master_id, service_id, location_id,
          start_time, end_time, timezone, status, price, currency, payment_status,
          is_online, meeting_url, client_notes, master_notes,
          cancelled_at, cancelled_by, cancel_reason,
          created_at, updated_at;

-- name: ConfirmAppointment :one
UPDATE appointments
SET status = 2, updated_at = NOW()
WHERE id = $1 AND status = 1
RETURNING id, client_id, master_id, service_id, location_id,
          start_time, end_time, timezone, status, price, currency, payment_status,
          is_online, meeting_url, client_notes, master_notes,
          cancelled_at, cancelled_by, cancel_reason,
          created_at, updated_at;

-- name: CompleteAppointment :one
UPDATE appointments
SET status = 5, master_notes = $2, updated_at = NOW()
WHERE id = $1 AND status IN (1, 2)
RETURNING id, client_id, master_id, service_id, location_id,
          start_time, end_time, timezone, status, price, currency, payment_status,
          is_online, meeting_url, client_notes, master_notes,
          cancelled_at, cancelled_by, cancel_reason,
          created_at, updated_at;

-- name: MarkAppointmentNoShow :one
UPDATE appointments
SET status = 6, updated_at = NOW()
WHERE id = $1 AND status IN (1, 2)
RETURNING id, client_id, master_id, service_id, location_id,
          start_time, end_time, timezone, status, price, currency, payment_status,
          is_online, meeting_url, client_notes, master_notes,
          cancelled_at, cancelled_by, cancel_reason,
          created_at, updated_at;

-- name: RescheduleAppointment :one
UPDATE appointments
SET start_time  = $2,
    end_time    = $3,
    location_id = $4,
    updated_at  = NOW()
WHERE id = $1 AND status IN (1, 2)
RETURNING id, client_id, master_id, service_id, location_id,
          start_time, end_time, timezone, status, price, currency, payment_status,
          is_online, meeting_url, client_notes, master_notes,
          cancelled_at, cancelled_by, cancel_reason,
          created_at, updated_at;

-- name: UpdateAppointmentPaymentStatus :exec
UPDATE appointments
SET payment_status = $2, updated_at = NOW()
WHERE id = $1;

-- name: AppointmentBelongsToClient :one
SELECT EXISTS(
    SELECT 1 FROM appointments WHERE id = $1 AND client_id = $2
) AS belongs;

-- name: AppointmentBelongsToMaster :one
SELECT EXISTS(
    SELECT 1 FROM appointments WHERE id = $1 AND master_id = $2
) AS belongs;