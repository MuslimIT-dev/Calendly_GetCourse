package payment

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type Deps struct {
	Payments domain.PaymentRepository
	Users    domain.UserRepository
}

type EventPublisher interface {
	Publish(ctx context.Context, topic string, event *any) error
}

func publishEvent(ctx context.Context, publisher EventPublisher, event any) error {
	return publisher.Publish(ctx, "payments", &event)
}

type PaymentGateway interface {
	CreatePayment(ctx context.Context, p *domain.Payment, returnURL string) (externalID string, paymentURL string, err error)
	Refund(ctx context.Context, externalID string, amount int32) error
	VerifyWebhook(provider domain.PaymentProvider, payload []byte, signature string) error
	ParseWebhook(provider domain.PaymentProvider, payload []byte) (*WebhookEvent, error)
}

type WebhookEvent struct {
	ExternalID    string
	Status        domain.PaymentStatus
	ReceiptURL    string
	FailureReason string
}
