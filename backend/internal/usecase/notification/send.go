package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type SendInput struct {
	UserID   int32
	Type     domain.NotificationType
	Channel  domain.NotificationChannel
	Title    string
	Body     string
	Link     string
	Metadata map[string]any
}

type SendOutput struct {
	Notification *domain.Notification
}

type SendUseCase struct {
	deps     Deps
	registry *NotifierRegistry
}

func NewSendUseCase(deps Deps, registry *NotifierRegistry) *SendUseCase {
	return &SendUseCase{deps: deps, registry: registry}
}

func (uc *SendUseCase) Execute(ctx context.Context, in SendInput) (*SendOutput, error) {
	if in.UserID <= 0 || in.Type < domain.NotifAppointmentCreated || in.Type > domain.NotifSystem ||
		strings.TrimSpace(in.Title) == "" {
		return nil, domain.ErrInvalidInput
	}
	if in.Channel == 0 {
		in.Channel = domain.ChannelInApp
	}
	if in.Channel < domain.ChannelInApp || in.Channel > domain.ChannelPush {
		return nil, domain.ErrInvalidInput
	}

	settings, err := uc.deps.Notifications.GetSettings(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("get notification settings: %w", err)
	}
	if settings == nil {
		return nil, fmt.Errorf("get notification settings: repository returned nil settings")
	}
	if !channelEnabled(settings, in.Channel) || typeMuted(settings, in.Type) {
		return &SendOutput{}, nil
	}

	metadata, err := json.Marshal(in.Metadata)
	if err != nil {
		return nil, fmt.Errorf("encode notification metadata: %w", err)
	}
	notification, err := uc.deps.Notifications.Create(ctx, &domain.Notification{
		UserID: in.UserID, Type: in.Type, Channel: in.Channel, Status: domain.NotifStatusPending,
		Title: strings.TrimSpace(in.Title), Body: in.Body, Link: in.Link, Metadata: metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	if in.Channel != domain.ChannelInApp {
		notifier, ok := uc.registry.Get(in.Channel)
		if !ok {
			return nil, uc.failNotification(ctx, notification.ID, fmt.Errorf("no notifier configured for channel %d", in.Channel))
		}
		user, err := uc.deps.Users.GetUserByID(ctx, in.UserID)
		if err != nil {
			return nil, uc.failNotification(ctx, notification.ID, fmt.Errorf("get notification recipient: %w", err))
		}
		if user == nil {
			return nil, uc.failNotification(ctx, notification.ID, fmt.Errorf("get notification recipient: repository returned nil user"))
		}
		if err := notifier.Send(ctx, notification, Recipient{UserID: user.ID, Email: user.Email}); err != nil {
			return nil, uc.failNotification(ctx, notification.ID, fmt.Errorf("send notification: %w", err))
		}
	}

	if err := uc.deps.Notifications.MarkSent(ctx, notification.ID); err != nil {
		return nil, fmt.Errorf("mark notification sent: %w", err)
	}
	notification.Status = domain.NotifStatusSent
	now := time.Now()
	notification.SentAt = &now
	return &SendOutput{Notification: notification}, nil
}

func (uc *SendUseCase) failNotification(ctx context.Context, id int64, cause error) error {
	if err := uc.deps.Notifications.MarkFailed(ctx, id); err != nil {
		return errors.Join(cause, fmt.Errorf("mark notification failed: %w", err))
	}
	return cause
}

func channelEnabled(settings *domain.NotificationSettings, channel domain.NotificationChannel) bool {
	switch channel {
	case domain.ChannelEmail:
		return settings.EmailEnabled
	case domain.ChannelTelegram:
		return settings.TelegramEnabled
	case domain.ChannelSMS:
		return settings.SMSEnabled
	case domain.ChannelPush:
		return settings.PushEnabled
	default:
		return true
	}
}

func typeMuted(settings *domain.NotificationSettings, notificationType domain.NotificationType) bool {
	for _, mutedType := range settings.MutedTypes {
		if mutedType == int32(notificationType) {
			return true
		}
	}
	return false
}
