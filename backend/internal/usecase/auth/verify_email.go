package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type VerifyEmailInput struct {
	Token string
}

type VerifyEmailOutput struct {
	UserID int32
}

type VerifyEmailUseCase struct {
	deps Deps
}

func NewVerifyEmailUseCase(deps Deps) *VerifyEmailUseCase {
	return &VerifyEmailUseCase{deps: deps}
}

func (uc *VerifyEmailUseCase) Execute(ctx context.Context, in VerifyEmailInput) (*VerifyEmailOutput, error) {
	if in.Token == "" {
		return nil, domain.ErrInvalidToken
	}

	tokenHash := sha256Hex(in.Token)
	key := "email_verify:" + tokenHash

	value, err := uc.deps.VerifyTokens.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("token lookup: %w", err)
	}
	if value == nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := uc.deps.Users.GetUserByID(ctx, value.UserID)
	if err != nil {
		return nil, err
	}

	if user.EmailVerified {
		_ = uc.deps.VerifyTokens.Delete(ctx, key)
		return &VerifyEmailOutput{UserID: user.ID}, nil
	}

	if err := uc.deps.Users.SetEmailVerified(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("set email verified: %w", err)
	}

	_ = uc.deps.VerifyTokens.Delete(ctx, key)

	return &VerifyEmailOutput{UserID: user.ID}, nil
}