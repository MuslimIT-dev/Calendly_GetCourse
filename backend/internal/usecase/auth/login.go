package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type LoginInput struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

type LoginOutput struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type LoginUseCase struct {
	deps Deps
}

func NewLoginUseCase(deps Deps) *LoginUseCase {
	return &LoginUseCase{deps: deps}
}

func (uc *LoginUseCase) Execute(ctx context.Context, in LoginInput) (*LoginOutput, error) {
	user, err := uc.deps.Users.GetUserByEmail(ctx, in.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := uc.deps.Hasher.Verify(user.PasswordHash, in.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	access, refresh, expiresIn, err := uc.deps.Tokens.GeneratePair(user.ID, user.Roles)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	session := SessionValue{
		UserID:    user.ID,
		Roles:     toInt32Slice(user.Roles),
		IPAddress: in.IPAddress,
		UserAgent: in.UserAgent,
		CreatedAt: time.Now().UTC(),
	}
	key := fmt.Sprintf("session:%d:%s", user.ID, sha256Hex(refresh))
	if err := uc.deps.Sessions.Set(ctx, key, &session, uc.deps.SessionTTL); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	return &LoginOutput{
		User:         user,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    expiresIn,
	}, nil
}