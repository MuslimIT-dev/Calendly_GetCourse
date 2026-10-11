package catalog

import (
	"context"
	"strings"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type UpdateDefaultServiceInput struct {
	ID           int32
	Slug         string
	Name         string
	Description  string
	DurationMins int32
	Price        int32
}

type UpdateDefaultServiceOutput struct {
	Service *domain.DefaultService
}

type UpdateDefaultServiceUseCase struct {
	deps Deps
}

func NewUpdateDefaultServiceUseCase(deps Deps) *UpdateDefaultServiceUseCase {
	return &UpdateDefaultServiceUseCase{deps: deps}
}

func (uc *UpdateDefaultServiceUseCase) Execute(ctx context.Context, in UpdateDefaultServiceInput) (*UpdateDefaultServiceOutput, error) {
	if !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return nil, domain.ErrForbidden
	}

	s, err := uc.deps.Catalog.GetDefaultService(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if in.Slug != "" {
		slug := strings.ToLower(strings.TrimSpace(in.Slug))
		if slug != s.Slug {
			exists, err := uc.deps.Catalog.DefaultServiceSlugExists(ctx, slug, s.ID)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, domain.ErrSlugTaken
			}
			s.Slug = slug
		}
	}
	if in.Name != "" {
		s.Name = strings.TrimSpace(in.Name)
	}
	if in.Description != "" {
		s.Description = in.Description
	}
	if in.DurationMins > 0 {
		s.DurationMins = in.DurationMins
	}
	if in.Price > 0 {
		s.Price = in.Price
	}

	updated, err := uc.deps.Catalog.UpdateDefaultService(ctx, s)
	if err != nil {
		return nil, err
	}

	invalidateServices(ctx, uc.deps.ServicesCache, &updated.CategoryID)

	return &UpdateDefaultServiceOutput{Service: updated}, nil
}
