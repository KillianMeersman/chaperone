package kvstore

import (
	"context"
	"sync"
	"time"
)

type MemoryKVStore[K comparable, V any] struct {
	values   map[K]V
	janitors map[K]context.CancelFunc
	lock     *sync.RWMutex
	ctx      context.Context
}

func NewMemoryKVStore[K comparable, V any](ctx context.Context) *MemoryKVStore[K, V] {
	return &MemoryKVStore[K, V]{
		values:   make(map[K]V),
		janitors: make(map[K]context.CancelFunc),
		lock:     &sync.RWMutex{},
		ctx:      ctx,
	}
}

func (s *MemoryKVStore[K, V]) Store(ctx context.Context, key K, value V, ttl time.Duration) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Cancel old janitor if it exists.
	if cancelJanitor, ok := s.janitors[key]; ok {
		cancelJanitor()
	}

	s.values[key] = value

	// Set up janitor if ttl > 0
	if ttl > 0 {
		ctx, cancel := context.WithDeadline(s.ctx, time.Now().Add(ttl))
		s.janitors[key] = cancel
		go func() {
			<-ctx.Done()
			switch ctx.Err() {
			// Only delete the current entry if the context ended due to timeout,
			// otherwise we will delete new values when they cancel our context
			// (and we acquire the lock AFTER their value is set).
			case context.DeadlineExceeded:
				s.lock.Lock()
				defer s.lock.Unlock()
				delete(s.values, key)
			default:
				return
			}
		}()
	}

	return nil
}

func (s *MemoryKVStore[K, V]) Get(ctx context.Context, key K) (V, bool, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	x, ok := s.values[key]
	return x, ok, nil
}
