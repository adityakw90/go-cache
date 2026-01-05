package integration

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/cache"
	"github.com/adityakw90/go-cache/internal/key"
	testutil "github.com/adityakw90/go-cache/test/util"
	"github.com/redis/go-redis/v9"
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
	tests := []struct {
		name      string
		keySuffix string
		setupFn   func() (func(context.Context, ...interface{}) (interface{}, error), *int)
		keyFunc   key.CustomKeyFunction
		runTest   func(*testing.T, *cache.Cache, context.Context, string, time.Duration, func(context.Context, ...interface{}) (interface{}, error), *int, key.CustomKeyFunction)
	}{
		{
			name:      "cache miss then hit",
			keySuffix: "",
			setupFn: func() (func(context.Context, ...interface{}) (interface{}, error), *int) {
				callCount := 0
				fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
					callCount++
					return map[string]interface{}{
						"result": "test",
						"count":  callCount,
					}, nil
				}
				return fn, &callCount
			},
			keyFunc: nil,
			runTest: func(t *testing.T, c *cache.Cache, ctx context.Context, keyName string, ttl time.Duration, fn func(context.Context, ...interface{}) (interface{}, error), callCount *int, keyFunc key.CustomKeyFunction) {
				cachedFn := c.Cached(keyName, ttl, false, "")(fn, keyFunc)
				var resultType map[string]interface{}

				result, err := cachedFn(&resultType, ctx, "arg1", "arg2")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)
				resultMap := result.(map[string]interface{})
				assert.Equal(t, "test", resultMap["result"])

				time.Sleep(200 * time.Millisecond)

				_, err = cachedFn(&resultType, ctx, "arg1", "arg2")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)
				assert.Equal(t, "test", resultType["result"])
			},
		},
		{
			name:      "different arguments produce different cache keys",
			keySuffix: "_diff",
			setupFn: func() (func(context.Context, ...interface{}) (interface{}, error), *int) {
				callCount := 0
				fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
					callCount++
					return args[0], nil
				}
				return fn, &callCount
			},
			keyFunc: nil,
			runTest: func(t *testing.T, c *cache.Cache, ctx context.Context, keyName string, ttl time.Duration, fn func(context.Context, ...interface{}) (interface{}, error), callCount *int, keyFunc key.CustomKeyFunction) {
				cachedFn := c.Cached(keyName, ttl, false, "")(fn, keyFunc)
				var resultType string

				result, err := cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)
				assert.Equal(t, "arg1", result.(string))

				result, err = cachedFn(&resultType, ctx, "arg2")
				require.NoError(t, err)
				assert.Equal(t, 2, *callCount)
				assert.Equal(t, "arg2", result.(string))
			},
		},
		{
			name:      "function error bypasses cache",
			keySuffix: "_error",
			setupFn: func() (func(context.Context, ...interface{}) (interface{}, error), *int) {
				callCount := 0
				fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
					callCount++
					return nil, assert.AnError
				}
				return fn, &callCount
			},
			keyFunc: nil,
			runTest: func(t *testing.T, c *cache.Cache, ctx context.Context, keyName string, ttl time.Duration, fn func(context.Context, ...interface{}) (interface{}, error), callCount *int, keyFunc key.CustomKeyFunction) {
				cachedFn := c.Cached(keyName, ttl, false, "")(fn, keyFunc)
				var resultType interface{}

				_, err := cachedFn(&resultType, ctx, "arg1")
				assert.Error(t, err)
				assert.Equal(t, 1, *callCount)
				assert.Nil(t, resultType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, client := setupCache(t)
			defer client.Close()

			ctx := context.Background()
			keyName := "test_function"
			ttl := 1 * time.Minute

			fn, callCount := tt.setupFn()
			tt.runTest(t, c, ctx, keyName+tt.keySuffix, ttl, fn, callCount, tt.keyFunc)
		})
	}
}

