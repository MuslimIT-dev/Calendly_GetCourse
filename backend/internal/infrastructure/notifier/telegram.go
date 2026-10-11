package notifier

import (
	"context"
	"errors"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	notificationuc "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/usecase/notification"
)

type TelegramNotifier struct{ botToken string }

func NewTelegramNotifier(botToken string) *TelegramNotifier {
	return &TelegramNotifier{botToken: botToken}
}

func (n *TelegramNotifier) Channel() domain.NotificationChannel { return domain.ChannelTelegram }

func (n *TelegramNotifier) Send(_ context.Context, _ *domain.Notification, recipient notificationuc.Recipient) error {
	if n.botToken == "" || recipient.Telegram == "" {
		return errors.New("telegram notifier requires bot token and recipient chat ID")
	}
	return errors.New("telegram delivery provider is not configured")
}
