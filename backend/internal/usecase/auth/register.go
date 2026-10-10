package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type RegisterInput struct {
	Name     string
	Email    string
	Password string
	Role     domain.Role
}

type RegisterOutput struct {
	User         *domain.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type RegisterUseCase struct {
	deps Deps
}

func NewRegisterUseCase(deps Deps) *RegisterUseCase {
	return &RegisterUseCase{deps: deps}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, in RegisterInput) (*RegisterOutput, error) {
	if in.Role != domain.RoleUser && in.Role != domain.RoleMaster {
		return nil, fmt.Errorf("invalid role for registration")
	}

	roles := []domain.Role{domain.RoleUser}
	if in.Role == domain.RoleMaster {
		roles = []domain.Role{domain.RoleMaster, domain.RoleUser}
	}

	hash, err := uc.deps.Hasher.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := uc.deps.Users.Create(ctx, &domain.User{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: hash,
	}, roles)
	if err != nil {
		return nil, err
	}

	access, refresh, expiresIn, err := uc.deps.Tokens.GeneratePair(user.ID, roles)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	session := SessionValue{
		UserID:    user.ID,
		Roles:     toInt32Slice(roles),
		CreatedAt: time.Now().UTC(),
	}

	key := fmt.Sprintf("session:%d:%s", user.ID, sha256Hex(refresh))
	if err := uc.deps.Sessions.Set(ctx, key, &session, uc.deps.SessionTTL); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	rawVerifyToken := generateToken(32)
	verifyHash := sha256Hex(rawVerifyToken)
	verifyValue := VerifyEmailValue{
		UserID:    user.ID,
		CreatedAt: time.Now().UTC(),
	}
	if err := uc.deps.VerifyTokens.Set(ctx, "email_verify:"+verifyHash, &verifyValue, uc.deps.VerifyTokenTTL); err != nil {
		return nil, fmt.Errorf("save verify token: %w", err)
	}

	_ = uc.deps.Events.Publish(ctx, "user.registered", &UserRegisteredEvent{
		UserID:      user.ID,
		Email:       user.Email,
		Name:        user.Name,
		VerifyToken: rawVerifyToken,
	})

	return &RegisterOutput{
		User:         user,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    expiresIn,
	}, nil
}