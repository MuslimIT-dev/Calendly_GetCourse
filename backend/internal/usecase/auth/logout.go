package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type LogoutInput struct {
	RefreshToken string
}

type LogoutOutput struct{}

type LogoutUseCase struct {
	deps Deps
}

func NewLogoutUseCase(deps Deps) *LogoutUseCase {
	return &LogoutUseCase{deps: deps}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, in LogoutInput) (*LogoutOutput, error) {
	if in.RefreshToken == "" {
		return &LogoutOutput{}, nil
	}

	userID, _, err := uc.deps.Tokens.VerifyRefresh(in.RefreshToken)
	if err != nil {
		return &LogoutOutput{}, nil
	}

	hash := sha256.Sum256([]byte(in.RefreshToken))
	key := fmt.Sprintf("session:%d:%s", userID, hex.EncodeToString(hash[:]))

	_ = uc.deps.Sessions.Delete(ctx, key)

	return &LogoutOutput{}, nil
}
