package auth

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type ResetPasswordInput struct {
	Token       string
	NewPassword string
}

type ResetPasswordOutput struct{}

type PasswordChangedEvent struct {
	UserID int32 `json:"user_id"`
}

type ResetPasswordUseCase struct {
	deps Deps
}

func NewResetPasswordUseCase(deps Deps) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{deps: deps}
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, in ResetPasswordInput) (*ResetPasswordOutput, error) {
	if in.Token == "" || len(in.NewPassword) < 8 {
		return nil, domain.ErrInvalidToken
	}

	tokenHash := sha256Hex(in.Token)
	key := "password_reset:" + tokenHash

	value, err := uc.deps.PasswordResets.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("token lookup: %w", err)
	}
	if value == nil {
		return nil, domain.ErrInvalidToken
	}

	hash, err := uc.deps.Hasher.Hash(in.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	if err := uc.deps.Users.UpdatePassword(ctx, value.UserID, hash); err != nil {
		return nil, fmt.Errorf("update password: %w", err)
	}

	_ = uc.deps.PasswordResets.Delete(ctx, key)

	_, _ = uc.deps.Sessions.DeleteByPattern(ctx, fmt.Sprintf("session:*:user:%d", value.UserID))

	_ = uc.deps.ResetEvents.Publish(ctx, "user.password_changed", &PasswordChangedEvent{
		UserID: value.UserID,
	})

	return &ResetPasswordOutput{}, nil
}