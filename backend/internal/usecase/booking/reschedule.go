package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type RescheduleInput struct {
	ID           int64
	LocationID   *int32
	NewStartTime time.Time
	Timezone     string
}

type RescheduleOutput struct {
	Appointment *domain.AppointmentCard
}

type RescheduleUseCase struct {
	deps   Deps
	events EventPublisher
}

func NewRescheduleUseCase(deps Deps, events EventPublisher) *RescheduleUseCase {
	return &RescheduleUseCase{deps: deps, events: events}
}

func (uc *RescheduleUseCase) Execute(ctx context.Context, in RescheduleInput) (*RescheduleOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	card, err := uc.deps.Bookings.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	a := card.Appointment

	if !isMasterOrClient(ctx, uc.deps.Masters, userID, a.MasterID, a.ClientID) {
		return nil, domain.ErrForbidden
	}

	if a.Status != domain.AppointmentPending && a.Status != domain.AppointmentConfirmed {
		return nil, domain.ErrInvalidStatus
	}

	svc, err := uc.deps.Services.GetByID(ctx, a.ServiceID)
	if err != nil {
		return nil, err
	}
	duration := svc.DurationMins
	for _, sl := range svc.Locations {
		if sl.LocationID == a.LocationID && sl.DurationMins > 0 {
			duration = sl.DurationMins
			break
		}
	}

	locationID := a.LocationID
	if in.LocationID != nil {
		locationID = *in.LocationID
	}

	newEnd := in.NewStartTime.Add(time.Duration(duration) * time.Minute)

	updated, err := uc.deps.Bookings.Reschedule(ctx, in.ID, in.NewStartTime, newEnd, locationID)
	if err != nil {
		return nil, fmt.Errorf("reschedule: %w", err)
	}

	_ = publishEvent(ctx, uc.events, AppointmentRescheduledEvent{
		AppointmentID: updated.ID,
		ClientID:      updated.ClientID,
		MasterID:      updated.MasterID,
		OldStartTime:  a.StartTime,
		NewStartTime:  in.NewStartTime,
	})

	newCard, err := uc.deps.Bookings.GetByID(ctx, updated.ID)
	if err != nil {
		return nil, err
	}

	return &RescheduleOutput{Appointment: newCard}, nil
}
