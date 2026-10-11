package catalog

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type GetCategoryInput struct {
	ID int32
}

type GetCategoryOutput struct {
	Category *domain.Category
}

type GetCategoryUseCase struct {
	deps Deps
}

func NewGetCategoryUseCase(deps Deps) *GetCategoryUseCase {
	return &GetCategoryUseCase{deps: deps}
}

func (uc *GetCategoryUseCase) Execute(ctx context.Context, in GetCategoryInput) (*GetCategoryOutput, error) {
	c, err := uc.deps.Catalog.GetCategory(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	return &GetCategoryOutput{Category: c}, nil
}
