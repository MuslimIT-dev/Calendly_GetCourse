package location

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type DeleteLocationInput struct {
	ID int32
}

type DeleteLocationOutput struct{}

type DeleteLocationUseCase struct {
	deps Deps
}

func NewDeleteLocationUseCase(deps Deps) *DeleteLocationUseCase {
	return &DeleteLocationUseCase{deps: deps}
}

func (uc *DeleteLocationUseCase) Execute(ctx context.Context, in DeleteLocationInput) (*DeleteLocationOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	master, err := uc.deps.Masters.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	loc, err := uc.deps.Locations.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if loc.MasterID != master.ID {
		return nil, domain.ErrForbidden
	}

	if err := uc.deps.Locations.Delete(ctx, in.ID); err != nil {
		return nil, fmt.Errorf("delete location: %w", err)
	}

	invalidate(ctx, uc.deps.Cache, master.ID)

	return &DeleteLocationOutput{}, nil
}