package notifier

import (
	"context"
	"errors"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
	notificationuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/notification"
)

type EmailNotifier struct{ from string }

func NewEmailNotifier(from string) *EmailNotifier { return &EmailNotifier{from: from} }

func (n *EmailNotifier) Channel() domain.NotificationChannel { return domain.ChannelEmail }

func (n *EmailNotifier) Send(_ context.Context, _ *domain.Notification, recipient notificationuc.Recipient) error {
	if n.from == "" || recipient.Email == "" {
		return errors.New("email notifier requires sender and recipient addresses")
	}
	return errors.New("email delivery provider is not configured")
}
