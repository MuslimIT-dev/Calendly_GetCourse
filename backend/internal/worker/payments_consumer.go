package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
	notificationuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/notification"
)

type PaymentEvent struct {
	PaymentID   int64     `json:"payment_id"`
	UserID      int32     `json:"user_id"`
	Purpose     int32     `json:"purpose"`
	ReferenceID string    `json:"reference_id"`
	Amount      int32     `json:"amount"`
	Currency    string    `json:"currency"`
	ExternalID  string    `json:"external_id"`
	Reason      string    `json:"reason"`
	PaidAt      time.Time `json:"paid_at"`
	RefundedAt  time.Time `json:"refunded_at"`
}

type PaymentsConsumer struct{ sendUC *notificationuc.SendUseCase }

func NewPaymentsConsumer(sendUC *notificationuc.SendUseCase) *PaymentsConsumer {
	return &PaymentsConsumer{sendUC: sendUC}
}

func (c *PaymentsConsumer) Consume(ctx context.Context, event *PaymentEvent) error {
	if event == nil || event.PaymentID <= 0 || event.UserID <= 0 {
		return fmt.Errorf("invalid payment event")
	}
	link := fmt.Sprintf("/payments/%d", event.PaymentID)
	amount := fmt.Sprintf("%d.%02d %s", event.Amount/100, event.Amount%100, event.Currency)
	var notificationType domain.NotificationType
	var title, body string
	switch {
	case !event.RefundedAt.IsZero():
		notificationType, title, body = domain.NotifSystem, "Возврат средств", fmt.Sprintf("Возврат %s выполнен", amount)
	case !event.PaidAt.IsZero():
		notificationType, title, body = domain.NotifPaymentSucceeded, "Оплата прошла", fmt.Sprintf("Платёж на %s успешно обработан", amount)
	case event.Reason != "":
		notificationType, title, body = domain.NotifPaymentFailed, "Ошибка оплаты", fmt.Sprintf("Платёж на %s не прошёл. Причина: %s", amount, event.Reason)
	default:
		return fmt.Errorf("unrecognized payment event for payment %d", event.PaymentID)
	}
	_, err := c.sendUC.Execute(ctx, notificationuc.SendInput{
		UserID: event.UserID, Type: notificationType, Channel: domain.ChannelInApp,
		Title: title, Body: body, Link: link,
		Metadata: map[string]any{"payment_id": event.PaymentID, "reference_id": event.ReferenceID},
	})
	return err
}
