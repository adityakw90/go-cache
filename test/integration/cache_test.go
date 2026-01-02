package integration

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/cache"
	"github.com/adityakw90/go-cache/internal/key"
	testutil "github.com/adityakw90/go-cache/test/util"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCache(t *testing.T) (*cache.Cache, *redis.Client) {
	client := testutil.CreateTestRedisClient(t)
	c, err := cache.NewCache(client, cache.Options{
		KeyPrefix:           "test",
		ExpireDefault:       1 * time.Hour,
		VersionExpire:       24 * time.Hour,
		LockDuration:        5 * time.Second,
		LockInterval:        100 * time.Millisecond,
		SemaphoreSize:       10,
		Tracer:              adapter.NewNoOpTracer(),
		Logger:              adapter.NewNoOpLogger(),
		Semaphore:           adapter.NewSemaphore(10),
		KeyGenerator:        key.KeyGenerator,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)
	return c, client
}

func TestCache_Cached_WithoutVersioning(t *testing.T) {
	c, client := setupCache(t)
	defer client.Close()

	ctx := context.Background()
	keyName := "test_function"
	ttl := 1 * time.Minute

	t.Run("cache miss then hit", func(t *testing.T) {
		callCount := 0
		fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
			callCount++
			return map[string]interface{}{
				"result": "test",
				"count":  callCount,
			}, nil
		}

		cachedFn := c.Cached(keyName, ttl, false, "")(fn, nil)

		var resultType map[string]interface{}

		result, err := cachedFn(&resultType, ctx, "arg1", "arg2")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)
		resultMap := result.(map[string]interface{})
		assert.Equal(t, "test", resultMap["result"])

		time.Sleep(200 * time.Millisecond)

		result, err = cachedFn(&resultType, ctx, "arg1", "arg2")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.Equal(t, "test", resultType["result"])
	})

	t.Run("different arguments produce different cache keys", func(t *testing.T) {
		callCount := 0
		fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
			callCount++
			return args[0], nil
		}

		cachedFn := c.Cached(keyName+"_diff", ttl, false, "")(fn, nil)

		var resultType string

		result, err := cachedFn(&resultType, ctx, "arg1")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.Equal(t, "arg1", result.(string))

		result, err = cachedFn(&resultType, ctx, "arg2")
		require.NoError(t, err)
		assert.Equal(t, 2, callCount)
		assert.Equal(t, "arg2", result.(string))
	})

	t.Run("function error bypasses cache", func(t *testing.T) {
		callCount := 0
		fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
			callCount++
			return nil, assert.AnError
		}

		cachedFn := c.Cached(keyName+"_error", ttl, false, "")(fn, nil)

		var resultType interface{}
		_, err := cachedFn(&resultType, ctx, "arg1")
		assert.Error(t, err)
		assert.Equal(t, 1, callCount)
		assert.Nil(t, resultType)
	})
}

func TestCache_Cached_WithVersioning(t *testing.T) {
	c, client := setupCache(t)
	defer client.Close()

	ctx := context.Background()
	keyName := "test_versioned"
	ttl := 1 * time.Minute

	t.Run("versioning creates different cache entries", func(t *testing.T) {
		callCount := 0
		fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
			callCount++
			return map[string]interface{}{
				"result": "v1",
				"count":  callCount,
			}, nil
		}

		cachedFn := c.Cached(keyName, ttl, true, "")(fn, nil)

		var resultType map[string]interface{}

		result, err := cachedFn(&resultType, ctx, "arg1")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)
		resultMap := result.(map[string]interface{})
		assert.Equal(t, "v1", resultMap["result"])

		time.Sleep(200 * time.Millisecond)

		result, err = cachedFn(&resultType, ctx, "arg1")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.Equal(t, "v1", resultType["result"])

		err = c.CleanCache(ctx, keyName, nil, nil, true, nil)
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)

		result, err = cachedFn(&resultType, ctx, "arg1")
		require.NoError(t, err)
		assert.Equal(t, 2, callCount)
		resultMap = result.(map[string]interface{})
		assert.Equal(t, "v1", resultMap["result"])
	})
}

