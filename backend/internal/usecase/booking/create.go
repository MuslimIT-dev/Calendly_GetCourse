package booking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type CreateInput struct {
	MasterID    int32
	ServiceID   int32
	LocationID  int32
	StartTime   time.Time
	Timezone    string
	ClientNotes string
}

type CreateOutput struct {
	Appointment *domain.AppointmentCard
}

type CreateUseCase struct {
	deps   Deps
	events EventPublisher
}

func NewCreateUseCase(deps Deps, events EventPublisher) *CreateUseCase {
	return &CreateUseCase{deps: deps, events: events}
}

func (uc *CreateUseCase) Execute(ctx context.Context, in CreateInput) (*CreateOutput, error) {
	clientID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	if in.MasterID == 0 || in.ServiceID == 0 || in.LocationID == 0 || in.StartTime.IsZero() {
		return nil, domain.ErrInvalidInput
	}
	if in.Timezone == "" {
		in.Timezone = "UTC"
	}

	svc, err := uc.deps.Services.GetByID(ctx, in.ServiceID)
	if err != nil {
		return nil, err
	}
	if svc.MasterID != in.MasterID || !svc.IsActive {
		return nil, domain.ErrServiceNotFound
	}

	loc, err := uc.deps.Locations.GetByID(ctx, in.LocationID)
	if err != nil {
		return nil, err
	}
	if loc.MasterID != in.MasterID || !loc.IsActive {
		return nil, domain.ErrServiceNotForLocation
	}

	price := svc.Price
	duration := svc.DurationMins

	for _, sl := range svc.Locations {
		if sl.LocationID == in.LocationID && sl.IsActive {
			if sl.Price > 0 {
				price = sl.Price
			}
			if sl.DurationMins > 0 {
				duration = sl.DurationMins
			}
			break
		}
	}

	endTime := in.StartTime.Add(time.Duration(duration) * time.Minute)

	appt, err := uc.deps.Bookings.Create(ctx, &domain.Appointment{
		ClientID:    clientID,
		CreatedBy:   &clientID,
		MasterID:    in.MasterID,
		ServiceID:   in.ServiceID,
		LocationID:  in.LocationID,
		StartTime:   in.StartTime,
		EndTime:     endTime,
		Timezone:    in.Timezone,
		Price:       price,
		Currency:    svc.Currency,
		IsOnline:    loc.IsOnline,
		MeetingURL:  loc.MeetingURL,
		ClientNotes: in.ClientNotes,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrSlotTaken
		}
		return nil, fmt.Errorf("create appointment: %w", err)
	}

	card, err := uc.deps.Bookings.GetByID(ctx, appt.ID)
	if err != nil {
		return nil, err
	}

	_ = publishEvent(ctx, uc.events, AppointmentCreatedEvent{
		AppointmentID: appt.ID,
		ClientID:      appt.ClientID,
		MasterID:      appt.MasterID,
		ServiceID:     appt.ServiceID,
		LocationID:    appt.LocationID,
		StartTime:     appt.StartTime,
		EndTime:       appt.EndTime,
		Price:         appt.Price,
		Currency:      appt.Currency,
		IsOnline:      appt.IsOnline,
		Timezone:      appt.Timezone,
	})

	return &CreateOutput{Appointment: card}, nil
}
