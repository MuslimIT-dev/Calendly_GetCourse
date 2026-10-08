package domain

import (
	"context"
	"errors"
)

type Session struct {
    ID           string
    UserID       int32
    RefreshHash  string
    IPAddress    string
    UserAgent    string
    ExpiresAt    time.Time
}

type SessionRepository interface {
    Create(ctx context.Context, s *Session) error
    GetByHash(ctx context.Context, hash string) (*Session, error)
    Revoke(ctx context.Context, id string) error
    RevokeAllByUser(ctx context.Context, userID int32) error
}