func TestCache_CleanCache(t *testing.T) {
	c, client := setupCache(t)
	defer client.Close()

	ctx := context.Background()
	keyName := "test_clean"
	ttl := 1 * time.Minute

	t.Run("clean cache increments version", func(t *testing.T) {
		callCount := 0
		fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
			callCount++
			return "result", nil
		}

		cachedFn := c.Cached(keyName, ttl, true, "")(fn, nil)

		var resultType string

		result, err := cachedFn(&resultType, ctx, "arg1")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.Equal(t, "result", result.(string))

		time.Sleep(200 * time.Millisecond)

		result, err = cachedFn(&resultType, ctx, "arg1")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.Equal(t, "result", resultType)

		err = c.CleanCache(ctx, keyName, nil, nil, true, nil)
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)

		result, err = cachedFn(&resultType, ctx, "arg1")
		require.NoError(t, err)
		assert.Equal(t, 2, callCount)
		assert.Equal(t, "result", result.(string))
	})

	t.Run("clean cache with custom key function", func(t *testing.T) {
		customKeyFunc, err := key.NewCustomKeyFunction(
			"user_id",
			func(args ...interface{}) string {
				if len(args) > 0 {
					return args[0].(string)
				}
				return "default"
			},
			[]string{"user_id"},
		)
		require.NoError(t, err)

		callCount := 0
		fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
			callCount++
			return "result", nil
		}

		cachedFn := c.Cached(keyName+"_custom", ttl, true, "")(fn, customKeyFunc)

		var resultType string

		_, err = cachedFn(&resultType, ctx, "user123")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)

		time.Sleep(100 * time.Millisecond)

		_, err = cachedFn(&resultType, ctx, "user123")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)

		err = c.CleanCache(ctx, keyName+"_custom", map[string]interface{}{
			"user_id": "user123",
		}, nil, true, nil)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		_, err = cachedFn(&resultType, ctx, "user123")
		require.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})

	t.Run("clean cache without execute", func(t *testing.T) {
		callCount := 0
		fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
			callCount++
			return "result", nil
		}

		cachedFn := c.Cached(keyName+"_no_exec", ttl, true, "")(fn, nil)

		var resultType string

		_, err := cachedFn(&resultType, ctx, "arg1")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)

		session := c.RedisClient.Pipeline()
		var listLockedKey []string

		err = c.CleanCache(ctx, keyName+"_no_exec", nil, session, false, &listLockedKey)
		require.NoError(t, err)
		assert.NotEmpty(t, listLockedKey)

		_, err = session.Exec(ctx)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		_, err = cachedFn(&resultType, ctx, "arg1")
		require.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})
}

func TestCache_TTL(t *testing.T) {
	c, client := setupCache(t)
	defer client.Close()

	ctx := context.Background()
	keyName := "test_ttl"
	shortTTL := 500 * time.Millisecond

	t.Run("cache expires after TTL", func(t *testing.T) {
		callCount := 0
		fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
			callCount++
			return "result", nil
		}

		cachedFn := c.Cached(keyName, shortTTL, false, "")(fn, nil)

		var resultType string

		_, err := cachedFn(&resultType, ctx, "arg1")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)

		time.Sleep(100 * time.Millisecond)

		_, err = cachedFn(&resultType, ctx, "arg1")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)

		time.Sleep(600 * time.Millisecond)

		_, err = cachedFn(&resultType, ctx, "arg1")
		require.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})

	t.Run("dynamic TTL function", func(t *testing.T) {
		callCount := 0
		fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
			callCount++
			return "result", nil
		}

		ttlFunc := func(result interface{}, args ...interface{}) time.Duration {
			if len(args) > 0 && args[0] == "short" {
				return 200 * time.Millisecond
			}
			return 1 * time.Hour
		}

		cachedFn := c.Cached(keyName+"_dynamic", ttlFunc, false, "")(fn, nil)

		var resultType string

		_, err := cachedFn(&resultType, ctx, "short")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)

		time.Sleep(100 * time.Millisecond)

		_, err = cachedFn(&resultType, ctx, "short")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)

		time.Sleep(300 * time.Millisecond)

		_, err = cachedFn(&resultType, ctx, "short")
		require.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})
}
