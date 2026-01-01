package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cached creates a cache decorator for a function.
// It returns a function that wraps the original function with caching logic.
//
// Parameters:
//   - keyName: The name of the cache key
//   - ttl: Time-to-live for cache entries. Can be time.Duration or func(result, args...) time.Duration
//   - versioning: Whether to use version-based cache invalidation
//   - prefix: Cache key prefix (uses default if empty)
//
// Returns a function that takes:
//   - fn: The function to cache
//   - customKeyFunc: Optional custom key function
//
// Which returns a cached function that takes:
//   - resultType: Pointer to the result type (for deserialization)
//   - ctx: Context
//   - args: Function arguments
//
// Example:
//
//	cachedFunc := cache.Cached(
//	    "getUser",
//	    10 * time.Minute,
//	    true,
//	    "user",
//	)(
//	    getUserFromDB,
//	    &cache.CustomKeyFunction{
//	        Name: "getUser",
//	        Callable: func(args ...interface{}) string {
//	            return fmt.Sprintf("user:%s", args[0])
//	        },
//	        Params: []string{"uid"},
//	    },
//	)
//
//	result, err := cachedFunc(&User{}, ctx, "user-123")
func (c *Cache) Cached(
	keyName string,
	ttl interface{},
	versioning bool,
	prefix string,
) func(
	fn func(ctx context.Context, args ...interface{}) (interface{}, error),
	customKeyFunc CustomKeyFunction,
) func(resultType interface{}, ctx context.Context, args ...interface{}) (interface{}, error) {
	// Set default prefix if not provided
	if prefix == "" {
		prefix = c.keyPrefix
	}

	// Register cache key usage
	c.registerCacheKey(keyName, prefix)

	return func(
		fn func(ctx context.Context, args ...interface{}) (interface{}, error),
		customKeyFunc CustomKeyFunction,
	) func(resultType interface{}, ctx context.Context, args ...interface{}) (interface{}, error) {
		if customKeyFunc != nil {
			// Register custom key function
			c.registerCustomKey(keyName, customKeyFunc)
		}

		return func(resultType interface{}, ctx context.Context, args ...interface{}) (interface{}, error) {
			// Create a tracing span
			ctx, cacheSpan := c.tracer.StartSpan(ctx, "cache.Cached")
			defer cacheSpan.End()

			// Get logger with span context
			logger := c.logger.WithSpanContext(cacheSpan.SpanContext())

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
				version, err = c.getCacheVersion(ctx, namespace, prefix)
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
			hashKey := c.getCacheHash(namespace, args)

			// Generate cache key
			key, err := c.keyVersionGenerator(map[string]string{
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

			// Attempt to load from cache
			cachedData, err := c.redisClient.Get(ctx, key).Bytes()
			if err == redis.Nil {
				// Cache miss
				logger.Info("cache miss", map[string]interface{}{
					"key": key,
				})

				// Execute the original function
				result, err := fn(ctx, args...)
				if err != nil {
					return nil, err
				}

				// Asynchronously update the cache
				go func() {
					c.semaphore.Acquire()
					defer c.semaphore.Release()

					newCtx, cacheUpdateSpan := c.tracer.NewSpanFromSpan(
						context.Background(),
						"cache.Cached.Update",
						cacheSpan,
					)
					defer cacheUpdateSpan.End()

					loggerBg := c.logger.WithSpanContext(cacheUpdateSpan.SpanContext())

					// Serialize the result
					serializedData, err := c.serialize(result)
					if err != nil {
						loggerBg.Error("failed to serialize result", map[string]interface{}{
							"error": err.Error(),
						})
						return
					}

					// Determine TTL
					var ttlValue time.Duration
					switch v := ttl.(type) {
					case time.Duration:
						ttlValue = v
					case func(result interface{}, args ...interface{}) time.Duration:
						ttlValue = v(result, args...)
					default:
						ttlValue = c.expireDefault
					}

					// Generate lock key
					lockKey, err := c.lockGenerator(map[string]string{
						"prefix":    prefix,
						"namespace": namespace,
					})
					if err != nil {
						loggerBg.Error("error generating lock key", map[string]interface{}{
							"error": err.Error(),
						})
						return
					}

					// Acquire lock to prevent cache stampede
					lock := c.acquireLock(
						newCtx, lockKey, c.lockDuration,
						c.lockInterval, false, 5*time.Second,
					)
					if lock.Error != nil {
						loggerBg.Error("error acquiring lock", map[string]interface{}{
							"error": lock.Error.Error(),
						})
						return
					}
					if !lock.Acquired {
						loggerBg.Error("failed to acquire lock", nil)
						return
					}
					defer func() {
						c.releaseLock(newCtx, lock)
						if lock.Error != nil {
							loggerBg.Error("failed to release lock", map[string]interface{}{
								"error": lock.Error.Error(),
							})
							return
						}
						loggerBg.Debug("lock released", map[string]interface{}{
							"key":   lock.Key,
							"token": lock.Token,
						})
					}()

					loggerBg.Debug("lock acquired", map[string]interface{}{
						"key":   lock.Key,
						"token": lock.Token,
					})

					// Set the cache value
					if err := c.redisClient.Set(newCtx, key, serializedData, ttlValue).Err(); err != nil {
						loggerBg.Error("failed to set cache data", map[string]interface{}{
							"error": err.Error(),
						})
						return
					}

					loggerBg.Debug("cache updated", map[string]interface{}{
						"key":         key,
						"ttl":         ttlValue.String(),
						"result_type": fmt.Sprintf("%T", resultType),
					})
				}()

				return result, nil
			} else if err != nil {
				// Error getting from cache
				logger.Info("failed to get cache", map[string]interface{}{
					"error": err.Error(),
				})
				// Fall back to executing the function
				return fn(ctx, args...)
			}

			// Cache hit
			logger.Info("cache hit", map[string]interface{}{
				"key":         key,
				"result_type": fmt.Sprintf("%T", resultType),
			})

			// Deserialize the cached data
			err = c.deserialize(cachedData, resultType)
			if err != nil {
				logger.Error("failed to deserialize cache data", map[string]interface{}{
					"error": err.Error(),
				})
				// Fall back to executing the function
				return fn(ctx, args...)
			}

			return resultType, nil
		}
	}
}
