package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
	paymentuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/payment"
)

type ManualGateway struct {
	secret []byte
}

func NewManualGateway() *ManualGateway {
	return &ManualGateway{secret: []byte("demo-secret")}
}

func (g *ManualGateway) CreatePayment(ctx context.Context, p *domain.Payment, returnURL string) (string, string, error) {
	externalID := uuid.NewString()
	paymentURL := fmt.Sprintf("https://demo.pay/%s?return=%s", externalID, returnURL)
	return externalID, paymentURL, nil
}

func (g *ManualGateway) Refund(ctx context.Context, externalID string, amount int32) error {
	return nil
}

func (g *ManualGateway) VerifyWebhook(provider domain.PaymentProvider, payload []byte, signature string) error {
	mac := hmac.New(sha256.New, g.secret)
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return errors.New("signature mismatch")
	}
	return nil
}

func (g *ManualGateway) ParseWebhook(provider domain.PaymentProvider, payload []byte) (*paymentuc.WebhookEvent, error) {
	var raw struct {
		ExternalID    string `json:"external_id"`
		Status        string `json:"status"`
		ReceiptURL    string `json:"receipt_url"`
		FailureReason string `json:"failure_reason"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}

	var status domain.PaymentStatus
	switch raw.Status {
	case "paid":
		status = domain.PaymentPaid
	case "failed":
		status = domain.PaymentFailed
	default:
		return nil, errors.New("unknown status")
	}

	return &paymentuc.WebhookEvent{
		ExternalID:    raw.ExternalID,
		Status:        status,
		ReceiptURL:    raw.ReceiptURL,
		FailureReason: raw.FailureReason,
	}, nil
}