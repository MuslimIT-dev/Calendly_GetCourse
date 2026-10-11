package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/repository/db"
)

type ScheduleRepo struct {
	q *db.Queries
}

func NewScheduleRepo(q *db.Queries) *ScheduleRepo {
	return &ScheduleRepo{q: q}
}

var _ domain.ScheduleRepository = (*ScheduleRepo)(nil)

func (r *ScheduleRepo) GetRulesByLocation(ctx context.Context, locationID int32) ([]*domain.ScheduleRule, error) {
	rows, err := r.q.GetScheduleRulesByLocation(ctx, locationID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.ScheduleRule, len(rows))
	for i, row := range rows {
		out[i] = mapRule(row)
	}
	return out, nil
}

func (r *ScheduleRepo) UpsertRule(ctx context.Context, rule *domain.ScheduleRule) (*domain.ScheduleRule, error) {
	row, err := r.q.UpsertScheduleRule(ctx, db.UpsertScheduleRuleParams{
		LocationID:             rule.LocationID,
		DayOfWeek:              int16(rule.DayOfWeek),
		IsWorkingDay:           rule.IsWorkingDay,
		WorkStart:              timeOfDayToPgTime(rule.WorkStart),
		WorkEnd:                timeOfDayToPgTime(rule.WorkEnd),
		RestStart:              timeOfDayToPgTime(rule.RestStart),
		RestEnd:                timeOfDayToPgTime(rule.RestEnd),
		BufferBetweenSlotsMins: rule.BufferBetweenSlotsMin,
		SlotStepMins:           rule.SlotStepMins,
	})
	if err != nil {
		return nil, err
	}
	return mapRuleFromUpsert(row), nil
}

func (r *ScheduleRepo) DeleteRulesByLocation(ctx context.Context, locationID int32) error {
	return r.q.DeleteScheduleRulesByLocation(ctx, locationID)
}

