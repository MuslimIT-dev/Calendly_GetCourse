package notification

import (
	"context"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type Deps struct {
	Notifications domain.NotificationRepository
	Users         domain.UserRepository
}

type Notifier interface {
	Channel() domain.NotificationChannel
	Send(ctx context.Context, notification *domain.Notification, recipient Recipient) error
}

type Recipient struct {
	UserID   int32
	Email    string
	Telegram string
}

type NotifierRegistry struct {
	notifiers map[domain.NotificationChannel]Notifier
}

func NewNotifierRegistry(notifiers ...Notifier) *NotifierRegistry {
	registry := &NotifierRegistry{notifiers: make(map[domain.NotificationChannel]Notifier, len(notifiers))}
	for _, notifier := range notifiers {
		registry.notifiers[notifier.Channel()] = notifier
	}
	return registry
}

func (r *NotifierRegistry) Get(channel domain.NotificationChannel) (Notifier, bool) {
	if r == nil {
		return nil, false
	}
	notifier, ok := r.notifiers[channel]
	return notifier, ok
}
