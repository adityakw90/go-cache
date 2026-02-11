package version

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/adityakw90/go-cache/adapter"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/redis/go-redis/v9"
)

// GetCacheVersion retrieves or initializes a cache version.
// If the version doesn't exist, it initializes it to 1.
func GetCacheVersion(
	ctx context.Context,
	redisClient *redis.Client,
	startSpan adapter.StartSpan,
	startChildSpan adapter.StartChildSpan,
	semaphore adapter.Semaphore,
	getLogger func(ctx context.Context) adapter.Logger,
	versionGenerator key.KeyGeneratorFunc,
	versionExpire time.Duration,
	namespace string,
	prefix string,
) (int, error) {
	ctx, cacheSpan := startSpan(ctx, "cache.getCacheVersion")
	defer cacheSpan.End()

	logger := getLogger(ctx)

	data := map[string]string{
		"prefix":    prefix,
		"namespace": namespace,
	}

	logger.Debug("getCacheVersion", map[string]interface{}{
		"prefix":    prefix,
		"namespace": namespace,
	})

	// Generate version key
	key, err := versionGenerator(data)
	if err != nil {
		return 0, fmt.Errorf("failed to generate version key: %w", err)
	}

	logger.Debug("version key generated", map[string]interface{}{
		"key": key,
	})

	// Try to get existing version
	version, err := redisClient.Get(ctx, key).Result()
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
	versionInt, err := redisClient.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to initialize version: %w", err)
	}

	// Asynchronously check TTL
	CheckVersionTtlAsync(
		redisClient, startSpan, startChildSpan, semaphore, getLogger,
		cacheSpan, key, versionExpire,
	)

	logger.Debug("version initialized", map[string]interface{}{
		"version": versionInt,
	})

	return int(versionInt), nil
}

// IncrementCacheVersion increments the version counter in a Redis pipeline.
// TODO: for future release, we should return the command instead of the result
func IncrementCacheVersion(
	ctx context.Context,
	redisClient *redis.Client,
	startSpan adapter.StartSpan,
	startChildSpan adapter.StartChildSpan,
	semaphore adapter.Semaphore,
	getLogger func(ctx context.Context) adapter.Logger,
	versionGenerator key.KeyGeneratorFunc,
	session redis.Pipeliner,
	prefix string,
	namespace string,
	ttl time.Duration,
) (int, error) {
	ctx, cacheSpan := startSpan(ctx, "cache.incrementCacheVersion")
	defer cacheSpan.End()

	logger := getLogger(ctx)

	data := map[string]string{
		"prefix":    prefix,
		"namespace": namespace,
	}

	logger.Debug("incrementCacheVersion", map[string]interface{}{
		"prefix":    prefix,
		"namespace": namespace,
	})

	// Generate version key
	key, err := versionGenerator(data)
	if err != nil {
		return 0, fmt.Errorf("failed to generate version key: %w", err)
	}

	logger.Debug("version key generated", map[string]interface{}{
		"key": key,
	})

	// Increment the version in Redis
	versionInt, err := redisClient.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment version (key: %s): %w", key, err)
	}

	// Asynchronously check TTL
	CheckVersionTtlAsync(
		redisClient, startSpan, startChildSpan, semaphore, getLogger,
		cacheSpan, key, ttl,
	)

	logger.Debug("version incremented", map[string]interface{}{
		"version": versionInt,
	})

	return int(versionInt), nil
}
