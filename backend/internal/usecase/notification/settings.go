package notification

import (
	"context"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type GetSettingsInput struct{}
type GetSettingsOutput struct{ Settings *domain.NotificationSettings }
type GetSettingsUseCase struct{ deps Deps }

func NewGetSettingsUseCase(deps Deps) *GetSettingsUseCase { return &GetSettingsUseCase{deps: deps} }

func (uc *GetSettingsUseCase) Execute(ctx context.Context, _ GetSettingsInput) (*GetSettingsOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	settings, err := uc.deps.Notifications.GetSettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &GetSettingsOutput{Settings: settings}, nil
}

type UpdateSettingsInput struct {
	EmailEnabled    bool
	TelegramEnabled bool
	SMSEnabled      bool
	PushEnabled     bool
	MutedTypes      []int32
	Reminder24h     bool
	Reminder2h      bool
	Reminder30m     bool
}
type UpdateSettingsOutput struct{ Settings *domain.NotificationSettings }
type UpdateSettingsUseCase struct{ deps Deps }

func NewUpdateSettingsUseCase(deps Deps) *UpdateSettingsUseCase {
	return &UpdateSettingsUseCase{deps: deps}
}

func (uc *UpdateSettingsUseCase) Execute(ctx context.Context, in UpdateSettingsInput) (*UpdateSettingsOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	for _, notificationType := range in.MutedTypes {
		if notificationType < int32(domain.NotifAppointmentCreated) || notificationType > int32(domain.NotifSystem) {
			return nil, domain.ErrInvalidInput
		}
	}
	settings, err := uc.deps.Notifications.UpsertSettings(ctx, &domain.NotificationSettings{
		UserID: userID, EmailEnabled: in.EmailEnabled, TelegramEnabled: in.TelegramEnabled,
		SMSEnabled: in.SMSEnabled, PushEnabled: in.PushEnabled, MutedTypes: in.MutedTypes,
		Reminder24h: in.Reminder24h, Reminder2h: in.Reminder2h, Reminder30m: in.Reminder30m,
	})
	if err != nil {
		return nil, err
	}
	return &UpdateSettingsOutput{Settings: settings}, nil
}
