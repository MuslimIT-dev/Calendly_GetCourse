package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

const paymentTTL = 15 * time.Minute

type CreateInput struct {
	Purpose     domain.PaymentPurpose
	Provider    domain.PaymentProvider
	ReferenceID string
	Amount      int32
	Currency    string
	ReturnURL   string
}

type CreateOutput struct {
	Payment *domain.Payment
}

type CreateUseCase struct {
	deps    Deps
	gateway PaymentGateway
}

func NewCreateUseCase(deps Deps, gateway PaymentGateway) *CreateUseCase {
	return &CreateUseCase{deps: deps, gateway: gateway}
}

func (uc *CreateUseCase) Execute(ctx context.Context, in CreateInput) (*CreateOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	if in.ReferenceID == "" || in.Amount <= 0 || in.Currency == "" {
		return nil, domain.ErrInvalidInput
	}
	if in.Purpose == domain.PaymentPurpose(0) || in.Provider == domain.PaymentProvider(0) {
		return nil, domain.ErrInvalidInput
	}

	if existing, _ := uc.deps.Payments.GetPaidByReference(ctx, in.Purpose, in.ReferenceID); existing != nil {
		return &CreateOutput{Payment: existing}, nil
	}

	p, err := uc.deps.Payments.Create(ctx, &domain.Payment{
		UserID:      userID,
		Amount:      in.Amount,
		Currency:    in.Currency,
		Status:      domain.PaymentPending,
		Provider:    in.Provider,
		Purpose:     in.Purpose,
		ReferenceID: in.ReferenceID,
		ExternalID:  "",
		PaymentURL:  "",
		ExpiresAt:   time.Now().UTC().Add(paymentTTL),
	})
	if err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}

	externalID, paymentURL, err := uc.gateway.CreatePayment(ctx, p, in.ReturnURL)
	if err != nil {
		_, _ = uc.deps.Payments.MarkFailed(ctx, p.ID, err.Error())
		return nil, fmt.Errorf("gateway: %w", err)
	}

	updated, err := uc.deps.Payments.SetExternal(ctx, p.ID, externalID, paymentURL)
	_ = updated
	if err != nil {
		return nil, fmt.Errorf("update external: %w", err)
	}

	return &CreateOutput{Payment: p}, nil
}
