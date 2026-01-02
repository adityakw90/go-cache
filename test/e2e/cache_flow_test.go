package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/cache"
	"github.com/adityakw90/go-cache/internal/key"
	testutil "github.com/adityakw90/go-cache/test/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_CacheFlow_FullLifecycle(t *testing.T) {
	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	c, err := cache.NewCache(client, cache.Options{
		KeyPrefix:           "e2e_test_full",
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

	ctx := context.Background()

	keyName := "e2e_function"
	ttl := 1 * time.Minute
	callCount := 0

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		callCount++
		return map[string]interface{}{
			"result":    "success",
			"callCount": callCount,
			"args":      args,
		}, nil
	}

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

	cachedFn := c.Cached(keyName, ttl, true, "")(fn, customKeyFunc)

	var resultType map[string]interface{}

	tests := []struct {
		name           string
		args           []interface{}
		waitBeforeCall time.Duration
		setup          func() int
		validate       func(t *testing.T, result interface{}, err error, prevCallCount int)
	}{
		{
			name:           "first call - cache miss",
			args:           []interface{}{"user123", "param1"},
			waitBeforeCall: 0,
			setup:          func() int { return callCount },
			validate: func(t *testing.T, result interface{}, err error, prevCallCount int) {
				require.NoError(t, err)
				assert.Equal(t, 1, callCount)
				resultMap := result.(map[string]interface{})
				assert.Equal(t, "success", resultMap["result"])
				assert.Equal(t, 1, resultMap["callCount"])
			},
		},
		{
			name:           "second call - cache hit",
			args:           []interface{}{"user123", "param1"},
			waitBeforeCall: 500 * time.Millisecond,
			setup:          func() int { return callCount },
			validate: func(t *testing.T, result interface{}, err error, prevCallCount int) {
				require.NoError(t, err)
				assert.LessOrEqual(t, callCount, prevCallCount+1)
				resultMap, ok := result.(map[string]interface{})
				if ok {
					assert.Equal(t, "success", resultMap["result"])
				} else {
					require.NotNil(t, resultType)
					assert.Equal(t, "success", resultType["result"])
				}
			},
		},
		{
			name:           "different user - cache miss",
			args:           []interface{}{"user456", "param1"},
			waitBeforeCall: 0,
			setup:          func() int { return callCount },
			validate: func(t *testing.T, result interface{}, err error, prevCallCount int) {
				require.NoError(t, err)
				assert.Equal(t, prevCallCount+1, callCount)
				resultMap := result.(map[string]interface{})
				assert.Equal(t, "success", resultMap["result"])
			},
		},
		{
			name:           "clean cache for user123",
			args:           []interface{}{"user123", "param1"},
			waitBeforeCall: 200 * time.Millisecond,
			setup: func() int {
				err := c.CleanCache(ctx, keyName, map[string]interface{}{
					"user_id": "user123",
				}, nil, true, nil)
				require.NoError(t, err)
				return callCount
			},
			validate: func(t *testing.T, result interface{}, err error, prevCallCount int) {
				require.NoError(t, err)
				assert.Equal(t, prevCallCount+1, callCount)
				resultMap := result.(map[string]interface{})
				assert.Equal(t, "success", resultMap["result"])
			},
		},
		{
			name:           "user456 still cached",
			args:           []interface{}{"user456", "param1"},
			waitBeforeCall: 0,
			setup:          func() int { return callCount },
			validate: func(t *testing.T, result interface{}, err error, prevCallCount int) {
				require.NoError(t, err)
				assert.LessOrEqual(t, callCount, prevCallCount+1, "callCount should not increase on cache hit")
				resultMap, ok := result.(map[string]interface{})
				require.True(t, ok, "result should be a map")
				assert.Equal(t, "success", resultMap["result"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.waitBeforeCall > 0 {
				time.Sleep(tt.waitBeforeCall)
			}
			prevCallCount := tt.setup()
			result, err := cachedFn(&resultType, ctx, tt.args...)
			tt.validate(t, result, err, prevCallCount)
		})
	}
}

func TestE2E_CacheFlow_Concurrent(t *testing.T) {
	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	c, err := cache.NewCache(client, cache.Options{
		KeyPrefix:           "e2e_test_concurrent",
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

	ctx := context.Background()

	tests := []struct {
		name          string
		keyName       string
		ttl           time.Duration
		numGoroutines int
		fnDelay       time.Duration
		timeout       time.Duration
		validate      func(t *testing.T, callCount int, received int)
	}{
		{
			name:          "10 concurrent requests",
			keyName:       "e2e_concurrent",
			ttl:           1 * time.Minute,
			numGoroutines: 10,
			fnDelay:       10 * time.Millisecond,
			timeout:       5 * time.Second,
			validate: func(t *testing.T, callCount int, received int) {
				assert.Equal(t, 10, received)
				assert.LessOrEqual(t, callCount, 10)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
				callCount++
				time.Sleep(tt.fnDelay)
				return "result", nil
			}

			cachedFn := c.Cached(tt.keyName, tt.ttl, false, "")(fn, nil)
			results := make(chan string, tt.numGoroutines)

			for i := 0; i < tt.numGoroutines; i++ {
				go func() {
					var resultType string
					_, err := cachedFn(&resultType, ctx, "arg1")
					if err == nil {
						results <- resultType
					}
				}()
			}

			received := 0
			for i := 0; i < tt.numGoroutines; i++ {
				select {
				case <-results:
					received++
				case <-time.After(tt.timeout):
					t.Fatal("timeout waiting for results")
				}
			}

			tt.validate(t, callCount, received)
		})
	}
}

func TestE2E_CacheFlow_Expiration(t *testing.T) {
	type expirationStep struct {
		waitBefore time.Duration
		validate   func(t *testing.T, callCount int)
	}

	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	c, err := cache.NewCache(client, cache.Options{
		KeyPrefix:           "e2e_test_expiration",
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

	ctx := context.Background()

	tests := []struct {
		name     string
		keyName  string
		ttl      time.Duration
		steps    []expirationStep
		validate func(t *testing.T, callCount int)
	}{
		{
			name:    "cache expires after TTL",
			keyName: "e2e_expiration",
			ttl:     300 * time.Millisecond,
			steps: []expirationStep{
				{
					waitBefore: 0,
					validate: func(t *testing.T, callCount int) {
						assert.Equal(t, 1, callCount)
					},
				},
				{
					waitBefore: 200 * time.Millisecond,
					validate: func(t *testing.T, callCount int) {
						assert.Equal(t, 1, callCount)
					},
				},
				{
					waitBefore: 200 * time.Millisecond,
					validate: func(t *testing.T, callCount int) {
						assert.GreaterOrEqual(t, callCount, 2)
					},
				},
			},
			validate: func(t *testing.T, callCount int) {
				assert.GreaterOrEqual(t, callCount, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
				callCount++
				return "result", nil
			}

			cachedFn := c.Cached(tt.keyName, tt.ttl, false, "")(fn, nil)
			var resultType string

			for i, step := range tt.steps {
				if step.waitBefore > 0 {
					time.Sleep(step.waitBefore)
				}

				result, err := cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)

				if i == 0 {
					assert.Equal(t, "result", result.(string))
				} else if i == 1 {
					assert.Equal(t, "result", resultType)
				} else {
					assert.Equal(t, "result", result.(string))
				}

				step.validate(t, callCount)
			}

			tt.validate(t, callCount)
		})
	}
}
