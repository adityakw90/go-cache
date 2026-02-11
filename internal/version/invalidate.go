package version

import (
	"context"
	"fmt"
	"time"

	"github.com/adityakw90/go-cache/adapter"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/redis/go-redis/v9"
)

// InvalidateVersion invalidates all cache entries for a namespace by incrementing the version.
func InvalidateVersion(
	ctx context.Context,
	redisClient *redis.Client,
	startSpan adapter.StartSpan,
	startChildSpan adapter.StartChildSpan,
	semaphore adapter.Semaphore,
	getLogger func(ctx context.Context) adapter.Logger,
	versionGenerator key.KeyGeneratorFunc,
	versionExpire time.Duration,
	keyPrefix string,
	namespace string,
	prefix string,
	getSession func() redis.Pipeliner,
) error {
	ctx, cacheSpan := startSpan(ctx, "cache.InvalidateVersion")
	defer cacheSpan.End()

	logger := getLogger(ctx)

	logger.Debug("invalidating version", map[string]interface{}{
		"prefix":    prefix,
		"namespace": namespace,
	})

	if prefix == "" {
		prefix = keyPrefix
	}

	// Use pipeline for atomic operation
	session := getSession()
	_, err := IncrementCacheVersion(
		ctx, redisClient, startSpan, startChildSpan, semaphore, getLogger,
		versionGenerator, session, prefix, namespace, versionExpire,
	)
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