func TestCache_Cached_WithVersioning(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func() (func(context.Context, ...interface{}) (interface{}, error), *int)
		runTest func(*testing.T, *cache.Cache, context.Context, string, time.Duration, func(context.Context, ...interface{}) (interface{}, error), *int)
	}{
		{
			name: "versioning creates different cache entries",
			setupFn: func() (func(context.Context, ...interface{}) (interface{}, error), *int) {
				callCount := 0
				fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
					callCount++
					return map[string]interface{}{
						"result": "v1",
						"count":  callCount,
					}, nil
				}
				return fn, &callCount
			},
			runTest: func(t *testing.T, c *cache.Cache, ctx context.Context, keyName string, ttl time.Duration, fn func(context.Context, ...interface{}) (interface{}, error), callCount *int) {
				cachedFn := c.Cached(keyName, ttl, true, "")(fn, nil)
				var resultType map[string]interface{}

				result, err := cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)
				resultMap := result.(map[string]interface{})
				assert.Equal(t, "v1", resultMap["result"])

				time.Sleep(200 * time.Millisecond)

				_, err = cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)
				assert.Equal(t, "v1", resultType["result"])

				err = c.CleanCache(ctx, keyName, nil, nil, true, nil)
				require.NoError(t, err)

				time.Sleep(200 * time.Millisecond)

				result, err = cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, 2, *callCount)
				resultMap = result.(map[string]interface{})
				assert.Equal(t, "v1", resultMap["result"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, client := setupCache(t)
			defer client.Close()

			ctx := context.Background()
			keyName := "test_versioned"
			ttl := 1 * time.Minute

			fn, callCount := tt.setupFn()
			tt.runTest(t, c, ctx, keyName, ttl, fn, callCount)
		})
	}
}

func TestCache_CleanCache(t *testing.T) {
	tests := []struct {
		name      string
		keySuffix string
		setupFn   func() (func(context.Context, ...interface{}) (interface{}, error), *int)
		keyFunc   key.CustomKeyFunction
		runTest   func(*testing.T, *cache.Cache, context.Context, string, time.Duration, func(context.Context, ...interface{}) (interface{}, error), *int, key.CustomKeyFunction)
	}{
		{
			name:      "clean cache increments version",
			keySuffix: "",
			setupFn: func() (func(context.Context, ...interface{}) (interface{}, error), *int) {
				callCount := 0
				fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
					callCount++
					return "result", nil
				}
				return fn, &callCount
			},
			keyFunc: nil,
			runTest: func(t *testing.T, c *cache.Cache, ctx context.Context, keyName string, ttl time.Duration, fn func(context.Context, ...interface{}) (interface{}, error), callCount *int, keyFunc key.CustomKeyFunction) {
				cachedFn := c.Cached(keyName, ttl, true, "")(fn, keyFunc)
				var resultType string

				result, err := cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)
				assert.Equal(t, "result", result.(string))

				time.Sleep(200 * time.Millisecond)

				_, err = cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)
				assert.Equal(t, "result", resultType)

				err = c.CleanCache(ctx, keyName, nil, nil, true, nil)
				require.NoError(t, err)

				time.Sleep(200 * time.Millisecond)

				result, err = cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, 2, *callCount)
				assert.Equal(t, "result", result.(string))
			},
		},
		{
			name:      "clean cache with custom key function",
			keySuffix: "_custom",
			setupFn: func() (func(context.Context, ...interface{}) (interface{}, error), *int) {
				callCount := 0
				fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
					callCount++
					return "result", nil
				}
				return fn, &callCount
			},
			keyFunc: func() key.CustomKeyFunction {
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
				return customKeyFunc
			}(),
			runTest: func(t *testing.T, c *cache.Cache, ctx context.Context, keyName string, ttl time.Duration, fn func(context.Context, ...interface{}) (interface{}, error), callCount *int, keyFunc key.CustomKeyFunction) {
				cachedFn := c.Cached(keyName, ttl, true, "")(fn, keyFunc)
				var resultType string

				_, err := cachedFn(&resultType, ctx, "user123")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)

				time.Sleep(100 * time.Millisecond)

				_, err = cachedFn(&resultType, ctx, "user123")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)

				err = c.CleanCache(ctx, keyName, map[string]interface{}{
					"user_id": "user123",
				}, nil, true, nil)
				require.NoError(t, err)

				time.Sleep(100 * time.Millisecond)

				_, err = cachedFn(&resultType, ctx, "user123")
				require.NoError(t, err)
				assert.Equal(t, 2, *callCount)
			},
		},
		{
			name:      "clean cache without execute",
			keySuffix: "_no_exec",
			setupFn: func() (func(context.Context, ...interface{}) (interface{}, error), *int) {
				callCount := 0
				fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
					callCount++
					return "result", nil
				}
				return fn, &callCount
			},
			keyFunc: nil,
			runTest: func(t *testing.T, c *cache.Cache, ctx context.Context, keyName string, ttl time.Duration, fn func(context.Context, ...interface{}) (interface{}, error), callCount *int, keyFunc key.CustomKeyFunction) {
				cachedFn := c.Cached(keyName, ttl, true, "")(fn, keyFunc)
				var resultType string

				_, err := cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)

				session := c.RedisClient.Pipeline()
				var listLockedKey []string

				err = c.CleanCache(ctx, keyName, nil, session, false, &listLockedKey)
				require.NoError(t, err)
				assert.NotEmpty(t, listLockedKey)

				_, err = session.Exec(ctx)
				require.NoError(t, err)

				time.Sleep(100 * time.Millisecond)

				_, err = cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, 2, *callCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, client := setupCache(t)
			defer client.Close()

			ctx := context.Background()
			keyName := "test_clean"
			ttl := 1 * time.Minute

			fn, callCount := tt.setupFn()
			tt.runTest(t, c, ctx, keyName+tt.keySuffix, ttl, fn, callCount, tt.keyFunc)
		})
	}
}

