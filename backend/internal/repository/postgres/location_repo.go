package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/repository/postgres/db"
)

type LocationRepo struct {
	q *db.Queries
}

func NewLocationRepo(q *db.Queries) *LocationRepo {
	return &LocationRepo{q: q}
}

var _ domain.LocationRepository = (*LocationRepo)(nil)

func (r *LocationRepo) ListByMaster(ctx context.Context, masterID int32, onlyActive bool) ([]*domain.Location, error) {
	var active pgtype.Bool
	if onlyActive {
		active = pgtype.Bool{Bool: true, Valid: true}
	}
	rows, err := r.q.ListLocationsByMaster(ctx, db.ListLocationsByMasterParams{
		MasterID:   masterID,
		OnlyActive: active,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Location, len(rows))
	for i, row := range rows {
		out[i] = mapLocation(row)
	}
	return out, nil
}

func (r *LocationRepo) GetByID(ctx context.Context, id int32) (*domain.Location, error) {
	row, err := r.q.GetLocationByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLocationNotFound
		}
		return nil, err
	}
	return mapLocation(row), nil
}

func (r *LocationRepo) Create(ctx context.Context, l *domain.Location) (*domain.Location, error) {
	row, err := r.q.CreateLocation(ctx, db.CreateLocationParams{
		MasterID:   l.MasterID,
		Name:       l.Name,
		Address:    l.Address,
		Timezone:   l.Timezone,
		IsOnline:   l.IsOnline,
		MeetingUrl: pgtype.Text{String: l.MeetingURL, Valid: l.MeetingURL != ""},
	})
	if err != nil {
		return nil, err
	}
	return mapLocation(row), nil
}

func (r *LocationRepo) Update(ctx context.Context, l *domain.Location) (*domain.Location, error) {
	row, err := r.q.UpdateLocation(ctx, db.UpdateLocationParams{
		ID:         l.ID,
		Name:       pgtype.Text{String: l.Name, Valid: true},
		Address:    pgtype.Text{String: l.Address, Valid: true},
		Timezone:   pgtype.Text{String: l.Timezone, Valid: true},
		IsOnline:   pgtype.Bool{Bool: l.IsOnline, Valid: true},
		MeetingUrl: pgtype.Text{String: l.MeetingURL, Valid: true},
		IsActive:   pgtype.Bool{Bool: l.IsActive, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return mapLocation(row), nil
}

func (r *LocationRepo) Delete(ctx context.Context, id int32) error {
	return r.q.DeleteLocation(ctx, id)
}

func (r *LocationRepo) BelongsToMaster(ctx context.Context, id, masterID int32) (bool, error) {
	return r.q.LocationBelongsToMaster(ctx, db.LocationBelongsToMasterParams{
		ID:       id,
		MasterID: masterID,
	})
}

func (r *LocationRepo) ExistsActiveByMaster(ctx context.Context, masterID int32) (bool, error) {
	n, err := r.q.CountActiveLocationsByMaster(ctx, masterID)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func mapLocation(row db.Location) *domain.Location {
	return &domain.Location{
		ID:         row.ID,
		MasterID:   row.MasterID,
		Name:       row.Name,
		Address:    row.Address,
		Timezone:   row.Timezone,
		IsOnline:   row.IsOnline.Bool,
		MeetingURL: row.MeetingUrl.String,
		IsActive:   row.IsActive.Bool,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}
}