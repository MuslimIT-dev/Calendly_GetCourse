package payment

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type ListMyInput struct {
	StatusFilter  *domain.PaymentStatus
	PurposeFilter *domain.PaymentPurpose
	PageSize      int32
	PageToken     string
}

type ListMyOutput struct {
	Payments      []*domain.Payment
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

	cursor, err := decodePaymentCursor(in.PageToken)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	rows, err := uc.deps.Payments.ListByUser(ctx, userID, in.StatusFilter, in.PurposeFilter, cursor, pageSize+1)
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

	return &ListMyOutput{Payments: payments, NextPageToken: next}, nil
}

func normalizePageSize(s int32) int32 {
	if s <= 0 {
		return 20
	}
	if s > 100 {
		return 100
	}
	return s
}

func encodePaymentCursor(c *domain.PaymentCursor) string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodePaymentCursor(s string) (*domain.PaymentCursor, error) {
	if s == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var c domain.PaymentCursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
