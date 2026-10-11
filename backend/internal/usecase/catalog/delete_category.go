package catalog

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type DeleteCategoryInput struct {
	ID int32
}

type DeleteCategoryOutput struct{}

type DeleteCategoryUseCase struct {
	deps Deps
}

func NewDeleteCategoryUseCase(deps Deps) *DeleteCategoryUseCase {
	return &DeleteCategoryUseCase{deps: deps}
}

func (uc *DeleteCategoryUseCase) Execute(ctx context.Context, in DeleteCategoryInput) (*DeleteCategoryOutput, error) {
	if !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return nil, domain.ErrForbidden
	}

	has, err := uc.deps.Catalog.CategoryHasServices(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, domain.ErrCategoryHasServices
	}

	if err := uc.deps.Catalog.DeleteCategory(ctx, in.ID); err != nil {
		return nil, err
	}

	invalidateCategories(ctx, uc.deps.CategoriesCache)
	invalidateServices(ctx, uc.deps.ServicesCache, nil)

	return &DeleteCategoryOutput{}, nil
}
