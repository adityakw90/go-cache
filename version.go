package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// getCacheVersion retrieves or initializes a cache version.
// If the version doesn't exist, it initializes it to 1.
func (c *Cache) getCacheVersion(ctx context.Context, namespace string, prefix string) (int, error) {
	ctx, cacheSpan := c.tracer.StartSpan(ctx, "cache.getCacheVersion")
	defer cacheSpan.End()

	logger := c.logger.WithSpanContext(cacheSpan.SpanContext())

	data := map[string]string{
		"prefix":    prefix,
		"namespace": namespace,
	}

	logger.Debug("getCacheVersion", map[string]interface{}{
		"prefix":    prefix,
		"namespace": namespace,
	})

	// Generate version key
	key, err := c.versionGenerator(data)
	if err != nil {
		return 0, fmt.Errorf("failed to generate version key: %w", err)
	}

	logger.Debug("version key generated", map[string]interface{}{
		"key": key,
	})

	// Try to get existing version
	version, err := c.redisClient.Get(ctx, key).Result()
	if err == nil {
		// Version exists, return it as an integer
		versionInt, err := strconv.Atoi(version)
		if err != nil {
			return 0, fmt.Errorf("failed to parse version: %w", err)
		}
		logger.Debug("version retrieved", map[string]interface{}{
			"version": versionInt,
		})
		return versionInt, nil
	} else if err != redis.Nil {
		// Error other than key not found
		return 0, fmt.Errorf("failed to get version: %w", err)
	}

	// Version doesn't exist, initialize it to 1
	versionInt, err := c.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to initialize version: %w", err)
	}

	// Asynchronously check TTL
	c.checkVersionTtlAsync(cacheSpan, key, c.versionExpire)

	logger.Debug("version initialized", map[string]interface{}{
		"version": versionInt,
	})

	return int(versionInt), nil
}

// incrementCacheVersion increments the version counter in a Redis pipeline.
func (c *Cache) incrementCacheVersion(
	ctx context.Context,
	session redis.Pipeliner,
	prefix string,
	namespace string,
	ttl time.Duration,
) (int, error) {
	ctx, cacheSpan := c.tracer.StartSpan(ctx, "cache.incrementCacheVersion")
	defer cacheSpan.End()

	logger := c.logger.WithSpanContext(cacheSpan.SpanContext())

	data := map[string]string{
		"prefix":    prefix,
		"namespace": namespace,
	}

	logger.Debug("incrementCacheVersion", map[string]interface{}{
		"prefix":    prefix,
		"namespace": namespace,
	})

	// Generate version key
	key, err := c.versionGenerator(data)
	if err != nil {
		return 0, fmt.Errorf("failed to generate version key: %w", err)
	}

	logger.Debug("version key generated", map[string]interface{}{
		"key": key,
	})

	// Increment the version in Redis pipeline
	versionInt, err := session.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment version: %w", err)
	}

	// Asynchronously check TTL
	c.checkVersionTtlAsync(cacheSpan, key, ttl)

	logger.Debug("version incremented", map[string]interface{}{
		"version": versionInt,
	})

	return int(versionInt), nil
}

// checkVersionTtlAsync asynchronously ensures the version key has a TTL.
// If the key doesn't have a TTL, it sets one.
func (c *Cache) checkVersionTtlAsync(span Span, key string, ttl time.Duration) {
	go func() {
		c.semaphore.Acquire()
		defer c.semaphore.Release()

		newCtx, cacheTtlSpan := c.tracer.NewSpanFromSpan(
			context.Background(),
			"cache.checkVersionTtlAsync",
			span,
		)
		defer cacheTtlSpan.End()

		loggerBg := c.logger.WithSpanContext(cacheTtlSpan.SpanContext())

		// Get TTL
		ttlResult, err := c.redisClient.TTL(newCtx, key).Result()
		if err != nil {
			loggerBg.Error("failed to get TTL", map[string]interface{}{
				"error": err.Error(),
				"key":   key,
			})
			return
		}

		// Set the TTL if it's not already set (ttlResult == -1 means no TTL)
		if ttlResult == -1 {
			loggerBg.Debug("setting expiration time", map[string]interface{}{
				"key":    key,
				"expire": ttl.String(),
			})

			if err := c.redisClient.Expire(newCtx, key, ttl).Err(); err != nil {
				loggerBg.Error("failed to set expiration time", map[string]interface{}{
					"error": err.Error(),
					"key":   key,
				})
				return
			}
		}

		loggerBg.Debug("version TTL checked", map[string]interface{}{
			"key": key,
			"ttl": ttlResult.String(),
		})
	}()
}

// InvalidateVersion invalidates all cache entries for a namespace by incrementing the version.
func (c *Cache) InvalidateVersion(ctx context.Context, namespace string, prefix string) error {
	if prefix == "" {
		prefix = c.keyPrefix
	}

	ctx, cacheSpan := c.tracer.StartSpan(ctx, "cache.InvalidateVersion")
	defer cacheSpan.End()

	logger := c.logger.WithSpanContext(cacheSpan.SpanContext())

	logger.Debug("invalidating version", map[string]interface{}{
		"prefix":    prefix,
		"namespace": namespace,
	})

	// Use pipeline for atomic operation
	session := c.GetSession()
	_, err := c.incrementCacheVersion(ctx, session, prefix, namespace, c.versionExpire)
	if err != nil {
		return fmt.Errorf("failed to increment version: %w", err)
	}

	// Execute the pipeline
	if _, err := session.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	logger.Debug("version invalidated", map[string]interface{}{
		"prefix":    prefix,
		"namespace": namespace,
	})

	return nil
}
