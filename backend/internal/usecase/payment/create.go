package payment

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

const paymentTTL = 15 * time.Minute

type CreateInput struct {
	Purpose     domain.PaymentPurpose
	Provider    domain.PaymentProvider
	ReferenceID string
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

	if in.ReferenceID == "" {
		return nil, domain.ErrInvalidInput
	}
	if in.Purpose == 0 || in.Provider == 0 {
		return nil, domain.ErrInvalidInput
	}

	if existing, _ := uc.deps.Payments.GetPaidByReference(ctx, in.Purpose, in.ReferenceID); existing != nil {
		return &CreateOutput{Payment: existing}, nil
	}

	var amount int32
	var currency string

	switch in.Purpose {
	case domain.PurposeBooking:
		refID, err := strconv.ParseInt(in.ReferenceID, 10, 32)
		if err != nil {
			return nil, domain.ErrInvalidInput
		}
		card, err := uc.deps.Bookings.GetByID(ctx, int64(refID))
		if err != nil {
			return nil, err
		}
		if card.Appointment.ClientID != userID {
			return nil, domain.ErrForbidden
		}
		amount = card.Appointment.Price
		currency = card.Appointment.Currency
	case domain.PurposeCourse:
		return nil, domain.ErrInvalidInput // не реализовано
	default:
		return nil, domain.ErrInvalidInput
	}

	if amount <= 0 || currency == "" {
		return nil, domain.ErrInvalidInput
	}

	p, err := uc.deps.Payments.Create(ctx, &domain.Payment{
		UserID:      userID,
		Amount:      amount,
		Currency:    currency,
		Status:      domain.PaymentPending,
		Provider:    in.Provider,
		Purpose:     in.Purpose,
		ReferenceID: in.ReferenceID,
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
	if err != nil {
		return nil, fmt.Errorf("set external: %w", err)
	}

	return &CreateOutput{Payment: updated}, nil
}
