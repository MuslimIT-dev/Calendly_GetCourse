package user

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type GetUserInput struct {
	UserID int32
}

type GetUserOutput struct {
	User *domain.User
}

type GetUserUseCase struct {
	deps Deps
}

func NewGetUserUseCase(deps Deps) *GetUserUseCase {
	return &GetUserUseCase{deps: deps}
}

func (uc *GetUserUseCase) Execute(ctx context.Context, in GetUserInput) (*GetUserOutput, error) {
	key := fmt.Sprintf("user:%d", in.UserID)

	if cached, err := uc.deps.Cache.Get(ctx, key); err == nil && cached != nil {
		return &GetUserOutput{User: fromCache(cached)}, nil
	}

	user, err := uc.deps.Users.GetUserByID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	_ = uc.deps.Cache.Set(ctx, key, toCache(user), uc.deps.CacheTTL)

	return &GetUserOutput{User: user}, nil
}

func toCache(u *domain.User) *domain.CachedUser {
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

func fromCache(c *domain.CachedUser) *domain.User {
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