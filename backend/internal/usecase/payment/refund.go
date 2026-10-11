package payment

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type RefundInput struct {
	ID     int64
	Reason string
}

type RefundOutput struct {
	Payment *domain.Payment
}

type RefundUseCase struct {
	deps    Deps
	gateway PaymentGateway
	events  EventPublisher
}

func NewRefundUseCase(deps Deps, gateway PaymentGateway, events EventPublisher) *RefundUseCase {
	return &RefundUseCase{deps: deps, gateway: gateway, events: events}
}

func (uc *RefundUseCase) Execute(ctx context.Context, in RefundInput) (*RefundOutput, error) {
	if !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return nil, domain.ErrForbidden
	}

	card, err := uc.deps.Payments.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	p := card.Payment

	if p.Status != domain.PaymentPaid {
		return nil, domain.ErrPaymentInvalidStatus
	}
	if p.Provider != domain.ProviderManual {
		if err := uc.gateway.Refund(ctx, p.ExternalID, p.Amount); err != nil {
			return nil, fmt.Errorf("gateway refund: %w", err)
		}
	}

	updated, err := uc.deps.Payments.MarkRefunded(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	_ = publishEvent(ctx, uc.events, PaymentRefundedEvent{
		PaymentID:   updated.ID,
		UserID:      updated.UserID,
		Purpose:     int32(updated.Purpose),
		ReferenceID: updated.ReferenceID,
		Amount:      updated.Amount,
		Currency:    updated.Currency,
		RefundedAt:  *updated.RefundedAt,
	})

	return &RefundOutput{Payment: updated}, nil
}
