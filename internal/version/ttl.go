package version

import (
	"context"
	"fmt"
	"time"

	"github.com/adityakw90/go-cache/adapter"
	"github.com/redis/go-redis/v9"
)

func CheckVersionTtl(
	ctx context.Context,
	redisClient *redis.Client,
	startSpan adapter.StartSpan,
	semaphore adapter.Semaphore,
	getLogger func(ctx context.Context) adapter.Logger,
	key string,
	ttl time.Duration,
) (time.Duration, error) {
	ctx, cacheSpan := startSpan(ctx, "cache.checkVersionTtl")
	defer cacheSpan.End()

	logger := getLogger(ctx)

	// Get TTL
	ttlResult, err := redisClient.TTL(ctx, key).Result()
	if err != nil {
		logger.Error("failed to get TTL", map[string]interface{}{
			"error": err.Error(),
			"key":   key,
		})
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	// Set the TTL if it's not already set (ttlResult == -1 means no TTL)
	if ttlResult == -1 {
		logger.Debug("setting expiration time", map[string]interface{}{
			"key":    key,
			"expire": ttl.String(),
		})

		if err := redisClient.Expire(ctx, key, ttl).Err(); err != nil {
			logger.Error("failed to set expiration time", map[string]interface{}{
				"error": err.Error(),
				"key":   key,
			})
			return 0, fmt.Errorf("failed to set expiration time: %w", err)
		}

		return ttl, nil
	}

	return ttlResult, nil
}

// CheckVersionTtlAsync asynchronously ensures the version key has a TTL.
// If the key doesn't have a TTL, it sets one.
// This is exported for testing purposes.
func CheckVersionTtlAsync(
	redisClient *redis.Client,
	startSpan adapter.StartSpan,
	startChildSpan adapter.StartChildSpan,
	semaphore adapter.Semaphore,
	getLogger func(ctx context.Context) adapter.Logger,
	parentSpan adapter.Span,
	key string,
	ttl time.Duration,
) {
	go func() {
		semaphore.Acquire()
		defer semaphore.Release()

		newCtx, cacheTtlSpan := startChildSpan(
			context.Background(),
			"cache.checkVersionTtlAsync",
			parentSpan,
		)
		defer cacheTtlSpan.End()

		loggerBg := getLogger(newCtx)

		ttlResult, err := CheckVersionTtl(newCtx, redisClient, startSpan, semaphore, getLogger, key, ttl)
		if err != nil {
			loggerBg.Error("failed to check version TTL asynchronously", map[string]interface{}{
				"error": err.Error(),
				"key":   key,
			})
			return
		}

		loggerBg.Debug("version TTL checked asynchronously", map[string]interface{}{
			"key": key,
			"ttl": ttlResult.String(),
		})

	}()
}
