package booking

import (
	"context"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type CompleteInput struct {
	ID          int64
	MasterNotes string
}

type CompleteOutput struct {
	Appointment *domain.AppointmentCard
}

type CompleteUseCase struct {
	deps   Deps
	events EventPublisher
}

func NewCompleteUseCase(deps Deps, events EventPublisher) *CompleteUseCase {
	return &CompleteUseCase{deps: deps, events: events}
}

func (uc *CompleteUseCase) Execute(ctx context.Context, in CompleteInput) (*CompleteOutput, error) {
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
	if a.Status != domain.AppointmentPending && a.Status != domain.AppointmentConfirmed {
		return nil, domain.ErrInvalidStatus
	}

	updated, err := uc.deps.Bookings.Complete(ctx, in.ID, in.MasterNotes)
	if err != nil {
		return nil, err
	}

	_ = publishEvent(ctx, uc.events, AppointmentCompletedEvent{
		AppointmentID: updated.ID,
		ClientID:      updated.ClientID,
		MasterID:      updated.MasterID,
		CompletedAt:   time.Now().UTC(),
	})

	newCard, err := uc.deps.Bookings.GetByID(ctx, updated.ID)
	if err != nil {
		return nil, err
	}

	return &CompleteOutput{Appointment: newCard}, nil
}
