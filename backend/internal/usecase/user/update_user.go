package user

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type UpdateUserInput struct {
	Name      *string
	AvatarURL *string
	Timezone  *string
}

type UpdateUserOutput struct {
	User *domain.User
}

type UpdateUserUseCase struct {
	deps Deps
}

func NewUpdateUserUseCase(deps Deps) *UpdateUserUseCase {
	return &UpdateUserUseCase{deps: deps}
}

func (uc *UpdateUserUseCase) Execute(ctx context.Context, in UpdateUserInput) (*UpdateUserOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	user, err := uc.deps.Users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if in.Name != nil {
		if *in.Name == "" {
			return nil, domain.ErrInvalidInput
		}
		user.Name = *in.Name
	}
	if in.AvatarURL != nil {
		user.AvatarURL = *in.AvatarURL
	}
	if in.Timezone != nil {
		user.Timezone = *in.Timezone
	}

	updated, err := uc.deps.Users.UpdateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	_ = uc.deps.Cache.Delete(ctx, fmt.Sprintf("user:%d", updated.ID))

	return &UpdateUserOutput{User: updated}, nil
}