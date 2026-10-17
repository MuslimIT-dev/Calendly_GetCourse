package booking

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type ListMyInput struct {
	StatusFilter *domain.AppointmentStatus
	PageSize     int32
	PageToken    string
}

type ListMyOutput struct {
	Appointments  []*domain.AppointmentCard
	NextPageToken string
}

type ListMyUseCase struct {
	deps Deps
}

func NewListMyUseCase(deps Deps) *ListMyUseCase {
	return &ListMyUseCase{deps: deps}
}

func (uc *ListMyUseCase) Execute(ctx context.Context, in ListMyInput) (*ListMyOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	pageSize := normalizePageSize(in.PageSize)

	cursor, err := decodeCursor(in.PageToken)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	rows, err := uc.deps.Bookings.ListByClient(ctx, userID, in.StatusFilter, cursor, pageSize+1)
	if err != nil {
		return nil, err
	}

	var next string
	if len(rows) > int(pageSize) {
		rows = rows[:pageSize]
		last := rows[len(rows)-1]
		next = encodeCursor(&domain.AppointmentCursor{
			StartTime: last.Appointment.StartTime,
			ID:        last.Appointment.ID,
		})
	}

	return &ListMyOutput{Appointments: rows, NextPageToken: next}, nil
}

func normalizePageSize(s int32) int32 {
	if s <= 0 {
		return defaultPageSize
	}
	if s > maxPageSize {
		return maxPageSize
	}
	return s
}

func encodeCursor(c *domain.AppointmentCursor) string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(s string) (*domain.AppointmentCursor, error) {
	if s == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var c domain.AppointmentCursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func parseDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
