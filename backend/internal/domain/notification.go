package domain

import (
	"context"
	"time"
)

type NotificationChannel int32

const (
	ChannelInApp NotificationChannel = iota + 1
	ChannelEmail
	ChannelTelegram
	ChannelSMS
	ChannelPush
)

type NotificationType int32

const (
	NotifAppointmentCreated NotificationType = iota + 1
	NotifAppointmentConfirmed
	NotifAppointmentCancelled
	NotifAppointmentReminder
	NotifAppointmentRescheduled
	NotifPaymentSucceeded
	NotifPaymentFailed
	NotifReviewCreated
	NotifReviewReplied
	NotifCourseEnrolled
	NotifSystem
)

type NotificationStatus int32

const (
	NotifStatusPending NotificationStatus = iota + 1
	NotifStatusSent
	NotifStatusFailed
	NotifStatusRead
)

type Notification struct {
	ID        int64
	UserID    int32
	Type      NotificationType
	Channel   NotificationChannel
	Status    NotificationStatus
	Title     string
	Body      string
	Link      string
	Metadata  []byte
	SentAt    *time.Time
	ReadAt    *time.Time
	CreatedAt time.Time
}

type NotificationSettings struct {
	UserID          int32
	EmailEnabled    bool
	TelegramEnabled bool
	SMSEnabled      bool
	PushEnabled     bool
	MutedTypes      []int32
	Reminder24h     bool
	Reminder2h      bool
	Reminder30m     bool
	UpdatedAt       time.Time
}

type NotificationCursor struct {
	CreatedAt time.Time
	ID        int64
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) (*Notification, error)
	GetByID(ctx context.Context, id int64) (*Notification, error)
	ListByUser(ctx context.Context, userID int32, status *NotificationStatus, cursor *NotificationCursor, limit int32) ([]*Notification, error)
	MarkRead(ctx context.Context, id int64, userID int32) (int64, error)
	MarkAllRead(ctx context.Context, userID int32) (int64, error)
	Delete(ctx context.Context, id int64, userID int32) (int64, error)
	CountUnread(ctx context.Context, userID int32) (int32, error)
	ListPending(ctx context.Context, limit int32) ([]*Notification, error)
	MarkSent(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
	GetSettings(ctx context.Context, userID int32) (*NotificationSettings, error)
	UpsertSettings(ctx context.Context, settings *NotificationSettings) (*NotificationSettings, error)
}
