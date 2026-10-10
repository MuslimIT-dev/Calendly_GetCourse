package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/repository/postgres/db"
)

type MasterRepo struct {
	q *db.Queries
}

func NewMasterRepo(q *db.Queries) *MasterRepo {
	return &MasterRepo{q: q}
}

var _ domain.MasterRepository = (*MasterRepo)(nil)

func (r *MasterRepo) GetBySlug(ctx context.Context, slug string) (*domain.MasterProfile, *domain.User, error) {
	row, err := r.q.GetMasterBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.ErrMasterNotFound
		}
		return nil, nil, err
	}
	return mapMasterFull(row), mapMasterUserFull(row), nil
}

func (r *MasterRepo) GetByUserID(ctx context.Context, userID int32) (*domain.MasterProfile, error) {
	row, err := r.q.GetMasterByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMasterNotFound
		}
		return nil, err
	}
	return mapMasterFull(row), nil
}

func (r *MasterRepo) Create(ctx context.Context, p *domain.MasterProfile) (*domain.MasterProfile, error) {
	row, err := r.q.CreateMaster(ctx, db.CreateMasterParams{
		UserID:              p.UserID,
		Slug:                p.Slug,
		Bio:                 pgtype.Text{String: p.Bio, Valid: p.Bio != ""},
		Specialization:      p.Specialization,
		YearsOfExperience:   p.YearsOfExperience,
		IsAcceptingBookings: p.IsAcceptingBookings,
	})
	if err != nil {
		return nil, err
	}
	return mapMasterBasic(row), nil
}

func (r *MasterRepo) Update(ctx context.Context, p *domain.MasterProfile) (*domain.MasterProfile, error) {
	row, err := r.q.UpdateMaster(ctx, db.UpdateMasterParams{
		ID: p.ID,
		Bio: pgtype.Text{String: p.Bio, Valid: true},
		Specialization: pgtype.Text{String: p.Specialization, Valid: true},
		YearsOfExperience: pgtype.Int4{Int32: p.YearsOfExperience, Valid: true},
		IsAcceptingBookings: pgtype.Bool{Bool: p.IsAcceptingBookings, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return mapMasterBasic(row), nil
}

func (r *MasterRepo) UpdateSlug(ctx context.Context, id int32, slug string) error {
	_, err := r.q.UpdateMasterSlug(ctx, db.UpdateMasterSlugParams{ID: id, Slug: slug})
	return err
}

func (r *MasterRepo) SlugExists(ctx context.Context, slug string, excludeID int32) (bool, error) {
	var excl pgtype.Int4
	if excludeID > 0 {
		excl = pgtype.Int4{Int32: excludeID, Valid: true}
	}
	return r.q.MasterSlugExists(ctx, db.MasterSlugExistsParams{Slug: slug, ExcludeID: excl})
}

func (r *MasterRepo) ReplaceLanguages(ctx context.Context, masterID int32, languages []domain.Language) error {
	if err := r.q.ReplaceMasterLanguages(ctx, masterID); err != nil {
		return err
	}
	if len(languages) == 0 {
		return nil
	}

	type langDTO struct {
		Language    string `json:"language"`
		Proficiency string `json:"proficiency"`
	}
	dtos := make([]langDTO, len(languages))
	for i, l := range languages {
		dtos[i] = langDTO{Language: l.Name, Proficiency: proficiencyToString(l.Proficiency)}
	}
	data, _ := json.Marshal(dtos)

	return r.q.InsertMasterLanguages(ctx, db.InsertMasterLanguagesParams{
		MasterID: masterID,
		Data:     data,
	})
}

func (r *MasterRepo) ReplaceCertificates(ctx context.Context, masterID int32, certs []domain.Certificate) error {
	if err := r.q.ReplaceMasterCertificates(ctx, masterID); err != nil {
		return err
	}
	if len(certs) == 0 {
		return nil
	}

	type certDTO struct {
		Name         string `json:"name"`
		Organization string `json:"organization"`
		Year         int32  `json:"year"`
		FileUrl      string `json:"file_url"`
	}
	dtos := make([]certDTO, len(certs))
	for i, c := range certs {
		dtos[i] = certDTO{Name: c.Name, Organization: c.Organization, Year: c.Year, FileUrl: c.FileURL}
	}
	data, _ := json.Marshal(dtos)

	return r.q.InsertMasterCertificates(ctx, db.InsertMasterCertificatesParams{
		MasterID: masterID,
		Data:     data,
	})
}

func (r *MasterRepo) ListCards(ctx context.Context, f domain.MasterFilter, c *domain.ListCursor, limit int32) ([]*domain.MasterCardData, error) {
	switch f.SortBy {
	case domain.SortTypeExperience:
		rows, err := r.q.ListMastersByExperience(ctx, buildExperienceParams(f, c, limit))
		if err != nil {
			return nil, err
		}
		out := make([]*domain.MasterCardData, len(rows))
		for i, row := range rows {
			out[i] = mapCardFromExperience(row)
		}
		return out, nil
	case domain.SortTypePrice:
		rows, err := r.q.ListMastersByPrice(ctx, buildPriceParams(f, c, limit))
		if err != nil {
			return nil, err
		}
		out := make([]*domain.MasterCardData, len(rows))
		for i, row := range rows {
			out[i] = mapCardFromPrice(row)
		}
		return out, nil
	default:
		rows, err := r.q.ListMastersByRating(ctx, buildRatingParams(f, c, limit))
		if err != nil {
			return nil, err
		}
		out := make([]*domain.MasterCardData, len(rows))
		for i, row := range rows {
			out[i] = mapCardFromRating(row)
		}
		return out, nil
	}
}

func proficiencyToString(p domain.ProficiencyLevel) string {
	switch p {
	case domain.ProficiencyBeginner:
		return "beginner"
	case domain.ProficiencyIntermediate:
		return "intermediate"
	case domain.ProficiencyAdvanced:
		return "advanced"
	}
	return "beginner"
}

func proficiencyFromString(s string) domain.ProficiencyLevel {
	switch s {
	case "beginner":
		return domain.ProficiencyBeginner
	case "intermediate":
		return domain.ProficiencyIntermediate
	case "advanced":
		return domain.ProficiencyAdvanced
	}
	return domain.ProficiencyUnspecified
}

func mapMasterBasic(row db.Master) *domain.MasterProfile {
	return &domain.MasterProfile{
		ID:                  row.ID,
		UserID:              row.UserID,
		Slug:                row.Slug,
		Bio:                 row.Bio.String,
		Specialization:      row.Specialization,
		YearsOfExperience:   row.YearsOfExperience,
		IsAcceptingBookings: row.IsAcceptingBookings.Bool,
		DefaultLocationID:   int4Ptr(row.DefaultLocationID),
		AvgRating:           numericToFloat32(row.AvgRating),
		ReviewsCount:        row.ReviewsCount.Int32,
		CreatedAt:           row.CreatedAt.Time,
		UpdatedAt:           row.UpdatedAt.Time,
	}
}

func int4Ptr(v pgtype.Int4) *int32 {
	if !v.Valid {
		return nil
	}
	return &v.Int32
}

func numericToFloat32(v pgtype.Numeric) float32 {
	f, _ := v.Float64Value()
	return float32(f.Float64)
}

func toTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}