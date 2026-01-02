package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/adityakw90/go-cache/internal/lock"
	"github.com/adityakw90/go-cache/internal/version"
	"github.com/go-redis/redis/v8"
)

func (c *Cache) CleanCache(
	ctx context.Context,
	key string,
	params map[string]interface{},
	session redis.Pipeliner,
	execute bool,
	listLockedKey *[]string,
) error {
	// Start a tracing span
	ctx, cacheSpan := c.Tracer.StartSpan(ctx, "cache.CleanCache")
	defer cacheSpan.End()

	// Get logger with span context
	logger := c.Logger.WithSpanContext(cacheSpan.SpanContext())

	// Find all cache key usages for the provided key
	listPrefix := c.getCacheKeyUsage(key)
	if len(listPrefix) == 0 {
		return nil
	}

	// Check if a Redis session (pipeline) is provided, if not, create a new session
	if session == nil {
		logger.Debug("creating session cache", nil)
		session = c.getSession()
	}
	if listLockedKey == nil {
		listLockedKey = &[]string{}
	}

	// If there is a custom key logic, process that
	if customKeys, ok := c.customKeys[key]; ok {
		var listNamespace []string

		// Iterate through the custom key items and generate keys based on the params
		for _, customItem := range customKeys {
			namespace, err := customItem.Call(params)
			if err != nil {
				return err
			}
			listNamespace = append(listNamespace, namespace)
		}

		// Increment cache version for each prefix and custom namespace
		for _, prefix := range listPrefix {
			for _, namespace := range listNamespace {
				logger.Debug("clean cache", map[string]interface{}{
					"prefix":    prefix,
					"namespace": namespace,
				})
				lockKey, err := c.LockGenerator(map[string]string{
					"prefix":    prefix,
					"namespace": namespace,
				})
				if err != nil {
					logger.Error("error generating lock key", map[string]interface{}{
						"error": err.Error(),
					})
					continue
				}
				*listLockedKey = append(*listLockedKey, lockKey)
				if _, err := version.IncrementCacheVersion(ctx, c.RedisClient, c.Tracer, c.Logger, c.Semaphore, c.VersionGenerator, session, prefix, namespace, c.VersionExpire); err != nil {
					return err
				}
			}
		}
	} else {
		// Increment cache version for standard keys
		for _, prefix := range listPrefix {
			logger.Debug("clean cache", map[string]interface{}{
				"prefix":    prefix,
				"namespace": key,
			})
			lockKey, err := c.LockGenerator(map[string]string{
				"prefix":    prefix,
				"namespace": key,
			})
			if err != nil {
				logger.Error("error generating lock key", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}
			*listLockedKey = append(*listLockedKey, lockKey)
			if _, err := version.IncrementCacheVersion(ctx, c.RedisClient, c.Tracer, c.Logger, c.Semaphore, c.VersionGenerator, session, prefix, key, c.VersionExpire); err != nil {
				return err
			}
		}
	}

	// If `execute` is true, commit the pipeline
	if execute {
		logger.Debug("commit session cache", nil)
		// acquire lock
		locks, err := lock.AcquireMultipleLock(
			ctx, c.RedisClient, *listLockedKey,
			c.LockDuration, c.LockInterval, true, 5*time.Second,
		)
		if err != nil {
			logger.Error("error acquiring lock", map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}
		defer func() {
			err := lock.ReleaseMultipleLock(ctx, c.RedisClient, locks)
			if err != nil {
				logger.Error("failed to release locks", map[string]interface{}{
					"error": err.Error(),
				})
				return
			}
			logger.Debug("multiple locks released", nil)
		}()
		if _, err := session.Exec(ctx); err != nil {
			return fmt.Errorf("failed to execute redis pipeline: %v", err)
		}
	}

	return nil
}