func TestCache_TTL(t *testing.T) {
	tests := []struct {
		name      string
		keySuffix string
		ttl       interface{}
		setupFn   func() (func(context.Context, ...interface{}) (interface{}, error), *int)
		runTest   func(*testing.T, *cache.Cache, context.Context, string, interface{}, func(context.Context, ...interface{}) (interface{}, error), *int)
	}{
		{
			name:      "cache expires after TTL",
			keySuffix: "",
			ttl:       500 * time.Millisecond,
			setupFn: func() (func(context.Context, ...interface{}) (interface{}, error), *int) {
				callCount := 0
				fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
					callCount++
					return "result", nil
				}
				return fn, &callCount
			},
			runTest: func(t *testing.T, c *cache.Cache, ctx context.Context, keyName string, ttl interface{}, fn func(context.Context, ...interface{}) (interface{}, error), callCount *int) {
				cachedFn := c.Cached(keyName, ttl, false, "")(fn, nil)
				var resultType string

				_, err := cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)

				time.Sleep(100 * time.Millisecond)

				_, err = cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)

				time.Sleep(600 * time.Millisecond)

				_, err = cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, 2, *callCount)
			},
		},
		{
			name:      "dynamic TTL function",
			keySuffix: "_dynamic",
			ttl: func(result interface{}, args ...interface{}) time.Duration {
				if len(args) > 0 && args[0] == "short" {
					return 200 * time.Millisecond
				}
				return 1 * time.Hour
			},
			setupFn: func() (func(context.Context, ...interface{}) (interface{}, error), *int) {
				callCount := 0
				fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
					callCount++
					return "result", nil
				}
				return fn, &callCount
			},
			runTest: func(t *testing.T, c *cache.Cache, ctx context.Context, keyName string, ttl interface{}, fn func(context.Context, ...interface{}) (interface{}, error), callCount *int) {
				cachedFn := c.Cached(keyName, ttl, false, "")(fn, nil)
				var resultType string

				_, err := cachedFn(&resultType, ctx, "short")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)

				time.Sleep(100 * time.Millisecond)

				_, err = cachedFn(&resultType, ctx, "short")
				require.NoError(t, err)
				assert.Equal(t, 1, *callCount)

				time.Sleep(300 * time.Millisecond)

				_, err = cachedFn(&resultType, ctx, "short")
				require.NoError(t, err)
				assert.Equal(t, 2, *callCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, client := setupCache(t)
			defer client.Close()

			ctx := context.Background()
			keyName := "test_ttl"

			fn, callCount := tt.setupFn()
			tt.runTest(t, c, ctx, keyName+tt.keySuffix, tt.ttl, fn, callCount)
		})
	}
}
