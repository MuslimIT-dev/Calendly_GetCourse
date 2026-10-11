package catalog

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type GetDefaultServiceInput struct {
	ID int32
}

type GetDefaultServiceOutput struct {
	Service *domain.DefaultService
}

type GetDefaultServiceUseCase struct {
	deps Deps
}

func NewGetDefaultServiceUseCase(deps Deps) *GetDefaultServiceUseCase {
	return &GetDefaultServiceUseCase{deps: deps}
}

func (uc *GetDefaultServiceUseCase) Execute(ctx context.Context, in GetDefaultServiceInput) (*GetDefaultServiceOutput, error) {
	s, err := uc.deps.Catalog.GetDefaultService(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	return &GetDefaultServiceOutput{Service: s}, nil
}
