package cache

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/errs"
	"github.com/adityakw90/go-cache/internal/hash"
	"github.com/adityakw90/go-cache/internal/lock"
	"github.com/adityakw90/go-cache/internal/serialize"
	"github.com/adityakw90/go-cache/internal/version"
)

// Cache is the main cache structure.
type Cache struct {
	redisClient         *redis.Client
	tracer              adapter.Tracer
	logger              adapter.Logger
	semaphore           adapter.Semaphore
	keyPrefix           string
	keyGenerator        KeyGeneratorFunc
	keyVersionGenerator KeyGeneratorFunc
	versionGenerator    KeyGeneratorFunc
	versionExpire       time.Duration
	lockGenerator       KeyGeneratorFunc
	lockDuration        time.Duration
	lockInterval        time.Duration
	expireDefault       time.Duration
	keyUsage            map[string][]string                     // To track cache key usage: keyUsage[keyName] = []prefix
	customKeys          map[string]map[string]CustomKeyFunction // To track custom key functions
	keyMutex            sync.Mutex                              // Mutex to handle concurrent map access
	versionManager      *version.Manager
}

// NewCache creates a new cache instance with functional options.
//
// Example:
//
//	cache, err := cache.NewCache(
//	    redisClient,
//	    cache.WithKeyPrefix("myapp"),
//	    cache.WithExpireDefault(5 * time.Minute),
//	    cache.WithTracer(myTracer),
//	    cache.WithLogger(myLogger),
//	)
func NewCache(redisClient *redis.Client, opts ...Option) (*Cache, error) {
	if redisClient == nil {
		return nil, errs.NewInvalidConfigError("redisClient", "cannot be nil")
	}

	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Use provided semaphore or create default
	semaphore := options.semaphore
	if semaphore == nil {
		semaphore = adapter.NewSemaphore(options.semaphoreSize)
	}

	// Use provided tracer or default to no-op
	tracer := options.tracer
	if tracer == nil {
		tracer = adapter.NewNoOpTracer()
	}

	// Use provided logger or default to no-op
	logger := options.logger
	if logger == nil {
		logger = adapter.NewNoOpLogger()
	}

	c := &Cache{
		redisClient:         redisClient,
		tracer:              tracer,
		logger:              logger,
		semaphore:           semaphore,
		keyPrefix:           options.keyPrefix,
		keyGenerator:        options.keyGenerator,
		keyVersionGenerator: options.keyVersionGenerator,
		versionGenerator:    options.versionGenerator,
		versionExpire:       options.versionExpire,
		lockGenerator:       options.lockGenerator,
		lockDuration:        options.lockDuration,
		lockInterval:        options.lockInterval,
		expireDefault:       options.expireDefault,
		keyUsage:            make(map[string][]string),
		customKeys:          make(map[string]map[string]CustomKeyFunction),
	}

	// Initialize version manager
	c.versionManager = version.NewManager(
		redisClient,
		tracer,
		logger,
		semaphore,
		func(data map[string]string) (string, error) {
			return options.versionGenerator(data)
		},
		options.versionExpire,
		options.keyPrefix,
	)

	return c, nil
}

// GetSession returns a Redis pipeline session.
func (c *Cache) GetSession() redis.Pipeliner {
	return c.redisClient.Pipeline()
}

// registerCacheKey registers a cache key for tracking.
// It stores the mapping: keyUsage[keyName] = []prefix
// This allows a single key to be registered with multiple prefixes.
func (c *Cache) registerCacheKey(keyName string, prefix string) {
	c.keyMutex.Lock()
	defer c.keyMutex.Unlock()

	// Check if the key is already registered
	if _, exists := c.keyUsage[keyName]; !exists {
		c.keyUsage[keyName] = []string{}
	}

	// Check if prefix already exists in the list
	for _, p := range c.keyUsage[keyName] {
		if p == prefix {
			return // Prefix is already registered
		}
	}

	// Register the new prefix
	c.keyUsage[keyName] = append(c.keyUsage[keyName], prefix)
}

// registerCustomKey registers a custom key function.
func (c *Cache) registerCustomKey(keyName string, customKeyFunc CustomKeyFunction) {
	if customKeyFunc == nil {
		return
	}

	c.keyMutex.Lock()
	defer c.keyMutex.Unlock()

	if _, exists := c.customKeys[keyName]; !exists {
		c.customKeys[keyName] = make(map[string]CustomKeyFunction)
	}

	c.customKeys[keyName][customKeyFunc.Name()] = customKeyFunc
}

// GetCacheKeyUsage returns the list of prefixes registered for a given key name.
// This is useful for debugging and monitoring cache key usage.
// Returns an empty slice if the key is not registered.
func (c *Cache) GetCacheKeyUsage(keyName string) []string {
	c.keyMutex.Lock()
	defer c.keyMutex.Unlock()

	if prefixes, exists := c.keyUsage[keyName]; exists {
		// Return a copy to prevent external modification
		result := make([]string, len(prefixes))
		copy(result, prefixes)
		return result
	}

	return []string{}
}

// getCacheHash generates an MD5 hash from function name and arguments.
func (c *Cache) getCacheHash(funcName string, args []interface{}) string {
	return hash.CacheKey(funcName, args)
}

// serialize converts a value to []byte using Gob encoding.
func (c *Cache) serialize(value interface{}) ([]byte, error) {
	return serialize.Serialize(value)
}

// deserialize converts []byte to a value using Gob decoding.
func (c *Cache) deserialize(data []byte, result interface{}) error {
	return serialize.Deserialize(data, result)
}

