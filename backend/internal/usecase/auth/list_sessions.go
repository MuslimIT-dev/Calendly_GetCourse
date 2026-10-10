package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type SessionInfo struct {
	ID        string
	IPAddress string
	UserAgent string
	CreatedAt time.Time
	IsCurrent bool
}

type ListSessionsInput struct {
	CurrentRefreshToken string
}

type ListSessionsOutput struct {
	Sessions []SessionInfo
}

type ListSessionsUseCase struct {
	deps Deps
}

func NewListSessionsUseCase(deps Deps) *ListSessionsUseCase {
	return &ListSessionsUseCase{deps: deps}
}

func (uc *ListSessionsUseCase) Execute(ctx context.Context, in ListSessionsInput) (*ListSessionsOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	pattern := fmt.Sprintf("session:%d:*", userID)
	raw, err := uc.deps.Sessions.ScanValues(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("scan sessions: %w", err)
	}

	currentHash := ""
	if in.CurrentRefreshToken != "" {
		currentHash = sha256Hex(in.CurrentRefreshToken)
	}

	out := make([]SessionInfo, 0, len(raw))
	for key, val := range raw {
		id := strings.TrimPrefix(key, fmt.Sprintf("session:%d:", userID))
		out = append(out, SessionInfo{
			ID:        id,
			IPAddress: val.IPAddress,
			UserAgent: val.UserAgent,
			CreatedAt: val.CreatedAt,
			IsCurrent: id == currentHash,
		})
	}

	return &ListSessionsOutput{Sessions: out}, nil
}
