package cache

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/errs"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/adityakw90/go-cache/internal/version"
)

// Cache is the internal cache implementation.
type Cache struct {
	RedisClient         *redis.Client
	Tracer              adapter.Tracer
	Logger              adapter.Logger
	Semaphore           adapter.Semaphore
	KeyPrefix           string
	KeyGenerator        key.KeyGeneratorFunc
	KeyVersionGenerator key.KeyGeneratorFunc
	VersionGenerator    key.KeyGeneratorFunc
	VersionExpire       time.Duration
	LockGenerator       key.KeyGeneratorFunc
	LockDuration        time.Duration
	LockInterval        time.Duration
	ExpireDefault       time.Duration
	keyUsage            map[string][]string                         // To track cache key usage: keyUsage[keyName] = []prefix
	customKeys          map[string]map[string]key.CustomKeyFunction // To track custom key functions
	keyMutex            sync.Mutex                                  // Mutex to handle concurrent map access
}

// Options holds all cache configuration.
type Options struct {
	KeyPrefix           string
	ExpireDefault       time.Duration
	VersionExpire       time.Duration
	LockDuration        time.Duration
	LockInterval        time.Duration
	SemaphoreSize       int
	Tracer              adapter.Tracer
	Logger              adapter.Logger
	Semaphore           adapter.Semaphore
	KeyGenerator        key.KeyGeneratorFunc
	KeyVersionGenerator key.KeyGeneratorFunc
	VersionGenerator    key.KeyGeneratorFunc
	LockGenerator       key.KeyGeneratorFunc
}

// NewCache creates a new cache instance with options.
func NewCache(redisClient *redis.Client, opts Options) (*Cache, error) {
	if redisClient == nil {
		return nil, errs.NewInvalidConfigError("redisClient", "cannot be nil")
	}

	// Use provided semaphore or create default
	semaphore := opts.Semaphore
	if semaphore == nil {
		semaphore = adapter.NewSemaphore(opts.SemaphoreSize)
	}

	// Use provided tracer or default to no-op
	tracer := opts.Tracer
	if tracer == nil {
		tracer = adapter.NewNoOpTracer()
	}

	// Use provided logger or default to no-op
	logger := opts.Logger
	if logger == nil {
		logger = adapter.NewNoOpLogger()
	}

	c := &Cache{
		RedisClient:         redisClient,
		Tracer:              tracer,
		Logger:              logger,
		Semaphore:           semaphore,
		KeyPrefix:           opts.KeyPrefix,
		KeyGenerator:        opts.KeyGenerator,
		KeyVersionGenerator: opts.KeyVersionGenerator,
		VersionGenerator:    opts.VersionGenerator,
		VersionExpire:       opts.VersionExpire,
		LockGenerator:       opts.LockGenerator,
		LockDuration:        opts.LockDuration,
		LockInterval:        opts.LockInterval,
		ExpireDefault:       opts.ExpireDefault,
		keyUsage:            make(map[string][]string),
		customKeys:          make(map[string]map[string]key.CustomKeyFunction),
	}

	return c, nil
}

// CleanCache invalidates cache entries for a given key by incrementing their versions.
// It supports both standard keys and custom key functions, handles multiple prefixes,
// and manages distributed locks during invalidation.
//
// Parameters:
//   - ctx: Context for tracing and cancellation
//   - key: The cache key name to invalidate
//   - params: Parameters map for custom key functions (can be nil for standard keys)
//   - session: Redis pipeline session (creates new one if nil)
//   - execute: If true, executes the pipeline and acquires/releases locks
//   - listLockedKey: Pointer to slice that collects lock keys (initialized if nil)
//
// Behavior:
//   - Retrieves all registered prefixes for the key using GetCacheKeyUsage
//   - If custom keys are registered, generates namespaces by calling Call(params) on each
//   - For standard keys (no custom keys), uses the key name as namespace
//   - Generates lock keys for each prefix/namespace combination
//   - Increments cache versions in the pipeline for all combinations
//   - If execute=true, acquires multiple locks, executes pipeline, then releases locks
//
// Use cases:
//   - Invalidate cache when data is updated (e.g., user profile changed)
//   - Batch invalidation across multiple prefixes
//   - Invalidation with custom key functions (e.g., user-specific cache)
//
// Example:
//
//	// Standard key invalidation
//	err := cache.CleanCache(ctx, "getUser", nil, nil, true, nil)
//
//	// Custom key invalidation with params
//	params := map[string]interface{}{"uid": "user-123"}
//	session := cache.GetSession()
//	var lockKeys []string
//	err := cache.CleanCache(ctx, "getUser", params, session, false, &lockKeys)
//	// ... do other operations ...
//	session.Exec(ctx) // Execute all operations together
//
// Returns an error if:
//   - Custom key function fails to generate namespace (missing params)
//   - Lock key generation fails
//   - Cache version increment fails
//   - Lock acquisition fails (when execute=true)
//   - Pipeline execution fails (when execute=true)

// GetCacheVersion retrieves or initializes a cache version.
// If the version doesn't exist, it initializes it to 1.
func (c *Cache) GetCacheVersion(ctx context.Context, namespace string, prefix string) (int, error) {
	return version.GetCacheVersion(
		ctx,
		c.RedisClient,
		c.Tracer,
		c.Logger,
		c.Semaphore,
		c.VersionGenerator,
		c.VersionExpire,
		namespace,
		prefix,
	)
}

// IncrementCacheVersion increments the version counter in a Redis pipeline.
func (c *Cache) IncrementCacheVersion(
	ctx context.Context,
	session redis.Pipeliner,
	prefix string,
	namespace string,
	ttl time.Duration,
) (int, error) {
	return version.IncrementCacheVersion(
		ctx,
		c.RedisClient,
		c.Tracer,
		c.Logger,
		c.Semaphore,
		c.VersionGenerator,
		session,
		prefix,
		namespace,
		ttl,
	)
}

// CheckVersionTtlAsync asynchronously ensures the version key has a TTL.
// If the key doesn't have a TTL, it sets one.
func (c *Cache) CheckVersionTtlAsync(span adapter.Span, key string, ttl time.Duration) {
	version.CheckVersionTtlAsync(
		c.RedisClient,
		c.Tracer,
		c.Logger,
		c.Semaphore,
		span,
		key,
		ttl,
	)
}

// InvalidateVersion invalidates all cache entries for a namespace by incrementing the version.
func (c *Cache) InvalidateVersion(ctx context.Context, namespace string, prefix string) error {
	return version.InvalidateVersion(
		ctx,
		c.RedisClient,
		c.Tracer,
		c.Logger,
		c.Semaphore,
		c.VersionGenerator,
		c.VersionExpire,
		c.KeyPrefix,
		namespace,
		prefix,
		c.GetSession,
	)
}
