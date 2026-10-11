package domain

import (
	"context"
	"time"
)

type AppointmentStatus int32

const (
	AppointmentPending           AppointmentStatus = 1
	AppointmentConfirmed         AppointmentStatus = 2
	AppointmentCancelledByClient AppointmentStatus = 3
	AppointmentCancelledByMaster AppointmentStatus = 4
	AppointmentCompleted         AppointmentStatus = 5
	AppointmentNoShow            AppointmentStatus = 6
)

const (
	AppointmentPaymentPending  PaymentStatus = 1
	AppointmentPaymentPaid     PaymentStatus = 2
	AppointmentPaymentRefunded PaymentStatus = 3
	AppointmentPaymentFailed   PaymentStatus = 4
)

type CancelledByRole int32

const (
	CancelledByUnspecified CancelledByRole = 0
	CancelledByClient      CancelledByRole = 1
	CancelledByMaster      CancelledByRole = 2
	CancelledByAdmin       CancelledByRole = 3
	CancelledBySystem      CancelledByRole = 4
)

type Appointment struct {
	ID         int64
	ClientID   int32
	CreatedBy  *int32
	MasterID   int32
	ServiceID  int32
	LocationID int32

	StartTime time.Time
	EndTime   time.Time
	Timezone  string

	Status        AppointmentStatus
	Price         int32
	Currency      string
	PaymentStatus PaymentStatus

	IsOnline   bool
	MeetingURL string

	ClientNotes string
	MasterNotes string

	CancelledAt     *time.Time
	CancelledBy     *int32
	CancelledByRole *CancelledByRole
	CancelReason    string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type AppointmentCard struct {
	Appointment  *Appointment
	MasterSlug   string
	MasterName   string
	ServiceName  string
	ServiceSlug  string
	LocationName string
	Address      string
}

type AppointmentCursor struct {
	StartTime time.Time
	ID        int64
}

type BookingRepository interface {
	Create(ctx context.Context, a *Appointment) (*Appointment, error)
	GetByID(ctx context.Context, id int64) (*AppointmentCard, error)
	ListByClient(ctx context.Context, clientID int32, status *AppointmentStatus, cursor *AppointmentCursor, limit int32) ([]*AppointmentCard, error)
	ListByMaster(ctx context.Context, masterID int32, locationID *int32, status *AppointmentStatus, from, to *time.Time, cursor *AppointmentCursor, limit int32) ([]*AppointmentCard, error)
	Cancel(ctx context.Context, id int64, status AppointmentStatus, byRole CancelledByRole, byUserID *int32, reason string) (*Appointment, error)
	Reschedule(ctx context.Context, id int64, newStart, newEnd time.Time, locationID int32) (*Appointment, error)
	Confirm(ctx context.Context, id int64) (*Appointment, error)
	Complete(ctx context.Context, id int64, masterNotes string) (*Appointment, error)
	MarkNoShow(ctx context.Context, id int64) (*Appointment, error)
	UpdatePaymentStatus(ctx context.Context, id int64, status PaymentStatus) error
}
