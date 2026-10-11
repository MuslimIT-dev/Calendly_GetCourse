package payment

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type WebhookInput struct {
	Provider  domain.PaymentProvider
	Payload   []byte
	Signature string
}

type WebhookOutput struct {
	Accepted bool
}

type WebhookUseCase struct {
	deps    Deps
	gateway PaymentGateway
	events  EventPublisher
}

func NewWebhookUseCase(deps Deps, gateway PaymentGateway, events EventPublisher) *WebhookUseCase {
	return &WebhookUseCase{deps: deps, gateway: gateway, events: events}
}

func (uc *WebhookUseCase) Execute(ctx context.Context, in WebhookInput) (*WebhookOutput, error) {
	if err := uc.gateway.VerifyWebhook(in.Provider, in.Payload, in.Signature); err != nil {
		return nil, domain.ErrPaymentSignatureBad
	}

	event, err := uc.gateway.ParseWebhook(in.Provider, in.Payload)
	if err != nil {
		return nil, err
	}

	card, err := uc.deps.Payments.GetByExternalID(ctx, in.Provider, event.ExternalID)
	if err != nil {
		return &WebhookOutput{Accepted: false}, nil
	}
	p := card.Payment

	switch event.Status {
	case domain.PaymentPaid:
		updated, err := uc.deps.Payments.MarkPaid(ctx, p.ID, event.ExternalID, event.ReceiptURL, p.PaymentURL)
		if err != nil {
			return &WebhookOutput{Accepted: false}, nil
		}
		_ = publishEvent(ctx, uc.events, PaymentSucceededEvent{
			PaymentID:   updated.ID,
			UserID:      updated.UserID,
			Purpose:     int32(updated.Purpose),
			ReferenceID: updated.ReferenceID,
			Amount:      updated.Amount,
			Currency:    updated.Currency,
			ExternalID:  updated.ExternalID,
			PaidAt:      *updated.PaidAt,
		})
	case domain.PaymentFailed:
		updated, err := uc.deps.Payments.MarkFailed(ctx, p.ID, event.FailureReason)
		if err != nil {
			return &WebhookOutput{Accepted: false}, nil
		}
		_ = publishEvent(ctx, uc.events, PaymentFailedEvent{
			PaymentID:   updated.ID,
			UserID:      updated.UserID,
			Purpose:     int32(updated.Purpose),
			ReferenceID: updated.ReferenceID,
			Reason:      event.FailureReason,
		})
	}

	return &WebhookOutput{Accepted: true}, nil
}
