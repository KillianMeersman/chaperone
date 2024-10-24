package kvstore

import (
	"context"
	"time"
)

type KVStore[K comparable, V any] interface {
	// Store the value under the provided key for ttl. If ttl is <= 0, store forever.
	Store(ctx context.Context, key K, value V, ttl time.Duration) error
	Get(ctx context.Context, key K) (V, bool, error)
}
