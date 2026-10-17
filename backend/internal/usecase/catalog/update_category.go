package catalog

import (
	"context"
	"strings"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type UpdateCategoryInput struct {
	ID   int32
	Name string
}

type UpdateCategoryOutput struct {
	Category *domain.Category
}

type UpdateCategoryUseCase struct {
	deps Deps
}

func NewUpdateCategoryUseCase(deps Deps) *UpdateCategoryUseCase {
	return &UpdateCategoryUseCase{deps: deps}
}

func (uc *UpdateCategoryUseCase) Execute(ctx context.Context, in UpdateCategoryInput) (*UpdateCategoryOutput, error) {
	if !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return nil, domain.ErrForbidden
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, domain.ErrInvalidInput
	}

	c, err := uc.deps.Catalog.GetCategory(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	c.Name = name

	updated, err := uc.deps.Catalog.UpdateCategory(ctx, c)
	if err != nil {
		return nil, err
	}

	invalidateCategories(ctx, uc.deps.CategoriesCache)
	invalidateServices(ctx, uc.deps.ServicesCache, nil)

	return &UpdateCategoryOutput{Category: updated}, nil
}
