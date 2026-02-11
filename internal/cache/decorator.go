package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/adityakw90/go-cache/adapter"
	"github.com/adityakw90/go-cache/internal/hash"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/adityakw90/go-cache/internal/lock"
	moduleVersion "github.com/adityakw90/go-cache/internal/version"
)

// Constants for cache configuration defaults.
const (
	// minLockWaitTimeout is the minimum wait timeout for lock acquisition.
	// This prevents extremely short timeouts that could cause lock starvation.
	minLockWaitTimeout = 5 * time.Second
)

// determineNamespace returns the namespace for cache operations.
// If a custom key function is provided, it generates the namespace from the function arguments.
// Otherwise, it uses the keyName as the namespace.
func (c *Cache) determineNamespace(
	keyName string,
	customKeyFunc key.CustomKeyFunction,
	args []interface{},
) string {
	if customKeyFunc != nil {
		return customKeyFunc.Callable(args...)
	}
	return keyName
}

// getCacheVersion retrieves the cache version for the given namespace.
// Returns 0 if versioning is disabled.
// Falls back to executing the function if version retrieval fails.
func (c *Cache) getCacheVersion(
	ctx context.Context,
	namespace string,
	prefix string,
	versioning bool,
) (int, error) {
	if !versioning {
		return 0, nil
	}

	version, err := moduleVersion.GetCacheVersion(
		ctx, c.RedisClient, c.StartSpan, c.StartChildSpan, c.Semaphore, c.LogProvider,
		c.VersionGenerator, c.VersionExpire, namespace, prefix,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get cache version: %w", err)
	}

	return version, nil
}

// generateCacheKey creates a cache key from the namespace, version, and hash of arguments.
// When useHashKey is true, it uses KeyHashedVersionGenerator which includes the key in the output.
// When useHashKey is false, it uses KeyVersionGenerator which produces a simpler key format.
func (c *Cache) generateCacheKey(
	prefix string,
	namespace string,
	version int,
	args []interface{},
	useHashKey bool,
) (string, error) {
	var cacheKey string
	var err error

	if useHashKey {
		hashKey := hash.CacheKey(namespace, args)
		// Use KeyHashedVersionGenerator which includes the key in the output
		// Format: {prefix}:{namespace}:v{version}-{key}.gob
		cacheKey, err = c.KeyHashedVersionGenerator(map[string]string{
			"prefix":    prefix,
			"namespace": namespace,
			"version":   strconv.Itoa(version),
			"key":       hashKey,
		})
	} else {
		// Use KeyVersionGenerator which produces a simpler key format
		// Format: {prefix}:{namespace}:v{version}.gob
		cacheKey, err = c.KeyVersionGenerator(map[string]string{
			"prefix":    prefix,
			"namespace": namespace,
			"version":   strconv.Itoa(version),
		})
	}

	if err != nil {
		return "", fmt.Errorf("error generating cache key: %w", err)
	}

	return cacheKey, nil
}

// generateLockKey creates a lock key for the given namespace.
func (c *Cache) generateLockKey(
	prefix string,
	namespace string,
) (string, error) {
	lockKey, err := c.LockGenerator(map[string]string{
		"prefix":    prefix,
		"namespace": namespace,
	})
	if err != nil {
		return "", fmt.Errorf("error generating lock key: %w", err)
	}

	return lockKey, nil
}

// calculateWaitTimeout determines the wait timeout for lock acquisition.
// The timeout is at least minLockWaitTimeout and defaults to 2x the lock duration.
func (c *Cache) calculateWaitTimeout() time.Duration {
	waitTimeout := c.LockDuration * 2
	if waitTimeout < minLockWaitTimeout {
		waitTimeout = minLockWaitTimeout
	}
	return waitTimeout
}

// determineTTL resolves the TTL value for cache storage.
// It can be a fixed duration, a dynamic function, or the default.
func (c *Cache) determineTTL(
	ttl interface{},
	result interface{},
	args []interface{},
) time.Duration {
	var ttlValue time.Duration
	switch v := ttl.(type) {
	case time.Duration:
		ttlValue = v
	case func(result interface{}, args ...interface{}) time.Duration:
		ttlValue = v(result, args...)
	default:
		ttlValue = c.ExpireDefault
	}
	return ttlValue
}

