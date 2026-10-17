package worker

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
	notificationuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/notification"
)

type UserRegisteredEvent struct {
	UserID int32  `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

type UsersConsumer struct{ sendUC *notificationuc.SendUseCase }

func NewUsersConsumer(sendUC *notificationuc.SendUseCase) *UsersConsumer {
	return &UsersConsumer{sendUC: sendUC}
}

func (c *UsersConsumer) Consume(ctx context.Context, event *UserRegisteredEvent) error {
	if event == nil || event.UserID <= 0 {
		return fmt.Errorf("invalid user registered event")
	}
	_, err := c.sendUC.Execute(ctx, notificationuc.SendInput{
		UserID: event.UserID, Type: domain.NotifSystem, Channel: domain.ChannelInApp,
		Title: "Добро пожаловать!", Body: "Спасибо за регистрацию на платформе", Link: "/profile",
	})
	return err
}
