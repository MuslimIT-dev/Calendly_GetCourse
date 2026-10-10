package master

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type GetMasterInput struct {
	Slug string
}

type GetMasterOutput struct {
	User    *domain.User
	Profile *domain.MasterProfile
}

type GetMasterUseCase struct {
	deps Deps
}

func NewGetMasterUseCase(deps Deps) *GetMasterUseCase {
	return &GetMasterUseCase{deps: deps}
}

func (uc *GetMasterUseCase) Execute(ctx context.Context, in GetMasterInput) (*GetMasterOutput, error) {
	if in.Slug == "" {
		return nil, domain.ErrInvalidInput
	}

	key := "master:slug:" + in.Slug

	if cached, err := uc.deps.Cache.Get(ctx, key); err == nil && cached != nil {
		return &GetMasterOutput{
			User:    fromCacheUser(cached.User),
			Profile: cached.Profile,
		}, nil
	}

	profile, user, err := uc.deps.Masters.GetBySlug(ctx, in.Slug)
	if err != nil {
		return nil, err
	}

	_ = uc.deps.Cache.Set(ctx, key, &domain.CachedMaster{
		Profile: profile,
		User:    toCacheUser(user),
	}, uc.deps.CacheTTL)

	return &GetMasterOutput{User: user, Profile: profile}, nil
}

func toCacheUser(u *domain.User) *domain.CachedUser {
	return &domain.CachedUser{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		AvatarURL:     u.AvatarURL,
		Timezone:      u.Timezone,
		EmailVerified: u.EmailVerified,
		Roles:         u.Roles,
	}
}

func fromCacheUser(c *domain.CachedUser) *domain.User {
	if c == nil {
		return nil
	}
	return &domain.User{
		ID:            c.ID,
		Name:          c.Name,
		Email:         c.Email,
		AvatarURL:     c.AvatarURL,
		Timezone:      c.Timezone,
		EmailVerified: c.EmailVerified,
		Roles:         c.Roles,
	}
}

func invalidateMasterCache(ctx context.Context, cache domain.Cache[domain.CachedMaster], slug string, userID int32) {
	if slug != "" {
		_ = cache.Delete(ctx, "master:slug:"+slug)
	}
	_ = cache.Delete(ctx, fmt.Sprintf("master:user:%d", userID))
}