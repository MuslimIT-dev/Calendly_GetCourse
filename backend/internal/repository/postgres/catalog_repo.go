package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
	"github.com/MuslimIT-dev/zaman/backend/internal/repository/db"
)

type CatalogRepo struct {
	q *db.Queries
}

func NewCatalogRepo(q *db.Queries) *CatalogRepo {
	return &CatalogRepo{q: q}
}

var _ domain.CatalogRepository = (*CatalogRepo)(nil)

func (r *CatalogRepo) ListCategories(ctx context.Context) ([]*domain.Category, error) {
	rows, err := r.q.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Category, len(rows))
	for i, row := range rows {
		out[i] = mapCategory(row)
	}
	return out, nil
}

func (r *CatalogRepo) GetCategory(ctx context.Context, id int32) (*domain.Category, error) {
	row, err := r.q.GetCategory(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return mapCategory(row), nil
}

func (r *CatalogRepo) CreateCategory(ctx context.Context, c *domain.Category) (*domain.Category, error) {
	row, err := r.q.CreateCategory(ctx, c.Name)
	if err != nil {
		return nil, err
	}
	return mapCategory(row), nil
}

func (r *CatalogRepo) UpdateCategory(ctx context.Context, c *domain.Category) (*domain.Category, error) {
	row, err := r.q.UpdateCategory(ctx, db.UpdateCategoryParams{
		ID:   c.ID,
		Name: c.Name,
	})
	if err != nil {
		return nil, err
	}
	return mapCategory(row), nil
}

func (r *CatalogRepo) DeleteCategory(ctx context.Context, id int32) error {
	return r.q.DeleteCategory(ctx, id)
}

func (r *CatalogRepo) CategoryHasServices(ctx context.Context, id int32) (bool, error) {
	return r.q.CategoryHasServices(ctx, id)
}

func (r *CatalogRepo) ListDefaultServices(ctx context.Context, categoryID *int32) ([]*domain.DefaultService, error) {
	var cid pgtype.Int4
	if categoryID != nil {
		cid = pgtype.Int4{Int32: *categoryID, Valid: true}
	}
	rows, err := r.q.ListDefaultServices(ctx, cid)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.DefaultService, len(rows))
	for i, row := range rows {
		out[i] = mapDefaultService(row)
	}
	return out, nil
}

func (r *CatalogRepo) GetDefaultService(ctx context.Context, id int32) (*domain.DefaultService, error) {
	row, err := r.q.GetDefaultService(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDefaultServiceNotFound
		}
		return nil, err
	}
	return mapDefaultService(row), nil
}

func (r *CatalogRepo) CreateDefaultService(ctx context.Context, s *domain.DefaultService) (*domain.DefaultService, error) {
	row, err := r.q.CreateDefaultService(ctx, db.CreateDefaultServiceParams{
		CategoryID:   s.CategoryID,
		Slug:         s.Slug,
		Name:         s.Name,
		Description:  pgtype.Text{String: s.Description, Valid: s.Description != ""},
		Price:        s.Price,
		Currency:     s.Currency,
		DurationMins: s.DurationMins,
	})
	if err != nil {
		return nil, err
	}
	return mapDefaultService(row), nil
}

func (r *CatalogRepo) UpdateDefaultService(ctx context.Context, s *domain.DefaultService) (*domain.DefaultService, error) {
	row, err := r.q.UpdateDefaultService(ctx, db.UpdateDefaultServiceParams{
		ID:           s.ID,
		Slug:         pgtype.Text{String: s.Slug, Valid: true},
		Name:         pgtype.Text{String: s.Name, Valid: true},
		Description:  pgtype.Text{String: s.Description, Valid: true},
		DurationMins: pgtype.Int4{Int32: s.DurationMins, Valid: true},
		Price:        pgtype.Int4{Int32: s.Price, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return mapDefaultService(row), nil
}

func (r *CatalogRepo) DeleteDefaultService(ctx context.Context, id int32) error {
	return r.q.DeleteDefaultService(ctx, id)
}

func (r *CatalogRepo) DefaultServiceSlugExists(ctx context.Context, slug string, excludeID int32) (bool, error) {
	var excl pgtype.Int4
	if excludeID > 0 {
		excl = pgtype.Int4{Int32: excludeID, Valid: true}
	}
	return r.q.DefaultServiceSlugExists(ctx, db.DefaultServiceSlugExistsParams{
		Slug:      slug,
		ExcludeID: excl,
	})
}

func mapCategory(row db.Category) *domain.Category {
	return &domain.Category{
		ID:        row.ID,
		Name:      row.Name,
		IsActive:  row.IsActive.Bool,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

func mapDefaultService(row db.DefaultService) *domain.DefaultService {
	return &domain.DefaultService{
		ID:           row.ID,
		CategoryID:   row.CategoryID,
		Slug:         row.Slug,
		Name:         row.Name,
		Description:  row.Description.String,
		Price:        row.Price,
		Currency:     row.Currency,
		DurationMins: row.DurationMins,
		IsActive:     row.IsActive.Bool,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}
