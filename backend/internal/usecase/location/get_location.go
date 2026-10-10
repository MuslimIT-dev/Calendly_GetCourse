package location

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type GetLocationInput struct {
	ID int32
}

type GetLocationOutput struct {
	Location *domain.Location
}

type GetLocationUseCase struct {
	deps Deps
}

func NewGetLocationUseCase(deps Deps) *GetLocationUseCase {
	return &GetLocationUseCase{deps: deps}
}

func (uc *GetLocationUseCase) Execute(ctx context.Context, in GetLocationInput) (*GetLocationOutput, error) {
	if in.ID == 0 {
		return nil, domain.ErrInvalidInput
	}

	loc, err := uc.deps.Locations.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	return &GetLocationOutput{Location: loc}, nil
}