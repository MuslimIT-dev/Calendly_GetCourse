package catalog

import (
	"context"
	"strings"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type CreateCategoryInput struct {
	Name string
}

type CreateCategoryOutput struct {
	Category *domain.Category
}

type CreateCategoryUseCase struct {
	deps Deps
}

func NewCreateCategoryUseCase(deps Deps) *CreateCategoryUseCase {
	return &CreateCategoryUseCase{deps: deps}
}

func (uc *CreateCategoryUseCase) Execute(ctx context.Context, in CreateCategoryInput) (*CreateCategoryOutput, error) {
	if !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return nil, domain.ErrForbidden
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, domain.ErrInvalidInput
	}

	c, err := uc.deps.Catalog.CreateCategory(ctx, &domain.Category{
		Name:     name,
		IsActive: true,
	})
	if err != nil {
		return nil, err
	}

	invalidateCategories(ctx, uc.deps.CategoriesCache)
	invalidateServices(ctx, uc.deps.ServicesCache, nil)

	return &CreateCategoryOutput{Category: c}, nil
}
