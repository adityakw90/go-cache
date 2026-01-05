package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// CreateTestRedisClient creates a test Redis client.
// The client will connect to Redis, verify connectivity, and flush the dedicated test DB.
// In a real test environment, you might use testcontainers or a mock.
func CreateTestRedisClient(
	t *testing.T,
	opts ...RedisTestOption,
) *redis.Client {
	t.Helper()

	options := &redisTestOptions{
		db: 15, // default test DB
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.db < 0 || options.db > 15 {
		t.Fatalf("invalid redis db number: %d", options.db)
	}

	lockFile, err := lockRedisDB(options.db)
	if err != nil {
		t.Fatalf("failed to acquire redis db %d lock: %v", options.db, err)
	}

	t.Cleanup(func() {
		unlockRedisDB(lockFile)
	})

	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   options.db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("failed to connect to redis db %d: %v", options.db, err)
	}

	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("failed to flush redis db %d: %v", options.db, err)
	}

	t.Cleanup(func() {
		_ = client.Close()
	})

	return client
}
