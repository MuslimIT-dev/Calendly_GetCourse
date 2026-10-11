package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	notificationuc "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/usecase/notification"
)

type AppointmentEvent struct {
	AppointmentID int64     `json:"appointment_id"`
	ClientID      int32     `json:"client_id"`
	MasterID      int32     `json:"master_id"`
	ServiceID     int32     `json:"service_id"`
	LocationID    int32     `json:"location_id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	CompletedAt   time.Time `json:"completed_at"`
	CancelledBy   int32     `json:"cancelled_by"`
	ByRole        int32     `json:"by_role"`
	Reason        string    `json:"reason"`
	OldStartTime  time.Time `json:"old_start_time"`
	NewStartTime  time.Time `json:"new_start_time"`
}

type masterOwnerLookup interface {
	GetUserIDByMasterID(ctx context.Context, masterID int32) (int32, error)
}

type AppointmentsConsumer struct {
	sendUC  *notificationuc.SendUseCase
	masters masterOwnerLookup
}

func NewAppointmentsConsumer(sendUC *notificationuc.SendUseCase, masters masterOwnerLookup) *AppointmentsConsumer {
	return &AppointmentsConsumer{sendUC: sendUC, masters: masters}
}

func (c *AppointmentsConsumer) Consume(ctx context.Context, event *AppointmentEvent) error {
	if event == nil || event.AppointmentID <= 0 || event.ClientID <= 0 || event.MasterID <= 0 {
		return fmt.Errorf("invalid appointment event")
	}
	link := fmt.Sprintf("/appointments/%d", event.AppointmentID)
	switch {
	case !event.OldStartTime.IsZero() && !event.NewStartTime.IsZero():
		return c.handleRescheduled(ctx, event, link)
	case event.ByRole != 0:
		return c.handleCancelled(ctx, event, link)
	case !event.CompletedAt.IsZero():
		return c.handleCompleted(ctx, event, link)
	case event.ServiceID != 0 || !event.EndTime.IsZero():
		return c.handleCreated(ctx, event, link)
	case !event.StartTime.IsZero():
		return c.handleConfirmed(ctx, event, link)
	default:
		return fmt.Errorf("unrecognized appointment event for appointment %d", event.AppointmentID)
	}
}

func (c *AppointmentsConsumer) handleCreated(ctx context.Context, event *AppointmentEvent, link string) error {
	masterUserID, err := c.masters.GetUserIDByMasterID(ctx, event.MasterID)
	if err != nil {
		return fmt.Errorf("resolve appointment master user: %w", err)
	}
	when := event.StartTime.Format("02 Jan 15:04")
	if err := c.send(ctx, masterUserID, domain.NotifAppointmentCreated, "Новая запись",
		fmt.Sprintf("Клиент записался на %s", when), link, map[string]any{"appointment_id": event.AppointmentID}); err != nil {
		return err
	}
	return c.send(ctx, event.ClientID, domain.NotifAppointmentCreated, "Запись создана",
		fmt.Sprintf("Ожидайте подтверждения. Время: %s", when), link, map[string]any{"appointment_id": event.AppointmentID})
}

func (c *AppointmentsConsumer) handleConfirmed(ctx context.Context, event *AppointmentEvent, link string) error {
	return c.send(ctx, event.ClientID, domain.NotifAppointmentConfirmed, "Запись подтверждена",
		fmt.Sprintf("Время встречи: %s", event.StartTime.Format("02 Jan 15:04")), link,
		map[string]any{"appointment_id": event.AppointmentID})
}

func (c *AppointmentsConsumer) handleCancelled(ctx context.Context, event *AppointmentEvent, link string) error {
	targetID := event.ClientID
	if event.CancelledBy == event.ClientID {
		var err error
		targetID, err = c.masters.GetUserIDByMasterID(ctx, event.MasterID)
		if err != nil {
			return fmt.Errorf("resolve appointment master user: %w", err)
		}
	}
	return c.send(ctx, targetID, domain.NotifAppointmentCancelled, "Запись отменена",
		fmt.Sprintf("Встреча %s отменена. Причина: %s", event.StartTime.Format("02 Jan 15:04"), event.Reason),
		link, map[string]any{"appointment_id": event.AppointmentID, "reason": event.Reason})
}

func (c *AppointmentsConsumer) handleRescheduled(ctx context.Context, event *AppointmentEvent, link string) error {
	oldWhen, newWhen := event.OldStartTime.Format("02 Jan 15:04"), event.NewStartTime.Format("02 Jan 15:04")
	body := fmt.Sprintf("Было %s, стало %s", oldWhen, newWhen)
	if err := c.send(ctx, event.ClientID, domain.NotifAppointmentRescheduled, "Встреча перенесена", body, link, nil); err != nil {
		return err
	}
	masterUserID, err := c.masters.GetUserIDByMasterID(ctx, event.MasterID)
	if err != nil {
		return fmt.Errorf("resolve appointment master user: %w", err)
	}
	return c.send(ctx, masterUserID, domain.NotifAppointmentRescheduled, "Запись перенесена", body, link, nil)
}

func (c *AppointmentsConsumer) handleCompleted(ctx context.Context, event *AppointmentEvent, link string) error {
	return c.send(ctx, event.ClientID, domain.NotifSystem, "Встреча завершена",
		"Ваша встреча завершена. Вы можете оставить отзыв.", link,
		map[string]any{"appointment_id": event.AppointmentID})
}

func (c *AppointmentsConsumer) send(ctx context.Context, userID int32, notificationType domain.NotificationType,
	title, body, link string, metadata map[string]any) error {
	if userID <= 0 {
		return fmt.Errorf("invalid notification recipient")
	}
	_, err := c.sendUC.Execute(ctx, notificationuc.SendInput{
		UserID: userID, Type: notificationType, Channel: domain.ChannelInApp,
		Title: title, Body: body, Link: link, Metadata: metadata,
	})
	return err
}
