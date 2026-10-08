package cache

import (
	"context"
	"encoding/json"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type RedisCache[T any] struct {
	client *redis.Client
}

func NewRedisCache[T any](client *redis.Client) *RedisCache[T] {
	return &RedisCache[T]{client: client}
}

func (r *RedisCache[T]) Get(ctx context.Context, key string) (*T, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var result T
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *RedisCache[T]) Set(ctx context.Context, key string, value *T, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}