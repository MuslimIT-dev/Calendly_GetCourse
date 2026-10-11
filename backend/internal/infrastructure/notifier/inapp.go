package notifier

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	notificationuc "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/usecase/notification"
)

type InAppNotifier struct{}

func NewInAppNotifier() *InAppNotifier { return &InAppNotifier{} }

func (n *InAppNotifier) Channel() domain.NotificationChannel { return domain.ChannelInApp }

func (n *InAppNotifier) Send(_ context.Context, _ *domain.Notification, _ notificationuc.Recipient) error {
	return nil
}
