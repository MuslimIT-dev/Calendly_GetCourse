package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/repository/db"
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
	return mapMasterFullByUserID(row), nil
}

func (r *MasterRepo) Create(ctx context.Context, p *domain.MasterProfile) (*domain.MasterProfile, error) {
	row, err := r.q.CreateMaster(ctx, db.CreateMasterParams{
		UserID:              p.UserID,
		Slug:                p.Slug,
		Bio:                 pgtype.Text{String: p.Bio, Valid: p.Bio != ""},
		Specialization:      p.Specialization,
		YearsOfExperience:   p.YearsOfExperience,
		IsAcceptingBookings: pgtype.Bool{Bool: p.IsAcceptingBookings, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return mapMasterBasic(row), nil
}

func (r *MasterRepo) Update(ctx context.Context, p *domain.MasterProfile) (*domain.MasterProfile, error) {
	row, err := r.q.UpdateMaster(ctx, db.UpdateMasterParams{
		ID:                  p.ID,
		Bio:                 pgtype.Text{String: p.Bio, Valid: true},
		Specialization:      pgtype.Text{String: p.Specialization, Valid: true},
		YearsOfExperience:   pgtype.Int4{Int32: p.YearsOfExperience, Valid: true},
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
	return r.q.MasterSlugExists(ctx, slug)
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
	data, err := json.Marshal(dtos)
	if err != nil {
		return err
	}

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
	data, err := json.Marshal(dtos)
	if err != nil {
		return err
	}

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
		params, err := buildRatingParams(f, c, limit)
		if err != nil {
			return nil, err
		}
		rows, err := r.q.ListMastersByRating(ctx, params)
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

func buildExperienceParams(f domain.MasterFilter, c *domain.ListCursor, limit int32) db.ListMastersByExperienceParams {
	params := db.ListMastersByExperienceParams{
		MinPrice:       int4Value(f.MinPrice),
		MaxPrice:       int4Value(f.MaxPrice),
		MinExperience:  int4Value(f.MinExperience),
		Specialization: textValue(f.Specialization),
		MasterTimezone: textValue(f.MasterTimezone),
		ServiceID:      int4Value(f.ServiceID),
		PageSize:       limit,
	}
	if c != nil && c.IntValue != nil {
		params.CursorExperience = pgtype.Int4{Int32: *c.IntValue, Valid: true}
		params.CursorID = pgtype.Int4{Int32: c.ID, Valid: true}
	}
	return params
}

func buildPriceParams(f domain.MasterFilter, c *domain.ListCursor, limit int32) db.ListMastersByPriceParams {
	params := db.ListMastersByPriceParams{
		MinPrice:       int4Value(f.MinPrice),
		MaxPrice:       int4Value(f.MaxPrice),
		MinExperience:  int4Value(f.MinExperience),
		Specialization: textValue(f.Specialization),
		MasterTimezone: textValue(f.MasterTimezone),
		ServiceID:      int4Value(f.ServiceID),
		PageSize:       limit,
	}
	if c != nil && c.IntValue != nil {
		params.CursorPrice = pgtype.Int4{Int32: *c.IntValue, Valid: true}
		params.CursorID = pgtype.Int4{Int32: c.ID, Valid: true}
	}
	return params
}

func buildRatingParams(f domain.MasterFilter, c *domain.ListCursor, limit int32) (db.ListMastersByRatingParams, error) {
	params := db.ListMastersByRatingParams{
		MinPrice:       int4Value(f.MinPrice),
		MaxPrice:       int4Value(f.MaxPrice),
		MinExperience:  int4Value(f.MinExperience),
		Specialization: textValue(f.Specialization),
		MasterTimezone: textValue(f.MasterTimezone),
		ServiceID:      int4Value(f.ServiceID),
		PageSize:       limit,
	}
	if c != nil && c.TimeValue != nil {
		rating := pgtype.Numeric{}
		if err := rating.ScanScientific(strconv.FormatFloat(*c.TimeValue, 'f', -1, 64)); err != nil {
			return params, err
		}
		params.CursorRating = rating
		params.CursorID = pgtype.Numeric{
			Int:   big.NewInt(int64(c.ID)),
			Valid: true,
		}
	}
	return params, nil
}

func int4Value(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}

func textValue(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
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
	return mapMasterProfileFields(
		row.ID, row.UserID, row.Slug, row.Bio, row.Specialization,
		row.YearsOfExperience, row.IsAcceptingBookings, row.DefaultLocationID,
		row.AvgRating, row.ReviewsCount, row.CreatedAt, row.UpdatedAt,
	)
}

func mapMasterFull(row db.GetMasterBySlugRow) *domain.MasterProfile {
	return mapMasterProfileFields(
		row.ID, row.UserID, row.Slug, row.Bio, row.Specialization,
		row.YearsOfExperience, row.IsAcceptingBookings, row.DefaultLocationID,
		row.AvgRating, row.ReviewsCount, row.CreatedAt, row.UpdatedAt,
	)
}

func mapMasterFullByUserID(row db.GetMasterByUserIDRow) *domain.MasterProfile {
	return mapMasterProfileFields(
		row.ID, row.UserID, row.Slug, row.Bio, row.Specialization,
		row.YearsOfExperience, row.IsAcceptingBookings, row.DefaultLocationID,
		row.AvgRating, row.ReviewsCount, row.CreatedAt, row.UpdatedAt,
	)
}

func mapMasterProfileFields(
	id, userID int32,
	slug string,
	bio pgtype.Text,
	specialization string,
	yearsOfExperience int32,
	isAcceptingBookings pgtype.Bool,
	defaultLocationID pgtype.Int4,
	avgRating pgtype.Numeric,
	reviewsCount pgtype.Int4,
	createdAt, updatedAt pgtype.Timestamptz,
) *domain.MasterProfile {
	return &domain.MasterProfile{
		ID:                  id,
		UserID:              userID,
		Slug:                slug,
		Bio:                 bio.String,
		Specialization:      specialization,
		YearsOfExperience:   yearsOfExperience,
		IsAcceptingBookings: isAcceptingBookings.Bool,
		DefaultLocationID:   int4Ptr(defaultLocationID),
		AvgRating:           numericToFloat32(avgRating),
		ReviewsCount:        reviewsCount.Int32,
		CreatedAt:           createdAt.Time,
		UpdatedAt:           updatedAt.Time,
	}
}

func mapMasterUserFull(row db.GetMasterBySlugRow) *domain.User {
	return mapMasterUserFields(row.UserID, row.UserName, row.UserEmail, row.UserAvatarUrl, row.UserTimezone, row.UserEmailVerified, row.UserRoleIds)
}

func mapMasterUserFullByUserID(row db.GetMasterByUserIDRow) *domain.User {
	return mapMasterUserFields(row.UserID, row.UserName, row.UserEmail, row.UserAvatarUrl, row.UserTimezone, row.UserEmailVerified, row.UserRoleIds)
}

func mapMasterUserFields(id int32, name pgtype.Text, email string, avatarURL, timezone pgtype.Text, emailVerified pgtype.Bool, roleIDs []int32) *domain.User {
	return &domain.User{
		ID:            id,
		Name:          name.String,
		Email:         email,
		AvatarURL:     avatarURL.String,
		Timezone:      timezone.String,
		EmailVerified: emailVerified.Bool,
		Roles:         convertRoleIDsToRoles(roleIDs),
	}
}

func mapCardFromExperience(row db.ListMastersByExperienceRow) *domain.MasterCardData {
	return mapMasterCard(row.ID, row.UserID, row.Slug, row.Bio, row.Specialization, row.YearsOfExperience,
		row.IsAcceptingBookings, row.DefaultLocationID, row.AvgRating, row.ReviewsCount,
		row.CreatedAt, row.UpdatedAt, row.UserName, row.UserEmail, row.UserAvatarUrl,
		row.UserTimezone, row.UserEmailVerified, row.UserRoleIds, row.MinPrice, row.TotalServices)
}

func mapCardFromPrice(row db.ListMastersByPriceRow) *domain.MasterCardData {
	return mapMasterCard(row.ID, row.UserID, row.Slug, row.Bio, row.Specialization, row.YearsOfExperience,
		row.IsAcceptingBookings, row.DefaultLocationID, row.AvgRating, row.ReviewsCount,
		row.CreatedAt, row.UpdatedAt, row.UserName, row.UserEmail, row.UserAvatarUrl,
		row.UserTimezone, row.UserEmailVerified, row.UserRoleIds, row.MinPrice, row.TotalServices)
}

func mapCardFromRating(row db.ListMastersByRatingRow) *domain.MasterCardData {
	return mapMasterCard(row.ID, row.UserID, row.Slug, row.Bio, row.Specialization, row.YearsOfExperience,
		row.IsAcceptingBookings, row.DefaultLocationID, row.AvgRating, row.ReviewsCount,
		row.CreatedAt, row.UpdatedAt, row.UserName, row.UserEmail, row.UserAvatarUrl,
		row.UserTimezone, row.UserEmailVerified, row.UserRoleIds, row.MinPrice, row.TotalServices)
}

func mapMasterCard(
	id, userID int32,
	slug string,
	bio pgtype.Text,
	specialization string,
	yearsOfExperience int32,
	isAcceptingBookings pgtype.Bool,
	defaultLocationID pgtype.Int4,
	avgRating pgtype.Numeric,
	reviewsCount pgtype.Int4,
	createdAt, updatedAt pgtype.Timestamptz,
	userName pgtype.Text,
	userEmail string,
	userAvatarURL, userTimezone pgtype.Text,
	userEmailVerified pgtype.Bool,
	userRoleIDs []int32,
	minPrice pgtype.Int4,
	totalServices int64,
) *domain.MasterCardData {
	profile := mapMasterProfileFields(
		id, userID, slug, bio, specialization, yearsOfExperience,
		isAcceptingBookings, defaultLocationID, avgRating, reviewsCount, createdAt, updatedAt,
	)
	user := mapMasterUserFields(userID, userName, userEmail, userAvatarURL, userTimezone, userEmailVerified, userRoleIDs)
	return &domain.MasterCardData{
		Profile:       profile,
		User:          user,
		MinPrice:      minPrice.Int32,
		TotalServices: int32(totalServices),
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
