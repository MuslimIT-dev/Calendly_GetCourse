package catalog

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type ListCategoriesInput struct{}

type ListCategoriesOutput struct {
	Categories []*domain.Category
}

type ListCategoriesUseCase struct {
	deps Deps
}

func NewListCategoriesUseCase(deps Deps) *ListCategoriesUseCase {
	return &ListCategoriesUseCase{deps: deps}
}

func (uc *ListCategoriesUseCase) Execute(ctx context.Context, _ ListCategoriesInput) (*ListCategoriesOutput, error) {
	if cached, err := uc.deps.CategoriesCache.Get(ctx, "catalog:categories"); err == nil && cached != nil {
		return &ListCategoriesOutput{Categories: cached.Categories}, nil
	}

	list, err := uc.deps.Catalog.ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	_ = uc.deps.CategoriesCache.Set(ctx, "catalog:categories", &domain.CachedCategories{
		Categories: list,
	}, uc.deps.CacheTTL)

	return &ListCategoriesOutput{Categories: list}, nil
}
