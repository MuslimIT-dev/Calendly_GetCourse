package booking

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type NoShowInput struct {
	ID int64
}

type NoShowOutput struct {
	Appointment *domain.AppointmentCard
}

type NoShowUseCase struct {
	deps Deps
}

func NewNoShowUseCase(deps Deps) *NoShowUseCase {
	return &NoShowUseCase{deps: deps}
}

func (uc *NoShowUseCase) Execute(ctx context.Context, in NoShowInput) (*NoShowOutput, error) {
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

	updated, err := uc.deps.Bookings.MarkNoShow(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	newCard, err := uc.deps.Bookings.GetByID(ctx, updated.ID)
	if err != nil {
		return nil, err
	}

	return &NoShowOutput{Appointment: newCard}, nil
}
