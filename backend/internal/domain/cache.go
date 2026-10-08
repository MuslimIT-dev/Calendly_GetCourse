package domain

import (
	"context"
	"time"
)

type Cache[T any] interface {
	Get(ctx context.Context, key string) (*T, error)
	Set(ctx context.Context, key string, value *T, ttl time.Duration) error
}