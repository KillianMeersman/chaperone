package kvstore

import (
	"context"
	"testing"
	"time"
)

func TestStoreGet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewMemoryKVStore[string, string](ctx)

	store.Store(ctx, "test", ":)", -1)
	val, exists, err := store.Get(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !exists || val != ":)" {
		t.Fatal()
	}

	store.Store(ctx, "test", ":)", -1)
	val, exists, err = store.Get(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !exists || val != ":)" {
		t.Fatal()
	}
}

func TestTTL(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := NewMemoryKVStore[string, string](ctx)
	store.Store(ctx, "test", ":)", 10*time.Millisecond)

	time.Sleep(time.Millisecond)
	val, exists, err := store.Get(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !exists || val != ":)" {
		t.Fatal()
	}

	time.Sleep(15 * time.Millisecond)
	_, exists, err = store.Get(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal()
	}
}
