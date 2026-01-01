package version

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/go-redis/redis/v8"
)

// Manager handles version-based cache invalidation.
type Manager struct {
	redisClient      *redis.Client
	tracer           adapter.Tracer
	logger           adapter.Logger
	semaphore        adapter.Semaphore
	versionGenerator key.KeyGeneratorFunc
	versionExpire    time.Duration
	keyPrefix        string
}

// NewManager creates a new version manager.
func NewManager(
	redisClient *redis.Client,
	tracer adapter.Tracer,
	logger adapter.Logger,
	semaphore adapter.Semaphore,
	versionGenerator key.KeyGeneratorFunc,
	versionExpire time.Duration,
	keyPrefix string,
) *Manager {
	return &Manager{
		redisClient:      redisClient,
		tracer:           tracer,
		logger:           logger,
		semaphore:        semaphore,
		versionGenerator: versionGenerator,
		versionExpire:    versionExpire,
		keyPrefix:        keyPrefix,
	}
}

// GetCacheVersion retrieves or initializes a cache version.
// If the version doesn't exist, it initializes it to 1.
func (m *Manager) GetCacheVersion(ctx context.Context, namespace string, prefix string) (int, error) {
	ctx, cacheSpan := m.tracer.StartSpan(ctx, "cache.getCacheVersion")
	defer cacheSpan.End()

	logger := m.logger.WithSpanContext(cacheSpan.SpanContext())

	data := map[string]string{
		"prefix":    prefix,
		"namespace": namespace,
	}

	logger.Debug("getCacheVersion", map[string]interface{}{
		"prefix":    prefix,
		"namespace": namespace,
	})

	// Generate version key
	key, err := m.versionGenerator(data)
	if err != nil {
		return 0, fmt.Errorf("failed to generate version key: %w", err)
	}

	logger.Debug("version key generated", map[string]interface{}{
		"key": key,
	})

	// Try to get existing version
	version, err := m.redisClient.Get(ctx, key).Result()
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
	versionInt, err := m.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to initialize version: %w", err)
	}

	// Asynchronously check TTL
	m.CheckVersionTtlAsync(cacheSpan, key, m.versionExpire)

	logger.Debug("version initialized", map[string]interface{}{
		"version": versionInt,
	})

	return int(versionInt), nil
}

// IncrementCacheVersion increments the version counter in a Redis pipeline.
func (m *Manager) IncrementCacheVersion(
	ctx context.Context,
	session redis.Pipeliner,
	prefix string,
	namespace string,
	ttl time.Duration,
) (int, error) {
	ctx, cacheSpan := m.tracer.StartSpan(ctx, "cache.incrementCacheVersion")
	defer cacheSpan.End()

	logger := m.logger.WithSpanContext(cacheSpan.SpanContext())

	data := map[string]string{
		"prefix":    prefix,
		"namespace": namespace,
	}

	logger.Debug("incrementCacheVersion", map[string]interface{}{
		"prefix":    prefix,
		"namespace": namespace,
	})

	// Generate version key
	key, err := m.versionGenerator(data)
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
	m.CheckVersionTtlAsync(cacheSpan, key, ttl)

	logger.Debug("version incremented", map[string]interface{}{
		"version": versionInt,
	})

	return int(versionInt), nil
}

// CheckVersionTtlAsync asynchronously ensures the version key has a TTL.
// If the key doesn't have a TTL, it sets one.
// This is exported for testing purposes.
func (m *Manager) CheckVersionTtlAsync(span adapter.Span, key string, ttl time.Duration) {
	go func() {
		m.semaphore.Acquire()
		defer m.semaphore.Release()

		newCtx, cacheTtlSpan := m.tracer.NewSpanFromSpan(
			context.Background(),
			"cache.checkVersionTtlAsync",
			span,
		)
		defer cacheTtlSpan.End()

		loggerBg := m.logger.WithSpanContext(cacheTtlSpan.SpanContext())

		// Get TTL
		ttlResult, err := m.redisClient.TTL(newCtx, key).Result()
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

			if err := m.redisClient.Expire(newCtx, key, ttl).Err(); err != nil {
				loggerBg.Error("failed to set expiration time", map[string]interface{}{
					"error": err.Error(),
					"key":   key,
				})
				return
			}
			ttlResult = ttl
		}

		loggerBg.Debug("version TTL checked", map[string]interface{}{
			"key": key,
			"ttl": ttlResult.String(),
		})
	}()
}

// InvalidateVersion invalidates all cache entries for a namespace by incrementing the version.
func (m *Manager) InvalidateVersion(ctx context.Context, namespace string, prefix string, getSession func() redis.Pipeliner) error {
	if prefix == "" {
		prefix = m.keyPrefix
	}

	ctx, cacheSpan := m.tracer.StartSpan(ctx, "cache.InvalidateVersion")
	defer cacheSpan.End()

	logger := m.logger.WithSpanContext(cacheSpan.SpanContext())

	logger.Debug("invalidating version", map[string]interface{}{
		"prefix":    prefix,
		"namespace": namespace,
	})

	// Use pipeline for atomic operation
	session := getSession()
	_, err := m.IncrementCacheVersion(ctx, session, prefix, namespace, m.versionExpire)
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
