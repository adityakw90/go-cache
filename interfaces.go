package cache

import (
	"context"
	"time"

	"github.com/adityakw90/go-cache/internal/key"
	"github.com/redis/go-redis/v9"
)

// Cache provides a Redis-based caching layer with support for
// distributed locking, versioning, and custom key generation.
// It wraps Redis operations with automatic serialization/deserialization
// and provides thread-safe cache access through distributed locks.
type Cache interface {
	// Get retrieves a cached value by key and deserializes it into resultType.
	// Returns ErrGetCacheMiss if the key does not exist.
	// Returns an error if Redis operation or deserialization fails.
	Get(ctx context.Context, key string, resultType interface{}) error

	// Set stores a value in cache with the specified TTL.
	// The value will be serialized before storing.
	// Returns an error if serialization or Redis operation fails.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// CleanCache invalidates cache entries for a given key.
	// It increments the cache version for all registered prefixes and namespaces,
	// effectively invalidating all cached entries that match the key.
	// Returns an error if lock acquisition or Redis operation fails.
	CleanCache(ctx context.Context, key string, params map[string]interface{}, session redis.Pipeliner, execute bool, listLockedKey *[]string) error

	// Cached returns a decorator function that wraps the given function with caching logic.
	// The decorator handles cache key generation, versioning, and distributed locking
	// to ensure thread-safe cache operations.
	// Parameters:
	//   - keyName: the base key name for cache identification
	//   - ttl: can be a time.Duration, a function(result, args) time.Duration, or use default
	//   - versioning: if true, enables cache versioning for automatic invalidation
	//   - prefix: optional key prefix, uses default if empty
	// Returned decorator parameters:
	//   - customKeyFunc: optional custom key generator
	//   - useHashKey: if true, uses a hashed version of the key
	// Returns a decorator function that wraps the target function.
	Cached(keyName string, ttl interface{}, versioning bool, prefix string) func(fn func(ctx context.Context, args ...interface{}) (interface{}, error), customKeyFunc CustomKeyFunction, useHashKey bool) func(resultType interface{}, ctx context.Context, args ...interface{}) (interface{}, error)

	// GetSession returns a Redis pipeline session.
	GetSession() redis.Pipeliner
}

// re-export custom key function type from key package.
type CustomKeyFunction = key.CustomKeyFunction

// NewCustomKeyFunction creates a new custom key function with the given name,
// callable function, and parameter names.
// Parameters:
//   - name: unique identifier for the custom key function
//   - callable: function that generates a cache key from arguments
//   - params: names of the parameters expected by the callable function
//
// Returns the custom key function and an error if validation fails.
func NewCustomKeyFunction(
	name string,
	callable func(args ...interface{}) string,
	params []string,
) (CustomKeyFunction, error) {
	return key.NewCustomKeyFunction(name, callable, params)
}
