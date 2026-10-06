-- name: GetScheduleRulesByLocation :many
SELECT
    id, location_id, day_of_week, is_working_day,
    work_start, work_end, rest_start, rest_end,
    buffer_between_slots_mins, slot_step_mins,
    created_at, updated_at
FROM schedule_rules
WHERE location_id = $1
ORDER BY day_of_week;

-- name: GetScheduleRuleByLocationAndDay :one
SELECT
    id, location_id, day_of_week, is_working_day,
    work_start, work_end, rest_start, rest_end,
    buffer_between_slots_mins, slot_step_mins,
    created_at, updated_at
FROM schedule_rules
WHERE location_id = $1 AND day_of_week = $2;

-- name: DeleteScheduleRulesByLocation :exec
DELETE FROM schedule_rules WHERE location_id = $1;

-- name: InsertScheduleRule :one
INSERT INTO schedule_rules (
    location_id, day_of_week, is_working_day,
    work_start, work_end, rest_start, rest_end,
    buffer_between_slots_mins, slot_step_mins
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, location_id, day_of_week, is_working_day,
          work_start, work_end, rest_start, rest_end,
          buffer_between_slots_mins, slot_step_mins,
          created_at, updated_at;

-- name: UpsertScheduleRule :one
INSERT INTO schedule_rules (
    location_id, day_of_week, is_working_day,
    work_start, work_end, rest_start, rest_end,
    buffer_between_slots_mins, slot_step_mins
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (location_id, day_of_week) DO UPDATE SET
    is_working_day            = EXCLUDED.is_working_day,
    work_start                = EXCLUDED.work_start,
    work_end                  = EXCLUDED.work_end,
    rest_start                = EXCLUDED.rest_start,
    rest_end                  = EXCLUDED.rest_end,
    buffer_between_slots_mins = EXCLUDED.buffer_between_slots_mins,
    slot_step_mins            = EXCLUDED.slot_step_mins,
    updated_at                = NOW()
RETURNING id, location_id, day_of_week, is_working_day,
          work_start, work_end, rest_start, rest_end,
          buffer_between_slots_mins, slot_step_mins,
          created_at, updated_at;

-- name: GetBookedSlotsInRange :many
SELECT
    a.start_time, a.end_time
FROM appointments a
WHERE a.master_id = sqlc.arg('master_id')
  AND a.location_id = sqlc.arg('location_id')
  AND a.status IN (1, 2)
  AND a.start_time >= sqlc.arg('from_time')
  AND a.start_time <  sqlc.arg('to_time')
ORDER BY a.start_time;

-- name: GetBookedSlotsForDay :many
SELECT
    a.start_time, a.end_time
FROM appointments a
WHERE a.master_id = sqlc.arg('master_id')
  AND a.location_id = sqlc.arg('location_id')
  AND a.status IN (1, 2)
  AND a.start_time >= sqlc.arg('day_start')
  AND a.start_time <  sqlc.arg('day_end')
ORDER BY a.start_time;