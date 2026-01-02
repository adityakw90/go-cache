package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/adityakw90/go-cache/internal/hash"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/adityakw90/go-cache/internal/lock"
	"github.com/adityakw90/go-cache/internal/serialize"
	moduleVersion "github.com/adityakw90/go-cache/internal/version"
	"github.com/go-redis/redis/v8"
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
	c.RegisterCacheKey(keyName, prefix)

	return func(
		fn func(ctx context.Context, args ...interface{}) (interface{}, error),
		customKeyFunc key.CustomKeyFunction,
	) func(resultType interface{}, ctx context.Context, args ...interface{}) (interface{}, error) {
		if customKeyFunc != nil {
			// Register custom key function
			c.RegisterCustomKey(keyName, customKeyFunc)
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

			// Attempt to load from cache
			cachedData, err := c.RedisClient.Get(ctx, key).Bytes()
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
					c.Semaphore.Acquire()
					defer c.Semaphore.Release()

					newCtx, cacheUpdateSpan := c.Tracer.NewSpanFromSpan(
						context.Background(),
						"cache.Cached.Update",
						cacheSpan,
					)
					defer cacheUpdateSpan.End()

					loggerBg := c.Logger.WithSpanContext(cacheUpdateSpan.SpanContext())

					// Serialize the result
					serializedData, err := serialize.Serialize(result)
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
						ttlValue = c.ExpireDefault
					}

					// Generate lock key
					lockKey, err := c.LockGenerator(map[string]string{
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
					lockData := lock.AcquireLock(
						newCtx, c.RedisClient, lockKey, c.LockDuration,
						c.LockInterval, false, 5*time.Second,
					)
					if lockData.Error != nil {
						loggerBg.Error("error acquiring lock", map[string]interface{}{
							"error": lockData.Error.Error(),
						})
						return
					}
					if !lockData.Acquired {
						loggerBg.Error("failed to acquire lock", nil)
						return
					}
					defer func() {
						lock.ReleaseLock(newCtx, c.RedisClient, lockData)
						if lockData.Error != nil {
							loggerBg.Error("failed to release lock", map[string]interface{}{
								"error": lockData.Error.Error(),
							})
							return
						}
						loggerBg.Debug("lock released", map[string]interface{}{
							"key":   lockData.Key,
							"token": lockData.Token,
						})
					}()

					loggerBg.Debug("lock acquired", map[string]interface{}{
						"key":   lockData.Key,
						"token": lockData.Token,
					})

					// Set the cache value
					if err := c.RedisClient.Set(newCtx, key, serializedData, ttlValue).Err(); err != nil {
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
			err = serialize.Deserialize(cachedData, resultType)
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