// handleCacheMiss handles the cache miss scenario by acquiring a lock
// and either returning cached data (populated by another request) or
// executing the function and updating the cache.
func (c *Cache) handleCacheMiss(
	ctx context.Context,
	key string,
	resultType interface{},
	fn func(ctx context.Context, args ...interface{}) (interface{}, error),
	lockData *lock.LockData,
	logger adapter.Logger,
	ttl interface{},
	args []interface{},
) (interface{}, error) {
	// Lock acquired - this is the first request
	logger.Debug("write lock acquired", map[string]interface{}{
		"key":   lockData.Key,
		"token": lockData.Token,
	})

	// Set up defer to release lock when function exits
	defer func() {
		if err := lock.ReleaseLock(ctx, c.RedisClient, lockData); err != nil {
			logger.Debug("failed to release lock", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			logger.Debug("lock released", map[string]interface{}{
				"key":   lockData.Key,
				"token": lockData.Token,
			})
		}
	}()

	// Re-check cache after acquiring lock (another request might have populated it)
	err := c.Get(ctx, key, resultType)
	if err == nil {
		// Cache was populated by another request, defer will release lock
		logger.Info("cache hit after lock acquisition", map[string]interface{}{
			"key": key,
		})
		return resultType, nil
	}

	// Still cache miss - execute function and update cache
	result, err := fn(ctx, args...)
	if err != nil {
		return nil, err
	}

	// Update cache synchronously using Set method
	ttlValue := c.determineTTL(ttl, result, args)
	if err := c.Set(ctx, key, result, ttlValue); err != nil {
		logger.Error("failed to set cache data", map[string]interface{}{
			"error": err.Error(),
		})
		// Continue to return result (defer will release lock)
	} else {
		logger.Debug("cache updated", map[string]interface{}{
			"key":         key,
			"ttl":         ttlValue.String(),
			"result_type": fmt.Sprintf("%T", resultType),
		})
	}

	return result, nil
}

// Cached returns a decorator function that wraps the given function with caching logic.
// The decorator handles cache key generation, versioning, and distributed locking
// to ensure thread-safe cache operations.
func (c *Cache) Cached(
	keyName string,
	ttl interface{},
	versioning bool,
	prefix string,
) func(
	fn func(ctx context.Context, args ...interface{}) (interface{}, error),
	customKeyFunc key.CustomKeyFunction,
	useHashKey bool,
) func(resultType interface{}, ctx context.Context, args ...interface{}) (interface{}, error) {
	// Set default prefix if not provided
	if prefix == "" {
		prefix = c.KeyPrefix
	}

	// Register cache key usage
	c.registerCacheKey(keyName, prefix)

	return func(
		fn func(ctx context.Context, args ...interface{}) (interface{}, error),
		customKeyFunc key.CustomKeyFunction,
		useHashKey bool,
	) func(resultType interface{}, ctx context.Context, args ...interface{}) (interface{}, error) {
		// resultType is a pointer to the result of the function
		if customKeyFunc != nil {
			// Register custom key function
			c.registerCustomKey(keyName, customKeyFunc)
		}

		return func(resultType interface{}, ctx context.Context, args ...interface{}) (interface{}, error) {
			// Create a tracing span
			ctx, cacheSpan := c.StartSpan(ctx, "cache.Cached")
			defer cacheSpan.End()

			// Get logger with span context
			logger := c.LogProvider(ctx)

			// Determine namespace
			namespace := c.determineNamespace(keyName, customKeyFunc, args)
			logger.Debug("cache namespace", map[string]interface{}{
				"namespace": namespace,
			})

			// Get cache version if versioning is enabled
			version, err := c.getCacheVersion(ctx, namespace, prefix, versioning)
			if err != nil {
				// Fall back to executing the function
				logger.Error("error getting cache version", map[string]interface{}{
					"error": err.Error(),
				})
				return fn(ctx, args...)
			}

			logger.Debug("cache version", map[string]interface{}{
				"version": version,
			})

			// Generate cache key
			cacheKey, err := c.generateCacheKey(prefix, namespace, version, args, useHashKey)
			if err != nil {
				// Fall back to executing the function
				logger.Error("error generating cache key", map[string]interface{}{
					"error": err.Error(),
				})
				return fn(ctx, args...)
			}

			// Generate lock key early (needed for cache miss handling)
			lockKey, err := c.generateLockKey(prefix, namespace)
			if err != nil {
				// Fall back to executing the function
				logger.Error("error generating lock key", map[string]interface{}{
					"error": err.Error(),
				})
				return fn(ctx, args...)
			}

			// Calculate wait timeout for lock acquisition
			waitTimeout := c.calculateWaitTimeout()

			// Create timeout context for the retry loop
			timeoutCtx, cancel := context.WithTimeout(ctx, waitTimeout)
			defer cancel()

			// Loop to check cache and handle cache miss with write lock
			for {
				err = c.Get(ctx, cacheKey, resultType)
				if err == nil {
					// Cache hit - return cached data
					logger.Info("cache hit", map[string]interface{}{
						"key":         cacheKey,
						"result_type": fmt.Sprintf("%T", resultType),
					})
					return resultType, nil
				}

				// Check if it's a cache miss
				if err != ErrGetCacheMiss {
					// Error getting from cache (other than cache miss)
					logger.Info("failed to get cache", map[string]interface{}{
						"error": err.Error(),
					})
					// Fall back to executing the function
					return fn(ctx, args...)
				}

				// Cache miss - try to acquire write lock
				logger.Info("cache miss", map[string]interface{}{
					"key": cacheKey,
				})

				lockData, err := lock.AcquireLock(
					ctx, c.RedisClient, lockKey, c.LockDuration,
					c.LockInterval, false, waitTimeout,
				)

				// Check if lock was acquired
				if err == nil && lockData != nil && lockData.Acquired {
					// Handle cache miss with lock
					return c.handleCacheMiss(
						ctx, cacheKey, resultType, fn, lockData, logger, ttl, args,
					)
				}

				// Lock not acquired or error occurred - wait and retry checking cache
				// Subsequent requests will wait here and keep checking cache
				if err != nil {
					logger.Info("failed to acquire lock, will retry", map[string]interface{}{
						"error": err.Error(),
					})
				}
				select {
				case <-time.After(c.LockInterval):
					// Retry after the retry interval
					continue
				case <-timeoutCtx.Done():
					// Timeout exceeded - fallback to executing function
					logger.Info("timeout waiting for cache", map[string]interface{}{
						"key": cacheKey,
					})
					return fn(ctx, args...)
				case <-ctx.Done():
					// Context canceled
					return fn(ctx, args...)
				}
			}
		}
	}
}
