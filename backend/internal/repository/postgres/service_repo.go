package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/repository/db"
)

type ServiceRepo struct {
	q *db.Queries
}

func NewServiceRepo(q *db.Queries) *ServiceRepo {
	return &ServiceRepo{q: q}
}

var _ domain.ServiceRepository = (*ServiceRepo)(nil)

func (r *ServiceRepo) GetByID(ctx context.Context, id int32) (*domain.Service, error) {
	row, err := r.q.GetServiceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrServiceNotFound
		}
		return nil, err
	}
	return mapServiceFull(row)
}

func (r *ServiceRepo) ListByMaster(ctx context.Context, masterID int32) ([]*domain.Service, error) {
	rows, err := r.q.GetMyServices(ctx, masterID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Service, len(rows))
	for i, row := range rows {
		svc, err := mapServiceRow(row)
		if err != nil {
			return nil, err
		}
		out[i] = svc
	}
	return out, nil
}

func (r *ServiceRepo) Create(ctx context.Context, s *domain.Service) (*domain.Service, error) {
	var defaultID pgtype.Int4
	if s.DefaultServiceID != nil {
		defaultID = pgtype.Int4{Int32: *s.DefaultServiceID, Valid: true}
	}

	row, err := r.q.CreateService(ctx, db.CreateServiceParams{
		MasterID:         s.MasterID,
		DefaultServiceID: defaultID,
		Slug:             s.Slug,
		Name:             s.Name,
		Description:      pgtype.Text{String: s.Description, Valid: s.Description != ""},
		Price:            s.Price,
		Currency:         s.Currency,
		DurationMins:     s.DurationMins,
	})
	if err != nil {
		return nil, err
	}
	return mapServiceBasic(row), nil
}

func (r *ServiceRepo) Update(ctx context.Context, s *domain.Service) (*domain.Service, error) {
	row, err := r.q.UpdateService(ctx, db.UpdateServiceParams{
		ID:           s.ID,
		MasterID:     s.MasterID,
		Description:  pgtype.Text{String: s.Description, Valid: true},
		DurationMins: pgtype.Int4{Int32: s.DurationMins, Valid: true},
		Price:        pgtype.Int4{Int32: s.Price, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return mapServiceBasic(row), nil
}

func (r *ServiceRepo) Delete(ctx context.Context, id, masterID int32) error {
	return r.q.DeleteService(ctx, db.DeleteServiceParams{
		ID:       id,
		MasterID: masterID,
	})
}

func (r *ServiceRepo) ExistsByDefaultID(ctx context.Context, masterID, defaultServiceID int32) (bool, error) {
	return r.q.ExistsServiceByDefaultID(ctx, db.ExistsServiceByDefaultIDParams{
		MasterID:         masterID,
		DefaultServiceID: pgtype.Int4{Int32: defaultServiceID, Valid: true},
	})
}

func (r *ServiceRepo) GetDefaultTemplate(ctx context.Context, id int32) (*domain.DefaultServiceTemplate, error) {
	row, err := r.q.GetDefaultServiceForCopy(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrServiceNotFound
		}
		return nil, err
	}
	return &domain.DefaultServiceTemplate{
		ID:           row.ID,
		Slug:         row.Slug,
		Name:         row.Name,
		Description:  row.Description.String,
		Price:        row.Price,
		Currency:     row.Currency,
		DurationMins: row.DurationMins,
	}, nil
}

func (r *ServiceRepo) ReplaceLocations(ctx context.Context, serviceID int32, locs []domain.ServiceLocation) error {
	if err := r.q.ReplaceServiceLocations(ctx, serviceID); err != nil {
		return err
	}
	if len(locs) == 0 {
		return nil
	}

	type locDTO struct {
		LocationID   int32 `json:"location_id"`
		DurationMins int32 `json:"duration_mins"`
		Price        int32 `json:"price"`
		IsActive     bool  `json:"is_active"`
	}
	dtos := make([]locDTO, len(locs))
	for i, l := range locs {
		dtos[i] = locDTO{
			LocationID:   l.LocationID,
			DurationMins: l.DurationMins,
			Price:        l.Price,
			IsActive:     l.IsActive,
		}
	}
	data, _ := json.Marshal(dtos)

	return r.q.InsertServiceLocations(ctx, db.InsertServiceLocationsParams{
		ServiceID: serviceID,
		Data:      data,
	})
}

func (r *ServiceRepo) UpdateMasterMinPrice(ctx context.Context, masterID int32) error {
	return r.q.UpdateServicesMinPrice(ctx, masterID)
}

func mapServiceBasic(row db.Service) *domain.Service {
	return &domain.Service{
		ID:           row.ID,
		MasterID:     row.MasterID,
		Slug:         row.Slug,
		Name:         row.Name,
		Description:  row.Description.String,
		Price:        row.Price,
		Currency:     row.Currency,
		DurationMins: row.DurationMins,
		IsActive:     row.IsActive,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}

type locationsJSON struct {
	LocationID   int32 `json:"location_id"`
	DurationMins int32 `json:"duration_mins"`
	Price        int32 `json:"price"`
	IsActive     bool  `json:"is_active"`
}

func parseLocations(v any) ([]domain.ServiceLocation, error) {
	if v == nil {
		return nil, nil
	}

	var data []byte
	switch val := v.(type) {
	case []byte:
		data = val
	case string:
		data = []byte(val)
	case []interface{}:
		b, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		data = b
	default:
		return nil, fmt.Errorf("unexpected locations JSON type %T", v)
	}

	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}

	var raw []locationsJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	out := make([]domain.ServiceLocation, len(raw))
	for i, l := range raw {
		out[i] = domain.ServiceLocation{
			LocationID:   l.LocationID,
			DurationMins: l.DurationMins,
			Price:        l.Price,
			IsActive:     l.IsActive,
		}
	}
	return out, nil
}

func mapServiceRow(row db.GetMyServicesRow) (*domain.Service, error) {
	svc := &domain.Service{
		ID:           row.ID,
		MasterID:     row.MasterID,
		Slug:         row.Slug,
		Name:         row.Name,
		Description:  row.Description.String,
		Price:        row.Price,
		Currency:     row.Currency,
		DurationMins: row.DurationMins,
		IsActive:     row.IsActive,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
	locs, err := parseLocations(row.LocationsJson)
	if err != nil {
		return nil, err
	}
	svc.Locations = locs
	return svc, nil
}

func mapServiceFull(row db.GetServiceByIDRow) (*domain.Service, error) {
	svc := &domain.Service{
		ID:           row.ID,
		MasterID:     row.MasterID,
		Slug:         row.Slug,
		Name:         row.Name,
		Description:  row.Description.String,
		Price:        row.Price,
		Currency:     row.Currency,
		DurationMins: row.DurationMins,
		IsActive:     row.IsActive,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
	if row.DefaultServiceID.Valid {
		id := row.DefaultServiceID.Int32
		svc.DefaultServiceID = &id
	}
	locs, err := parseLocations(row.LocationsJson)
	if err != nil {
		return nil, err
	}
	svc.Locations = locs
	return svc, nil
}
