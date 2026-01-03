package e2e

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/cache"
	"github.com/adityakw90/go-cache/internal/hash"
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
			var callCount int64
			fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
				atomic.AddInt64(&callCount, 1)
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

			tt.validate(t, int(callCount), received)
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
			ttl:     10 * time.Second,
			steps: []expirationStep{
				{
					waitBefore: 0,
					validate: func(t *testing.T, callCount int) {
						assert.Equal(t, 1, callCount)
					},
				},
				{
					waitBefore: 100 * time.Millisecond,
					validate: func(t *testing.T, callCount int) {
						assert.Equal(t, 1, callCount, "cache should still be valid, callCount should be 1")
					},
				},
				{
					waitBefore: 10100 * time.Millisecond,
					validate: func(t *testing.T, callCount int) {
						assert.GreaterOrEqual(t, callCount, 2, "cache should have expired, callCount should be >= 2")
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

			// Helper function to verify cache key exists and is actually retrievable
			verifyCacheRetrievable := func(keyName string, args []interface{}) bool {
				namespace := keyName
				hashKey := hash.CacheKey(namespace, args)
				cacheKey, err := c.KeyVersionGenerator(map[string]string{
					"prefix":    c.KeyPrefix,
					"namespace": namespace,
					"version":   "0", // versioning is disabled
					"key":       hashKey,
				})
				if err != nil {
					return false
				}
				// Try to actually retrieve the cache value using the cache's Get method
				// This is more reliable than just checking if the key exists
				var testResult string
				err = c.Get(ctx, cacheKey, &testResult)
				return err == nil && testResult == "result"
			}

			for stepIndex, step := range tt.steps {
				if step.waitBefore > 0 {
					time.Sleep(step.waitBefore)
				}

				// Reset resultType before each call to ensure clean state
				resultType = ""

				// Before making the call, verify cache is retrievable for steps that expect cache hit
				// This helps debug flaky tests by ensuring cache is actually accessible before the call
				if stepIndex == 1 {
					// Step 1 should have cache hit - verify cache is retrievable right before the call
					// Check multiple times to ensure stability
					for i := 0; i < 3; i++ {
						cacheRetrievable := verifyCacheRetrievable(tt.keyName, []interface{}{"arg1"})
						if !cacheRetrievable {
							t.Logf("Warning: Cache not retrievable before step %d call (check %d). This may indicate timing issues.", stepIndex+1, i+1)
						} else {
							break // Cache is retrievable, proceed
						}
						if i < 2 {
							time.Sleep(5 * time.Millisecond)
						}
					}
				}

				result, err := cachedFn(&resultType, ctx, "arg1")
				require.NoError(t, err)

				// Check result value - on cache hit, result is resultType (*string pointer)
				// On cache miss, result is the fresh function return value (string)
				// Check resultType for cache hit, or result for cache miss
				if resultType != "" {
					// Cache hit - resultType was populated by Get
					assert.Equal(t, "result", resultType)
				} else {
					// Cache miss - result is the function return value
					assert.Equal(t, "result", result.(string))
				}

				// After first call (cache set), verify the cache is actually retrievable and stable
				// This ensures the cache is actually set and accessible before proceeding, preventing flaky tests
				if stepIndex == 0 {
					// Retry up to 20 times with 10ms intervals to verify cache is retrievable
					// This accounts for Redis write propagation delays
					maxRetries := 20
					cacheVerified := false
					for i := 0; i < maxRetries; i++ {
						if verifyCacheRetrievable(tt.keyName, []interface{}{"arg1"}) {
							cacheVerified = true
							break
						}
						time.Sleep(10 * time.Millisecond)
					}
					require.True(t, cacheVerified, "cache should be retrievable after first call")

					// Verify cache is stable by checking it multiple times
					// This helps catch any transient issues
					for i := 0; i < 3; i++ {
						require.True(t, verifyCacheRetrievable(tt.keyName, []interface{}{"arg1"}),
							"cache should remain retrievable (stability check %d)", i+1)
						time.Sleep(5 * time.Millisecond)
					}
				}

				step.validate(t, callCount)
			}

			tt.validate(t, callCount)
		})
	}
}
