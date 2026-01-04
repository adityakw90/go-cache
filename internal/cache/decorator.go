package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/adityakw90/go-cache/internal/hash"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/adityakw90/go-cache/internal/lock"
	moduleVersion "github.com/adityakw90/go-cache/internal/version"
)

func (c *Cache) Cached(
	keyName string,
	ttl interface{},
	versioning bool,
	prefix string,
) func(
	fn func(ctx context.Context, args ...interface{}) (interface{}, error),
	customKeyFunc key.CustomKeyFunction,
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
	) func(resultType interface{}, ctx context.Context, args ...interface{}) (interface{}, error) {
		// resultType is a pointer to the result of the function
		if customKeyFunc != nil {
			// Register custom key function
			c.registerCustomKey(keyName, customKeyFunc)
		}

		return func(resultType interface{}, ctx context.Context, args ...interface{}) (interface{}, error) {
			// Create a tracing span
			ctx, cacheSpan := c.Tracer.StartSpan(ctx, "cache.Cached")
			defer cacheSpan.End()

			// Get logger with span context
			logger := c.Logger.WithSpanContext(cacheSpan.SpanContext())

			// Determine namespace
			var namespace string
			if customKeyFunc != nil {
				namespace = customKeyFunc.Callable(args...)
			} else {
				namespace = keyName
			}

			logger.Debug("cache namespace", map[string]interface{}{
				"namespace": namespace,
			})

			// Get cache version if versioning is enabled
			version := 0
			if versioning {
				var err error
				version, err = moduleVersion.GetCacheVersion(
					ctx, c.RedisClient, c.Tracer, c.Logger, c.Semaphore,
					c.VersionGenerator, c.VersionExpire, namespace, prefix,
				)
				if err != nil {
					logger.Info("failed to get cache version", map[string]interface{}{
						"error": err.Error(),
					})
					// Fall back to executing the function
					return fn(ctx, args...)
				}
			}

			logger.Debug("cache version", map[string]interface{}{
				"version": version,
			})

			// Generate hash from function arguments
			hashKey := hash.CacheKey(namespace, args)

			// Generate cache key
			key, err := c.KeyVersionGenerator(map[string]string{
				"prefix":    prefix,
				"namespace": namespace,
				"version":   strconv.Itoa(version),
				"key":       hashKey,
			})
			if err != nil {
				logger.Error("error generating cache key", map[string]interface{}{
					"error": err.Error(),
				})
				// Fall back to executing the function
				return fn(ctx, args...)
			}

			logger.Debug("cache key generated", map[string]interface{}{
				"key": key,
			})

			// Generate lock key early (needed for cache miss handling)
			lockKey, err := c.LockGenerator(map[string]string{
				"prefix":    prefix,
				"namespace": namespace,
			})
			if err != nil {
				logger.Error("error generating lock key", map[string]interface{}{
					"error": err.Error(),
				})
				// Fall back to executing the function
				return fn(ctx, args...)
			}

			// Calculate wait timeout for lock acquisition
			waitTimeout := c.LockDuration * 2
			if waitTimeout < 5*time.Second {
				waitTimeout = 5 * time.Second
			}

			// Create timeout context for the retry loop
			timeoutCtx, cancel := context.WithTimeout(ctx, waitTimeout)
			defer cancel()

			// Loop to check cache and handle cache miss with write lock
			for {
				err = c.Get(ctx, key, resultType)
				if err == nil {
					// Cache hit - return cached data
					logger.Info("cache hit", map[string]interface{}{
						"key":         key,
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
					"key": key,
				})

				lockData := lock.AcquireLock(
					ctx, c.RedisClient, lockKey, c.LockDuration,
					c.LockInterval, false, waitTimeout,
				)

				// Check if lock was acquired
				if lockData.Error == nil && lockData.Acquired {
					// Lock acquired - this is the first request
					logger.Debug("write lock acquired", map[string]interface{}{
						"key":   lockData.Key,
						"token": lockData.Token,
					})

					// Set up defer to release lock when function exits
					defer func() {
						lock.ReleaseLock(ctx, c.RedisClient, lockData)
						if lockData.Error != nil {
							logger.Debug("failed to release lock", map[string]interface{}{
								"error": lockData.Error.Error(),
							})
						} else {
							logger.Debug("lock released", map[string]interface{}{
								"key":   lockData.Key,
								"token": lockData.Token,
							})
						}
					}()

					// Re-check cache after acquiring lock (another request might have populated it)
					err = c.Get(ctx, key, resultType)
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

					// Determine TTL
					var ttlValue time.Duration
					switch v := ttl.(type) {
					case time.Duration:
						ttlValue = v
					case func(result interface{}, args ...interface{}) time.Duration:
						ttlValue = v(result, args...)
					default:
						ttlValue = c.ExpireDefault
					}

					// Update cache synchronously using Set method
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

				// Lock not acquired - wait and retry checking cache
				// Subsequent requests will wait here and keep checking cache
				select {
				case <-time.After(c.LockInterval):
					// Retry after the retry interval
					continue
				case <-timeoutCtx.Done():
					// Timeout exceeded - fallback to executing function
					logger.Info("timeout waiting for cache", map[string]interface{}{
						"key": key,
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
