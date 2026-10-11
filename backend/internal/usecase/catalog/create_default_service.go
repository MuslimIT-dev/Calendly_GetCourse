package catalog

import (
	"context"
	"strings"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type CreateDefaultServiceInput struct {
	CategoryID   int32
	Slug         string
	Name         string
	Description  string
	DurationMins int32
	Price        int32
	Currency     string
}

type CreateDefaultServiceOutput struct {
	Service *domain.DefaultService
}

type CreateDefaultServiceUseCase struct {
	deps Deps
}

func NewCreateDefaultServiceUseCase(deps Deps) *CreateDefaultServiceUseCase {
	return &CreateDefaultServiceUseCase{deps: deps}
}

func (uc *CreateDefaultServiceUseCase) Execute(ctx context.Context, in CreateDefaultServiceInput) (*CreateDefaultServiceOutput, error) {
	if !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return nil, domain.ErrForbidden
	}

	slug := strings.ToLower(strings.TrimSpace(in.Slug))
	name := strings.TrimSpace(in.Name)
	if slug == "" || name == "" || in.DurationMins <= 0 || in.Price < 0 || in.Currency == "" {
		return nil, domain.ErrInvalidInput
	}

	exists, err := uc.deps.Catalog.DefaultServiceSlugExists(ctx, slug, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrSlugTaken
	}

	s, err := uc.deps.Catalog.CreateDefaultService(ctx, &domain.DefaultService{
		CategoryID:   in.CategoryID,
		Slug:         slug,
		Name:         name,
		Description:  in.Description,
		Price:        in.Price,
		Currency:     in.Currency,
		DurationMins: in.DurationMins,
		IsActive:     true,
	})
	if err != nil {
		return nil, err
	}

	invalidateServices(ctx, uc.deps.ServicesCache, &in.CategoryID)

	return &CreateDefaultServiceOutput{Service: s}, nil
}
