package domain

import (
	"context"
	"time"
)

type PaymentStatus int32

const (
	PaymentPending  PaymentStatus = 1
	PaymentPaid     PaymentStatus = 2
	PaymentFailed   PaymentStatus = 3
	PaymentRefunded PaymentStatus = 4
	PaymentExpired  PaymentStatus = 5
)

type PaymentProvider int32

const (
	ProviderYooKassa PaymentProvider = 1
	ProviderStripe   PaymentProvider = 2
	ProviderTinkoff  PaymentProvider = 3
	ProviderManual   PaymentProvider = 4
)

type PaymentPurpose int32

const (
	PurposeBooking      PaymentPurpose = 1
	PurposeCourse       PaymentPurpose = 2
	PurposeSubscription PaymentPurpose = 3
)

type Payment struct {
	ID            int64
	UserID        int32
	Amount        int32
	Currency      string
	Status        PaymentStatus
	Provider      PaymentProvider
	Purpose       PaymentPurpose
	ReferenceID   string
	ExternalID    string
	PaymentURL    string
	FailureReason string
	ReceiptURL    string
	PaidAt        *time.Time
	RefundedAt    *time.Time
	ExpiresAt     time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PaymentCard struct {
	Payment    *Payment
	UserName   string
	UserEmail  string
	UserAvatar string
}

type PaymentCursor struct {
	CreatedAt time.Time
	ID        int64
}

type PaymentRepository interface {
	Create(ctx context.Context, p *Payment) (*Payment, error)
	SetExternal(ctx context.Context, id int64, externalID, paymentURL string) (*Payment, error)
	GetByID(ctx context.Context, id int64) (*PaymentCard, error)
	GetByExternalID(ctx context.Context, provider PaymentProvider, externalID string) (*PaymentCard, error)
	GetPaidByReference(ctx context.Context, purpose PaymentPurpose, referenceID string) (*Payment, error)
	ListByUser(ctx context.Context, userID int32, status *PaymentStatus, purpose *PaymentPurpose, cursor *PaymentCursor, limit int32) ([]*PaymentCard, error)
	ListAll(ctx context.Context, status *PaymentStatus, userID *int32, from, to *time.Time, cursor *PaymentCursor, limit int32) ([]*PaymentCard, error)
	MarkPaid(ctx context.Context, id int64, externalID, receiptURL, paymentURL string) (*Payment, error)
	MarkFailed(ctx context.Context, id int64, reason string) (*Payment, error)
	MarkRefunded(ctx context.Context, id int64) (*Payment, error)
	Cancel(ctx context.Context, id int64) (*Payment, error)
	ExpirePending(ctx context.Context) (int64, error)
}
