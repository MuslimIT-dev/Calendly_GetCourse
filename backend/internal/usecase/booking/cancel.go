package booking

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type CancelInput struct {
	ID     int64
	Reason string
}

type CancelOutput struct {
	Appointment *domain.AppointmentCard
}

type CancelUseCase struct {
	deps   Deps
	events EventPublisher
}

func NewCancelUseCase(deps Deps, events EventPublisher) *CancelUseCase {
	return &CancelUseCase{deps: deps, events: events}
}

func (uc *CancelUseCase) Execute(ctx context.Context, in CancelInput) (*CancelOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	card, err := uc.deps.Bookings.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	a := card.Appointment

	if a.Status != domain.AppointmentPending && a.Status != domain.AppointmentConfirmed {
		return nil, domain.ErrInvalidStatus
	}

	var status domain.AppointmentStatus
	var byRole domain.CancelledByRole

	switch {
	case a.ClientID == userID:
		status = domain.AppointmentCancelledByClient
		byRole = domain.CancelledByClient
	case isMasterOwner(ctx, uc.deps.Masters, userID, a.MasterID):
		status = domain.AppointmentCancelledByMaster
		byRole = domain.CancelledByMaster
	case appcontext.HasRole(ctx, int32(domain.RoleAdmin)):
		status = domain.AppointmentCancelledByMaster
		byRole = domain.CancelledByAdmin
	default:
		return nil, domain.ErrForbidden
	}

	updated, err := uc.deps.Bookings.Cancel(ctx, in.ID, status, byRole, &userID, in.Reason)
	if err != nil {
		return nil, fmt.Errorf("cancel: %w", err)
	}

	_ = publishEvent(ctx, uc.events, AppointmentCancelledEvent{
		AppointmentID: updated.ID,
		ClientID:      updated.ClientID,
		MasterID:      updated.MasterID,
		CancelledBy:   userID,
		ByRole:        int32(byRole),
		Reason:        in.Reason,
		StartTime:     updated.StartTime,
	})

	newCard, err := uc.deps.Bookings.GetByID(ctx, updated.ID)
	if err != nil {
		return nil, err
	}

	return &CancelOutput{Appointment: newCard}, nil
}
