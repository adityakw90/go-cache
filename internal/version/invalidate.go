package version

import (
	"context"
	"fmt"
	"time"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/redis/go-redis/v9"
)

// InvalidateVersion invalidates all cache entries for a namespace by incrementing the version.
func InvalidateVersion(
	ctx context.Context,
	redisClient *redis.Client,
	tracer adapter.Tracer,
	logger adapter.Logger,
	semaphore adapter.Semaphore,
	versionGenerator key.KeyGeneratorFunc,
	versionExpire time.Duration,
	keyPrefix string,
	namespace string,
	prefix string,
	getSession func() redis.Pipeliner,
) error {
	if prefix == "" {
		prefix = keyPrefix
	}

	ctx, cacheSpan := tracer.StartSpan(ctx, "cache.InvalidateVersion")
	defer cacheSpan.End()

	logger = logger.WithSpanContext(cacheSpan.SpanContext())

	logger.Debug("invalidating version", map[string]interface{}{
		"prefix":    prefix,
		"namespace": namespace,
	})

	// Use pipeline for atomic operation
	session := getSession()
	_, err := IncrementCacheVersion(ctx, redisClient, tracer, logger, semaphore, versionGenerator, session, prefix, namespace, versionExpire)
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
