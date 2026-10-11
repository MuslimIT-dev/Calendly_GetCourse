package schedule

import (
	"context"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type ListExceptionsInput struct {
	LocationID int32
	From       time.Time
	To         time.Time
}

type ListExceptionsOutput struct {
	Exceptions []*domain.Exception
}

type ListExceptionsUseCase struct {
	deps Deps
}

func NewListExceptionsUseCase(deps Deps) *ListExceptionsUseCase {
	return &ListExceptionsUseCase{deps: deps}
}

func (uc *ListExceptionsUseCase) Execute(ctx context.Context, in ListExceptionsInput) (*ListExceptionsOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	master, err := uc.deps.Masters.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	belongs, err := uc.deps.Locations.BelongsToMaster(ctx, in.LocationID, master.ID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, domain.ErrForbidden
	}

	list, err := uc.deps.Schedule.ListExceptions(ctx, in.LocationID, in.From, in.To)
	if err != nil {
		return nil, err
	}
	return &ListExceptionsOutput{Exceptions: list}, nil
}