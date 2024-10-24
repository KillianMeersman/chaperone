package kvstore

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisKVStore(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("redis is not running")
		return
	}

	store := NewRedisKVStore(client)

	v, exists, err := store.Get(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fail()
	}
	if v != "" {
		t.Fail()
	}

	err = store.Store(context.Background(), "test", "test2", 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	v, exists, err = store.Get(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fail()
	}
	if v != "test2" {
		t.Fail()
	}

	err = store.Store(context.Background(), "test", "test3", 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	v, exists, err = store.Get(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fail()
	}
	if v != "test3" {
		t.Fail()
	}

	err = store.Store(context.Background(), "test", "test3", 1*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Millisecond)

	v, exists, err = store.Get(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fail()
	}
	if v != "" {
		t.Fail()
	}
}
