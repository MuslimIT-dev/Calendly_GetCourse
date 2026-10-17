package payment

import (
	"context"
	"time"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type ListAllInput struct {
	StatusFilter *domain.PaymentStatus
	UserID       *int32
	FromDate     string
	ToDate       string
	PageSize     int32
	PageToken    string
}

type ListAllOutput struct {
	Payments      []*domain.Payment
	NextPageToken string
}

type ListAllUseCase struct {
	deps Deps
}

func NewListAllUseCase(deps Deps) *ListAllUseCase {
	return &ListAllUseCase{deps: deps}
}

func (uc *ListAllUseCase) Execute(ctx context.Context, in ListAllInput) (*ListAllOutput, error) {
	if !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return nil, domain.ErrForbidden
	}

	from, err := parseDate(in.FromDate)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	to, err := parseDate(in.ToDate)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	if to != nil {
		next := to.AddDate(0, 0, 1)
		to = &next
	}

	pageSize := normalizePageSize(in.PageSize)

	cursor, err := decodePaymentCursor(in.PageToken)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	rows, err := uc.deps.Payments.ListAll(ctx, in.StatusFilter, in.UserID, from, to, cursor, pageSize+1)
	if err != nil {
		return nil, err
	}

	var next string
	if len(rows) > int(pageSize) {
		rows = rows[:pageSize]
		last := rows[len(rows)-1]
		next = encodePaymentCursor(&domain.PaymentCursor{
			CreatedAt: last.Payment.CreatedAt,
			ID:        last.Payment.ID,
		})
	}

	payments := make([]*domain.Payment, len(rows))
	for i, r := range rows {
		payments[i] = r.Payment
	}

	return &ListAllOutput{Payments: payments, NextPageToken: next}, nil
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
