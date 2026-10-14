package booking

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type GetInput struct {
	ID int64
}

type GetOutput struct {
	Appointment *domain.AppointmentCard
}

type GetUseCase struct {
	deps Deps
}

func NewGetUseCase(deps Deps) *GetUseCase {
	return &GetUseCase{deps: deps}
}

func (uc *GetUseCase) Execute(ctx context.Context, in GetInput) (*GetOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	card, err := uc.deps.Bookings.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if !isMasterOrClient(ctx, uc.deps.Masters, userID, card.Appointment.MasterID, card.Appointment.ClientID) {
		return nil, domain.ErrForbidden
	}

	return &GetOutput{Appointment: card}, nil
}
