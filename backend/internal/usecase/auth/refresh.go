package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type RefreshInput struct {
	RefreshToken string
	IPAddress    string
	UserAgent    string
}

type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type RefreshUseCase struct {
	deps Deps
}

func NewRefreshUseCase(deps Deps) *RefreshUseCase {
	return &RefreshUseCase{deps: deps}
}

func (uc *RefreshUseCase) Execute(ctx context.Context, in RefreshInput) (*RefreshOutput, error) {
	userID, _, err := uc.deps.Tokens.VerifyRefresh(in.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	oldHash := sha256Hex(in.RefreshToken)
	oldKey := "session:" + oldHash

	session, err := uc.deps.Sessions.Get(ctx, oldKey)
	if err != nil {
		return nil, fmt.Errorf("session lookup: %w", err)
	}
	if session == nil {
		_ = uc.deps.Sessions.DeleteByPattern(ctx, "session:*")
		return nil, domain.ErrInvalidToken
	}

	if err := uc.deps.Sessions.Delete(ctx, oldKey); err != nil {
		return nil, fmt.Errorf("revoke old session: %w", err)
	}

	user, err := uc.deps.Users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	access, refresh, expiresIn, err := uc.deps.Tokens.GeneratePair(user.ID, user.Roles)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	newHash := sha256Hex(refresh)
	newSession := SessionValue{
		UserID:    user.ID,
		Roles:     toInt32Slice(user.Roles),
		IPAddress: in.IPAddress,
		UserAgent: in.UserAgent,
		CreatedAt: time.Now().UTC(),
	}
	if err := uc.deps.Sessions.Set(ctx, "session:"+newHash, &newSession, uc.deps.SessionTTL); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	return &RefreshOutput{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    expiresIn,
	}, nil
}

func hashRefresh(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}