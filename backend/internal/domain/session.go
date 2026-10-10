package domain

import (
	"context"
	"time"
)

type Session struct {
	ID          string
	UserID      int32
	RefreshHash string
	IPAddress   string
	UserAgent   string
	ExpiresAt   time.Time
}

type SessionValue struct {
	UserID    int32     `json:"user_id"`
	Roles     []int32   `json:"roles"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionRepository interface {
	Create(ctx context.Context, s *Session) error
	GetByHash(ctx context.Context, hash string) (*Session, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllByUser(ctx context.Context, userID int32) error
}
