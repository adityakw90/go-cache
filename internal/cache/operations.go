package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/adityakw90/go-cache/internal/serialize"
	"github.com/go-redis/redis/v8"
)

// Get retrieves a value from cache by key and deserializes it into resultType.
// Returns an error if the key doesn't exist or if deserialization fails.
func (c *Cache) Get(ctx context.Context, key string, resultType interface{}) error {
	ctx, cacheSpan := c.Tracer.StartSpan(ctx, "cache.Get")
	defer cacheSpan.End()

	logger := c.Logger.WithSpanContext(cacheSpan.SpanContext())

	logger.Debug("cache get", map[string]interface{}{
		"key": key,
	})

	cachedData, err := c.RedisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		logger.Info("cache miss", map[string]interface{}{
			"key": key,
		})
		return fmt.Errorf("cache key not found: %s", key)
	}
	if err != nil {
		logger.Error("failed to get cache", map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to get cache: %w", err)
	}

	logger.Info("cache hit", map[string]interface{}{
		"key":         key,
		"result_type": fmt.Sprintf("%T", resultType),
	})

	err = serialize.Deserialize(cachedData, resultType)
	if err != nil {
		logger.Error("failed to deserialize cache data", map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to deserialize cache data: %w", err)
	}

	return nil
}

// Set stores a value in cache with the specified TTL.
// The value will be serialized before storing.
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	ctx, cacheSpan := c.Tracer.StartSpan(ctx, "cache.Set")
	defer cacheSpan.End()

	logger := c.Logger.WithSpanContext(cacheSpan.SpanContext())

	logger.Debug("cache set", map[string]interface{}{
		"key": key,
		"ttl": ttl.String(),
	})

	serializedData, err := serialize.Serialize(value)
	if err != nil {
		logger.Error("failed to serialize value", map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	err = c.RedisClient.Set(ctx, key, serializedData, ttl).Err()
	if err != nil {
		logger.Error("failed to set cache data", map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		})
		return fmt.Errorf("failed to set cache data: %w", err)
	}

	logger.Debug("cache set successfully", map[string]interface{}{
		"key":         key,
		"ttl":         ttl.String(),
		"result_type": fmt.Sprintf("%T", value),
	})

	return nil
}
