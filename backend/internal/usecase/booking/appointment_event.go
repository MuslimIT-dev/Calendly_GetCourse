package booking

import "time"

type AppointmentCreatedEvent struct {
	AppointmentID int64     `json:"appointment_id"`
	ClientID      int32     `json:"client_id"`
	MasterID      int32     `json:"master_id"`
	ServiceID     int32     `json:"service_id"`
	LocationID    int32     `json:"location_id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Price         int32     `json:"price"`
	Currency      string    `json:"currency"`
	IsOnline      bool      `json:"is_online"`
	Timezone      string    `json:"timezone"`
}

type AppointmentCancelledEvent struct {
	AppointmentID int64     `json:"appointment_id"`
	ClientID      int32     `json:"client_id"`
	MasterID      int32     `json:"master_id"`
	CancelledBy   int32     `json:"cancelled_by"`
	ByRole        int32     `json:"by_role"`
	Reason        string    `json:"reason"`
	StartTime     time.Time `json:"start_time"`
}

type AppointmentConfirmedEvent struct {
	AppointmentID int64     `json:"appointment_id"`
	ClientID      int32     `json:"client_id"`
	MasterID      int32     `json:"master_id"`
	StartTime     time.Time `json:"start_time"`
}

type AppointmentCompletedEvent struct {
	AppointmentID int64     `json:"appointment_id"`
	ClientID      int32     `json:"client_id"`
	MasterID      int32     `json:"master_id"`
	CompletedAt   time.Time `json:"completed_at"`
}

type AppointmentRescheduledEvent struct {
	AppointmentID int64     `json:"appointment_id"`
	ClientID      int32     `json:"client_id"`
	MasterID      int32     `json:"master_id"`
	OldStartTime  time.Time `json:"old_start_time"`
	NewStartTime  time.Time `json:"new_start_time"`
}
