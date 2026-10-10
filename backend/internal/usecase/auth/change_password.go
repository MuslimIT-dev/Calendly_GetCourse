package auth

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/pkg/password"
)

type ChangePasswordInput struct {
	CurrentPassword string
	NewPassword     string
}

type ChangePasswordOutput struct{}

type ChangePasswordUseCase struct {
	deps Deps
}

func NewChangePasswordUseCase(deps Deps) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{deps: deps}
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, in ChangePasswordInput) (*ChangePasswordOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	if password.IsWeak(in.NewPassword) {
		return nil, domain.ErrWeakPassword
	}

	if uc.deps.Breach.IsPwned(ctx, in.NewPassword) {
		return nil, domain.ErrPwnedPassword
	}

	user, err := uc.deps.Users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := uc.deps.Hasher.Verify(user.PasswordHash, in.CurrentPassword); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	hash, err := uc.deps.Hasher.Hash(in.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	if err := uc.deps.Users.UpdatePassword(ctx, user.ID, hash); err != nil {
		return nil, fmt.Errorf("update password: %w", err)
	}

	_, _ = uc.deps.Sessions.DeleteByPattern(ctx, fmt.Sprintf("session:%d:*", user.ID))

	return &ChangePasswordOutput{}, nil
}
