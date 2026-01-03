package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

// CreateTestRedisClient creates a test Redis client.
// The client will connect to Redis, verify connectivity, and flush the dedicated test DB.
// In a real test environment, you might use testcontainers or a mock.
func CreateTestRedisClient(t *testing.T) *redis.Client {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   15, // Use dedicated test DB instead of default DB 0
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to connect to Redis at %s: %v", addr, err)
	}

	// prepare empty data
	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("Failed to flush Redis database during cleanup: %v", err)
	}
	return client
}
