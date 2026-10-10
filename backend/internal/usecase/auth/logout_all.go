package auth

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
)

type LogoutAllInput struct{}

type LogoutAllOutput struct {
	RevokedCount int32
}

type LogoutAllUseCase struct {
	deps Deps
}

func NewLogoutAllUseCase(deps Deps) *LogoutAllUseCase {
	return &LogoutAllUseCase{deps: deps}
}

func (uc *LogoutAllUseCase) Execute(ctx context.Context, _ LogoutAllInput) (*LogoutAllOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}

	pattern := fmt.Sprintf("session:%d:*", userID)
	revokedCount, err := uc.deps.Sessions.DeleteByPattern(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("delete sessions: %w", err)
	}

	return &LogoutAllOutput{RevokedCount: int32(revokedCount)}, nil
}