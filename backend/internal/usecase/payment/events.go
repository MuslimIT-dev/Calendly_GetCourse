package payment

import "time"

type PaymentSucceededEvent struct {
	PaymentID   int64     `json:"payment_id"`
	UserID      int32     `json:"user_id"`
	Purpose     int32     `json:"purpose"`
	ReferenceID string    `json:"reference_id"`
	Amount      int32     `json:"amount"`
	Currency    string    `json:"currency"`
	ExternalID  string    `json:"external_id"`
	PaidAt      time.Time `json:"paid_at"`
}

type PaymentFailedEvent struct {
	PaymentID   int64  `json:"payment_id"`
	UserID      int32  `json:"user_id"`
	Purpose     int32  `json:"purpose"`
	ReferenceID string `json:"reference_id"`
	Reason      string `json:"reason"`
}

type PaymentRefundedEvent struct {
	PaymentID   int64     `json:"payment_id"`
	UserID      int32     `json:"user_id"`
	Purpose     int32     `json:"purpose"`
	ReferenceID string    `json:"reference_id"`
	Amount      int32     `json:"amount"`
	Currency    string    `json:"currency"`
	RefundedAt  time.Time `json:"refunded_at"`
}
