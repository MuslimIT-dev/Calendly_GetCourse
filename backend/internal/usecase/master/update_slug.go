package master

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{1,62}[a-z0-9])?$`)

type UpdateSlugInput struct {
	Slug string
}

type UpdateSlugOutput struct {
	Profile *domain.MasterProfile
}

type UpdateSlugUseCase struct {
	deps Deps
}

func NewUpdateSlugUseCase(deps Deps) *UpdateSlugUseCase {
	return &UpdateSlugUseCase{deps: deps}
}

func (uc *UpdateSlugUseCase) Execute(ctx context.Context, in UpdateSlugInput) (*UpdateSlugOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	slug := strings.ToLower(strings.TrimSpace(in.Slug))
	if !slugRegex.MatchString(slug) {
		return nil, domain.ErrInvalidInput
	}

	profile, err := uc.deps.Masters.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if profile.Slug == slug {
		return &UpdateSlugOutput{Profile: profile}, nil
	}

	exists, err := uc.deps.Masters.SlugExists(ctx, slug, profile.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrSlugTaken
	}

	if err := uc.deps.Masters.UpdateSlug(ctx, profile.ID, slug); err != nil {
		return nil, fmt.Errorf("update slug: %w", err)
	}

	invalidateMasterCache(ctx, uc.deps.Cache, profile.Slug, userID)
	profile.Slug = slug

	return &UpdateSlugOutput{Profile: profile}, nil
}