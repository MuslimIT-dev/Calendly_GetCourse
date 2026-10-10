package auth

import (
	"context"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type ForgotPasswordInput struct {
	Email string
}

type ForgotPasswordOutput struct{}

type PasswordResetValue struct {
	UserID    int32     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type PasswordResetRequestedEvent struct {
	UserID       int32  `json:"user_id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	ResetToken   string `json:"reset_token"`
}

type ForgotPasswordUseCase struct {
	deps Deps
}

func NewForgotPasswordUseCase(deps Deps) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{deps: deps}
}

func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, in ForgotPasswordInput) (*ForgotPasswordOutput, error) {
	user, err := uc.deps.Users.GetUserByEmail(ctx, in.Email)
	if err != nil {
		return &ForgotPasswordOutput{}, nil
	}

	rawToken := generateToken(32)
	tokenHash := sha256Hex(rawToken)

	value := PasswordResetValue{
		UserID:    user.ID,
		CreatedAt: time.Now().UTC(),
	}

	key := "password_reset:" + tokenHash
	if err := uc.deps.PasswordResets.Set(ctx, key, &value, uc.deps.PasswordResetTTL); err != nil {
		return nil, err
	}

	_ = uc.deps.ResetEvents.Publish(ctx, "user.password_reset_requested", &PasswordResetRequestedEvent{
		UserID:     user.ID,
		Email:      user.Email,
		Name:       user.Name,
		ResetToken: rawToken,
	})

	return &ForgotPasswordOutput{}, nil
}