// acquireLock attempts to acquire a distributed lock.
func (c *Cache) acquireLock(
	ctx context.Context,
	key string,
	timeout time.Duration,
	interval time.Duration,
	wait bool,
	waitTimeout time.Duration,
) *lock.LockData {
	return lock.AcquireLock(ctx, c.redisClient, key, timeout, interval, wait, waitTimeout)
}

// releaseLock releases a distributed lock.
func (c *Cache) releaseLock(ctx context.Context, lockData *lock.LockData) {
	lock.ReleaseLock(ctx, c.redisClient, lockData)
}

// acquireMultipleLock attempts to acquire multiple locks atomically.
func (c *Cache) acquireMultipleLock(
	ctx context.Context,
	keys []string,
	timeout time.Duration,
	interval time.Duration,
	wait bool,
	waitTimeout time.Duration,
) ([]*lock.LockData, error) {
	return lock.AcquireMultipleLock(ctx, c.redisClient, keys, timeout, interval, wait, waitTimeout)
}

// releaseMultipleLock releases multiple locks.
func (c *Cache) releaseMultipleLock(ctx context.Context, locks []*lock.LockData) error {
	return lock.ReleaseMultipleLock(ctx, c.redisClient, locks)
}

// getCacheVersion retrieves or initializes a cache version.
func (c *Cache) getCacheVersion(ctx context.Context, namespace string, prefix string) (int, error) {
	return c.versionManager.GetCacheVersion(ctx, namespace, prefix)
}

// incrementCacheVersion increments the version counter in a Redis pipeline.
func (c *Cache) incrementCacheVersion(
	ctx context.Context,
	session redis.Pipeliner,
	prefix string,
	namespace string,
	ttl time.Duration,
) (int, error) {
	return c.versionManager.IncrementCacheVersion(ctx, session, prefix, namespace, ttl)
}

// InvalidateVersion invalidates all cache entries for a namespace by incrementing the version.
func (c *Cache) InvalidateVersion(ctx context.Context, namespace string, prefix string) error {
	return c.versionManager.InvalidateVersion(ctx, namespace, prefix, c.GetSession)
}

// checkVersionTtlAsync is a test helper that wraps the version manager's method.
// This is used for testing internal TTL checking behavior.
func (c *Cache) checkVersionTtlAsync(span adapter.Span, key string, ttl time.Duration) {
	c.versionManager.CheckVersionTtlAsync(span, key, ttl)
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
func (c *Cache) CleanCache(
	ctx context.Context,
	key string,
	params map[string]interface{},
	session redis.Pipeliner,
	execute bool,
	listLockedKey *[]string,
) error {
	// Start a tracing span
	ctx, cacheSpan := c.tracer.StartSpan(ctx, "cache.CleanCache")
	defer cacheSpan.End()

	// Get logger with span context
	logger := c.logger.WithSpanContext(cacheSpan.SpanContext())

	// Find all cache key usages for the provided key
	listPrefix := c.GetCacheKeyUsage(key)
	if len(listPrefix) == 0 {
		return nil
	}

	// Check if a Redis session (pipeline) is provided, if not, create a new session
	if session == nil {
		logger.Debug("creating session cache", nil)
		session = c.GetSession()
	}
	if listLockedKey == nil {
		listLockedKey = &[]string{}
	}

	// If there is a custom key logic, process that
	if customKeys, ok := c.customKeys[key]; ok {
		var listNamespace []string

		// Iterate through the custom key items and generate keys based on the params
		for _, customItem := range customKeys {
			namespace, err := customItem.Call(params)
			if err != nil {
				return err
			}
			listNamespace = append(listNamespace, namespace)
		}

		// Increment cache version for each prefix and custom namespace
		for _, prefix := range listPrefix {
			for _, namespace := range listNamespace {
				logger.Debug("clean cache", map[string]interface{}{
					"prefix":    prefix,
					"namespace": namespace,
				})
				lockKey, err := c.lockGenerator(map[string]string{
					"prefix":    prefix,
					"namespace": namespace,
				})
				if err != nil {
					logger.Error("error generating lock key", map[string]interface{}{
						"error": err.Error(),
					})
					continue
				}
				*listLockedKey = append(*listLockedKey, lockKey)
				if _, err := c.incrementCacheVersion(ctx, session, prefix, namespace, c.versionExpire); err != nil {
					return err
				}
			}
		}
	} else {
		// Increment cache version for standard keys
		for _, prefix := range listPrefix {
			logger.Debug("clean cache", map[string]interface{}{
				"prefix":    prefix,
				"namespace": key,
			})
			lockKey, err := c.lockGenerator(map[string]string{
				"prefix":    prefix,
				"namespace": key,
			})
			if err != nil {
				logger.Error("error generating lock key", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}
			*listLockedKey = append(*listLockedKey, lockKey)
			if _, err := c.incrementCacheVersion(ctx, session, prefix, key, c.versionExpire); err != nil {
				return err
			}
		}
	}

	// If `execute` is true, commit the pipeline
	if execute {
		logger.Debug("commit session cache", nil)
		// acquire lock
		locks, err := c.acquireMultipleLock(
			ctx, *listLockedKey, c.lockDuration,
			c.lockInterval, true, 5*time.Second,
		)
		if err != nil {
			logger.Error("error acquiring lock", map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}
		defer func() {
			err := c.releaseMultipleLock(ctx, locks)
			if err != nil {
				logger.Error("failed to release locks", map[string]interface{}{
					"error": err.Error(),
				})
				return
			}
			logger.Debug("multiple locks released", nil)
		}()
		if _, err := session.Exec(ctx); err != nil {
			return errs.Wrap(err, "failed to execute redis pipeline")
		}
	}

	return nil
}
