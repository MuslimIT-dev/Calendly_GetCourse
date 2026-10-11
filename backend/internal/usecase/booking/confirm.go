package booking

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type ConfirmInput struct {
	ID int64
}

type ConfirmOutput struct {
	Appointment *domain.AppointmentCard
}

type ConfirmUseCase struct {
	deps   Deps
	events EventPublisher
}

func NewConfirmUseCase(deps Deps, events EventPublisher) *ConfirmUseCase {
	return &ConfirmUseCase{deps: deps, events: events}
}

func (uc *ConfirmUseCase) Execute(ctx context.Context, in ConfirmInput) (*ConfirmOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	card, err := uc.deps.Bookings.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	a := card.Appointment

	if a.MasterID != userID && !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return nil, domain.ErrForbidden
	}
	if a.Status != domain.AppointmentPending {
		return nil, domain.ErrInvalidStatus
	}

	updated, err := uc.deps.Bookings.Confirm(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	_ = publishEvent(ctx, uc.events, AppointmentConfirmedEvent{
		AppointmentID: updated.ID,
		ClientID:      updated.ClientID,
		MasterID:      updated.MasterID,
		StartTime:     updated.StartTime,
	})

	newCard, err := uc.deps.Bookings.GetByID(ctx, updated.ID)
	if err != nil {
		return nil, err
	}

	return &ConfirmOutput{Appointment: newCard}, nil
}