func (r *ScheduleRepo) ListExceptions(ctx context.Context, locationID int32, from, to time.Time) ([]*domain.Exception, error) {
	rows, err := r.q.GetExceptionsByLocationAndRange(ctx, db.GetExceptionsByLocationAndRangeParams{
		LocationID: locationID,
		FromDate:   pgtype.Date{Time: from, Valid: true},
		ToDate:     pgtype.Date{Time: to, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Exception, len(rows))
	for i, row := range rows {
		out[i] = mapException(row)
	}
	return out, nil
}

func (r *ScheduleRepo) GetExceptionByDate(ctx context.Context, locationID int32, date time.Time) (*domain.Exception, error) {
	row, err := r.q.GetExceptionByLocationAndDate(ctx, db.GetExceptionByLocationAndDateParams{
		LocationID: locationID,
		Date:       pgtype.Date{Time: date, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return mapException(row), nil
}

func (r *ScheduleRepo) CreateException(ctx context.Context, e *domain.Exception) (*domain.Exception, error) {
	row, err := r.q.CreateException(ctx, db.CreateExceptionParams{
		LocationID:  e.LocationID,
		Date:        pgtype.Date{Time: e.Date, Valid: true},
		Type:        int16(e.Type),
		IsFullDay:   e.IsFullDay,
		CustomStart: timeOfDayToPgTime(e.CustomStart),
		CustomEnd:   timeOfDayToPgTime(e.CustomEnd),
		RestStart:   timeOfDayToPgTime(e.RestStart),
		RestEnd:     timeOfDayToPgTime(e.RestEnd),
		Reason:      pgtype.Text{String: e.Reason, Valid: e.Reason != ""},
	})
	if err != nil {
		return nil, err
	}
	return mapExceptionFromCreate(row), nil
}

func (r *ScheduleRepo) DeleteException(ctx context.Context, id int32) error {
	return r.q.DeleteException(ctx, id)
}

func (r *ScheduleRepo) DeleteExceptionByDate(ctx context.Context, locationID int32, date time.Time) error {
	return r.q.DeleteExceptionByLocationAndDate(ctx, db.DeleteExceptionByLocationAndDateParams{
		LocationID: locationID,
		Date:       pgtype.Date{Time: date, Valid: true},
	})
}

func (r *ScheduleRepo) GetBookedSlots(ctx context.Context, masterID, locationID int32, from, to time.Time) ([]domain.TimeSlot, error) {
	rows, err := r.q.GetBookedSlotsInRange(ctx, db.GetBookedSlotsInRangeParams{
		MasterID:   masterID,
		LocationID: locationID,
		FromTime:   pgtype.Timestamptz{Time: from, Valid: true},
		ToTime:     pgtype.Timestamptz{Time: to, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.TimeSlot, len(rows))
	for i, row := range rows {
		out[i] = domain.TimeSlot{Start: row.StartTime.Time, End: row.EndTime.Time}
	}
	return out, nil
}

func timeOfDayToPgTime(t *domain.TimeOfDay) pgtype.Time {
	if t == nil {
		return pgtype.Time{}
	}
	micros := int64(t.Hours)*3600*1_000_000 + int64(t.Minutes)*60*1_000_000 + int64(t.Seconds)*1_000_000
	return pgtype.Time{Microseconds: micros, Valid: true}
}

func pgTimeToTimeOfDay(t pgtype.Time) *domain.TimeOfDay {
	if !t.Valid {
		return nil
	}
	micros := t.Microseconds
	hours := micros / (3600 * 1_000_000)
	micros %= 3600 * 1_000_000
	minutes := micros / (60 * 1_000_000)
	micros %= 60 * 1_000_000
	seconds := micros / 1_000_000
	return &domain.TimeOfDay{Hours: int32(hours), Minutes: int32(minutes), Seconds: int32(seconds)}
}

func mapRule(row db.ScheduleRule) *domain.ScheduleRule {
	return &domain.ScheduleRule{
		ID:                    row.ID,
		LocationID:            row.LocationID,
		DayOfWeek:             int32(row.DayOfWeek),
		IsWorkingDay:          row.IsWorkingDay,
		WorkStart:             pgTimeToTimeOfDay(row.WorkStart),
		WorkEnd:               pgTimeToTimeOfDay(row.WorkEnd),
		RestStart:             pgTimeToTimeOfDay(row.RestStart),
		RestEnd:               pgTimeToTimeOfDay(row.RestEnd),
		BufferBetweenSlotsMin: row.BufferBetweenSlotsMins,
		SlotStepMins:          row.SlotStepMins,
		CreatedAt:             row.CreatedAt.Time,
		UpdatedAt:             row.UpdatedAt.Time,
	}
}

func mapRuleFromUpsert(row db.ScheduleRule) *domain.ScheduleRule {
	return &domain.ScheduleRule{
		ID:                    row.ID,
		LocationID:            row.LocationID,
		DayOfWeek:             int32(row.DayOfWeek),
		IsWorkingDay:          row.IsWorkingDay,
		WorkStart:             pgTimeToTimeOfDay(row.WorkStart),
		WorkEnd:               pgTimeToTimeOfDay(row.WorkEnd),
		RestStart:             pgTimeToTimeOfDay(row.RestStart),
		RestEnd:               pgTimeToTimeOfDay(row.RestEnd),
		BufferBetweenSlotsMin: row.BufferBetweenSlotsMins,
		SlotStepMins:          row.SlotStepMins,
		CreatedAt:             row.CreatedAt.Time,
		UpdatedAt:             row.UpdatedAt.Time,
	}
}

func mapException(row db.Exception) *domain.Exception {
	return &domain.Exception{
		ID:          row.ID,
		LocationID:  row.LocationID,
		Date:        row.Date.Time,
		Type:        domain.ExceptionType(row.Type),
		IsFullDay:   row.IsFullDay,
		CustomStart: pgTimeToTimeOfDay(row.CustomStart),
		CustomEnd:   pgTimeToTimeOfDay(row.CustomEnd),
		RestStart:   pgTimeToTimeOfDay(row.RestStart),
		RestEnd:     pgTimeToTimeOfDay(row.RestEnd),
		Reason:      row.Reason.String,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

func mapExceptionFromCreate(row db.Exception) *domain.Exception {
	return &domain.Exception{
		ID:          row.ID,
		LocationID:  row.LocationID,
		Date:        row.Date.Time,
		Type:        domain.ExceptionType(row.Type),
		IsFullDay:   row.IsFullDay,
		CustomStart: pgTimeToTimeOfDay(row.CustomStart),
		CustomEnd:   pgTimeToTimeOfDay(row.CustomEnd),
		RestStart:   pgTimeToTimeOfDay(row.RestStart),
		RestEnd:     pgTimeToTimeOfDay(row.RestEnd),
		Reason:      row.Reason.String,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
