package kvstore

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisKVStore[K string, V string] struct {
	client *redis.Client
}

func NewRedisKVStore[K string, V string](client *redis.Client) *RedisKVStore[K, V] {
	return &RedisKVStore[K, V]{
		client: client,
	}
}

func (s *RedisKVStore[K, V]) Store(ctx context.Context, key string, value V, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

func (s *RedisKVStore[K, V]) Get(ctx context.Context, key string) (string, bool, error) {
	value, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	return value, true, err
}
