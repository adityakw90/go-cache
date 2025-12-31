package testutil

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
)

// CreateTestRedisClient creates a test Redis client.
// In a real test environment, you might use testcontainers or a mock.
func CreateTestRedisClient(t *testing.T) *redis.Client {
	// For now, return a client that may not be connected
	// In integration tests, this would connect to a test Redis instance
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	// prepare empty data
	ctx := context.Background()
	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("Failed to flush Redis database during cleanup: %v", err)
	}
	return client
}
