package e2e

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/adapter"
	"github.com/adityakw90/go-cache/internal/cache"
	"github.com/adityakw90/go-cache/internal/hash"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/adityakw90/go-cache/internal/serialize"
	testutil "github.com/adityakw90/go-cache/test/util"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_CacheFlow_FullLifecycle(t *testing.T) {
	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	c, err := cache.NewCache(client, cache.Options{
		KeyPrefix:                 "e2e_test_full",
		ExpireDefault:             1 * time.Hour,
		VersionExpire:             24 * time.Hour,
		LockDuration:              5 * time.Second,
		LockInterval:              100 * time.Millisecond,
		StartSpan:                 adapter.NoOpStartSpan,
		StartChildSpan:            adapter.NoOpStartChildSpan,
		LogProvider:               adapter.GetNoOpLogger,
		Semaphore:                 adapter.NewSemaphore(10),
		KeyGenerator:              key.KeyGenerator,
		KeyVersionGenerator:       key.KeyVersionGenerator,
		KeyHashedVersionGenerator: key.KeyHashedVersionGenerator,
		VersionGenerator:          key.VersionGenerator,
		LockGenerator:             key.LockGenerator,
	})
	require.NoError(t, err)

	ctx := context.Background()

	keyName := "e2e_function"
	ttl := 1 * time.Minute
	callCount := 0

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		// result must be pointer to avoid state pollution
		callCount++
		argsListString := []string{
			args[0].(string),
			args[1].(string),
		}
		return &map[string]interface{}{
			"result":    "success",
			"callCount": callCount,
			"args":      argsListString,
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

	cachedHashedVersionedFn := c.Cached(keyName, ttl, true, "")(fn, customKeyFunc, true)
	cachedVersionedFn := c.Cached(keyName, ttl, true, "")(fn, customKeyFunc, false)
	cachedHashedFn := c.Cached(keyName, ttl, false, "")(fn, customKeyFunc, true)
	cachedFn := c.Cached(keyName, ttl, false, "")(fn, customKeyFunc, false)

	tests := []struct {
		name           string
		args           []interface{}
		cachedFn       func(resultType interface{}, ctx context.Context, args ...interface{}) (interface{}, error)
		waitBeforeCall time.Duration
		setup          func()
		validate       func(t *testing.T, result interface{}, err error)
	}{
		{
			name:           "first call - cache miss",
			args:           []interface{}{"user123", "param1"},
			cachedFn:       cachedHashedVersionedFn,
			waitBeforeCall: 0,
			setup:          func() {},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, 1, callCount)
				resultMap := result.(*map[string]interface{})
				assert.Equal(t, "success", (*resultMap)["result"])
				assert.Equal(t, 1, (*resultMap)["callCount"])

				// expected version
				versionKey := "e2e_test_full:user123:version"
				expectedVersion := "1"
				version, err := client.Get(ctx, versionKey).Result()
				require.NoError(t, err)
				assert.Equal(t, expectedVersion, version, "version should be 1")

				// expected cache data and use hardcoded hash value to make sure the test is deterministic
				cacheKey := "e2e_test_full:user123:v1-f4090ccea693930796fba4d3fcba0147.gob"
				expectedCacheData := map[string]interface{}{
					"result":    "success",
					"callCount": 1,
					"args":      []string{"user123", "param1"},
				}
				cacheData, err := client.Get(ctx, cacheKey).Bytes()
				require.NoError(t, err)
				var actualCacheData map[string]interface{}
				err = serialize.Deserialize(cacheData, &actualCacheData)
				require.NoError(t, err)
				assert.Equal(t, expectedCacheData, actualCacheData, "cache data should be equal")
			},
		},
		{
			name:           "second call - cache hit",
			args:           []interface{}{"user123", "param1"},
			cachedFn:       cachedHashedVersionedFn,
			waitBeforeCall: 500 * time.Millisecond,
			setup:          func() {},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, 1, callCount, "callCount should be 1")
				resultMap, ok := result.(*map[string]interface{})
				require.True(t, ok, "result should be a map[string]interface{}")
				assert.Equal(t, "success", (*resultMap)["result"])
			},
		},
		{
			name:           "different user - cache miss",
			args:           []interface{}{"user456", "param1"},
			cachedFn:       cachedHashedVersionedFn,
			waitBeforeCall: 0,
			setup:          func() {},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, 2, callCount)
				resultMap := result.(*map[string]interface{})
				assert.Equal(t, "success", (*resultMap)["result"])
				assert.Equal(t, 2, (*resultMap)["callCount"])

				// expected version
				versionKey := "e2e_test_full:user456:version"
				expectedVersion := "1"
				version, err := client.Get(ctx, versionKey).Result()
				require.NoError(t, err)
				assert.Equal(t, expectedVersion, version, "version should be 1")

				// expected cache data and use hardcoded hash value to make sure the test is deterministic
				cacheKey := "e2e_test_full:user456:v1-585d04d8ffcf93141ca460eec30acbe9.gob"
				expectedCacheData := map[string]interface{}{
					"result":    "success",
					"callCount": 2,
					"args":      []string{"user456", "param1"},
				}
				cacheData, err := client.Get(ctx, cacheKey).Bytes()
				require.NoError(t, err)
				var actualCacheData map[string]interface{}
				err = serialize.Deserialize(cacheData, &actualCacheData)
				require.NoError(t, err)
				assert.Equal(t, expectedCacheData, actualCacheData, "cache data should be equal")
			},
		},
		{
			name:           "clean cache for user123",
			args:           []interface{}{"user123", "param1"},
			cachedFn:       cachedHashedVersionedFn,
			waitBeforeCall: 200 * time.Millisecond,
			setup: func() {
				err := c.CleanCache(ctx, keyName, map[string]interface{}{
					"user_id": "user123",
				}, nil, true, nil)
				require.NoError(t, err)
			},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, 3, callCount)
				resultMap := result.(*map[string]interface{})
				assert.Equal(t, "success", (*resultMap)["result"])
				assert.Equal(t, 3, (*resultMap)["callCount"])

				// expected version
				versionKey := "e2e_test_full:user123:version"
				expectedVersion := "2"
				version, err := client.Get(ctx, versionKey).Result()
				require.NoError(t, err)
				assert.Equal(t, expectedVersion, version, "version should be 1")

				// expected cache data and use hardcoded hash value to make sure the test is deterministic
				cacheKey := "e2e_test_full:user123:v2-f4090ccea693930796fba4d3fcba0147.gob"
				expectedCacheData := map[string]interface{}{
					"result":    "success",
					"callCount": 3,
					"args":      []string{"user123", "param1"},
				}
				cacheData, err := client.Get(ctx, cacheKey).Bytes()
				require.NoError(t, err)
				var actualCacheData map[string]interface{}
				err = serialize.Deserialize(cacheData, &actualCacheData)
				require.NoError(t, err)
				assert.Equal(t, expectedCacheData, actualCacheData, "cache data should be equal")
			},
		},
		{
			name:           "user456 still cached",
			args:           []interface{}{"user456", "param1"},
			cachedFn:       cachedHashedVersionedFn,
			waitBeforeCall: 0,
			setup:          func() {},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, 3, callCount, "callCount should be 3") // total call count
				resultMap, ok := result.(*map[string]interface{})
				require.True(t, ok, "result should be a map")
				assert.Equal(t, "success", (*resultMap)["result"])
				assert.Equal(t, 2, (*resultMap)["callCount"]) // its from cached data
			},
		},
		{
			name:           "user123 not hashed - cache miss",
			args:           []interface{}{"user123", "param1"},
			cachedFn:       cachedVersionedFn,
			waitBeforeCall: 0,
			setup:          func() {},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, 4, callCount)
				resultMap := result.(*map[string]interface{})
				assert.Equal(t, "success", (*resultMap)["result"])
				assert.Equal(t, 4, (*resultMap)["callCount"])

				// expected version
				versionKey := "e2e_test_full:user123:version"
				expectedVersion := "2"
				version, err := client.Get(ctx, versionKey).Result()
				require.NoError(t, err)
				assert.Equal(t, expectedVersion, version, "version should be 2")

				// expected cache data and use hardcoded hash value to make sure the test is deterministic
				cacheKey := "e2e_test_full:user123:v2.gob"
				expectedCacheData := map[string]interface{}{
					"result":    "success",
					"callCount": 4,
					"args":      []string{"user123", "param1"},
				}
				cacheData, err := client.Get(ctx, cacheKey).Bytes()
				require.NoError(t, err)
				var actualCacheData map[string]interface{}
				err = serialize.Deserialize(cacheData, &actualCacheData)
				require.NoError(t, err)
				assert.Equal(t, expectedCacheData, actualCacheData, "cache data should be equal")
			},
		},
		{
			name:           "user123 not hashed - cache hit",
			args:           []interface{}{"user123", "param1"},
			cachedFn:       cachedVersionedFn,
			waitBeforeCall: 500 * time.Millisecond,
			setup:          func() {},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, 4, callCount, "callCount should be 4")
				resultMap, ok := result.(*map[string]interface{})
				require.True(t, ok, "result should be a map[string]interface{}")
				assert.Equal(t, "success", (*resultMap)["result"])
			},
		},
		{
			name:           "user789 not versioned - cache miss",
			args:           []interface{}{"user789", "param1"},
			cachedFn:       cachedHashedFn,
			waitBeforeCall: 0,
			setup:          func() {},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, 5, callCount)
				resultMap := result.(*map[string]interface{})
				assert.Equal(t, "success", (*resultMap)["result"])
				assert.Equal(t, 5, (*resultMap)["callCount"])

				// expected cache data and use hardcoded hash value to make sure the test is deterministic
				cacheKey := "e2e_test_full:user789:v0-ddf19f31c5a9f5783bcfa335ba24f374.gob"
				expectedCacheData := map[string]interface{}{
					"result":    "success",
					"callCount": 5,
					"args":      []string{"user789", "param1"},
				}
				cacheData, err := client.Get(ctx, cacheKey).Bytes()
				require.NoError(t, err)
				var actualCacheData map[string]interface{}
				err = serialize.Deserialize(cacheData, &actualCacheData)
				require.NoError(t, err)
				assert.Equal(t, expectedCacheData, actualCacheData, "cache data should be equal")
			},
		},
		{
			name:           "user789 not versioned - cache hit",
			args:           []interface{}{"user789", "param1"},
			cachedFn:       cachedHashedFn,
			waitBeforeCall: 500 * time.Millisecond,
			setup:          func() {},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, 5, callCount, "callCount should be 5")
				resultMap, ok := result.(*map[string]interface{})
				require.True(t, ok, "result should be a map[string]interface{}")
				assert.Equal(t, "success", (*resultMap)["result"])
			},
		},
		{
			name:           "user123456 not versioned not hashed - cache miss",
			args:           []interface{}{"user123456", "param1"},
			cachedFn:       cachedFn,
			waitBeforeCall: 0,
			setup:          func() {},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, 6, callCount)
				resultMap := result.(*map[string]interface{})
				assert.Equal(t, "success", (*resultMap)["result"])
				assert.Equal(t, 6, (*resultMap)["callCount"])

				// expected cache data and use hardcoded hash value to make sure the test is deterministic
				cacheKey := "e2e_test_full:user123456:v0.gob"
				expectedCacheData := map[string]interface{}{
					"result":    "success",
					"callCount": 6,
					"args":      []string{"user123456", "param1"},
				}
				cacheData, err := client.Get(ctx, cacheKey).Bytes()
				require.NoError(t, err)
				var actualCacheData map[string]interface{}
				err = serialize.Deserialize(cacheData, &actualCacheData)
				require.NoError(t, err)
				assert.Equal(t, expectedCacheData, actualCacheData, "cache data should be equal")
			},
		},
		{
			name:           "user123456 not versioned not hashed - cache hit",
			args:           []interface{}{"user123456", "param1"},
			cachedFn:       cachedFn,
			waitBeforeCall: 500 * time.Millisecond,
			setup:          func() {},
			validate: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				assert.Equal(t, 6, callCount, "callCount should be 6")
				resultMap, ok := result.(*map[string]interface{})
				require.True(t, ok, "result should be a map[string]interface{}")
				assert.Equal(t, "success", (*resultMap)["result"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.waitBeforeCall > 0 {
				time.Sleep(tt.waitBeforeCall)
			}
			tt.setup()
			var result interface{}
			var err error
			// Declare resultType inside each test to avoid state pollution
			var resultType map[string]interface{}
			result, err = tt.cachedFn(&resultType, ctx, tt.args...)
			tt.validate(t, result, err)
		})
	}
}

func TestE2E_CacheFlow_Concurrent(t *testing.T) {
	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	c, err := cache.NewCache(client, cache.Options{
		KeyPrefix:                 "e2e_test_concurrent",
		ExpireDefault:             1 * time.Hour,
		VersionExpire:             24 * time.Hour,
		LockDuration:              5 * time.Second,
		LockInterval:              100 * time.Millisecond,
		LogProvider:               func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
		Semaphore:                 adapter.NewSemaphore(10),
		KeyGenerator:              key.KeyGenerator,
		KeyVersionGenerator:       key.KeyVersionGenerator,
		KeyHashedVersionGenerator: key.KeyHashedVersionGenerator,
		VersionGenerator:          key.VersionGenerator,
		LockGenerator:             key.LockGenerator,
		StartSpan:                 adapter.NoOpStartSpan,
		StartChildSpan:            adapter.NoOpStartChildSpan,
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

			cachedFn := c.Cached(tt.keyName, tt.ttl, false, "")(fn, nil, true)
			results := make(chan string, tt.numGoroutines)
			var wg sync.WaitGroup

			for i := 0; i < tt.numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
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

			wg.Wait()
			close(results)

			tt.validate(t, int(callCount), received)
		})
	}
}

func TestE2E_CacheFlow_Expiration(t *testing.T) {
	type expirationStep struct {
		waitBefore time.Duration
		prepare    func(t *testing.T, redisClient *redis.Client, keyName string, ctx context.Context)
		validate   func(t *testing.T, redisClient *redis.Client, keyName string, callCount int, ctx context.Context)
	}
	tests := []struct {
		name     string
		keyName  string
		ttl      time.Duration
		fnDelay  time.Duration
		steps    []expirationStep
		validate func(t *testing.T, callCount int)
	}{
		{
			name:    "cache expires after TTL",
			keyName: "e2e_expiration",
			ttl:     10 * time.Minute,
			fnDelay: 100 * time.Millisecond,
			steps: []expirationStep{
				{
					waitBefore: 0,
					prepare:    nil,
					validate: func(t *testing.T, redisClient *redis.Client, keyName string, callCount int, ctx context.Context) {
						assert.Equal(t, 1, callCount)
						// verify the cache is set
						var testResult string
						cachedData, err := redisClient.Get(ctx, keyName).Bytes()
						require.NoError(t, err)
						err = serialize.Deserialize(cachedData, &testResult)
						require.NoError(t, err)
						assert.Equal(t, "result", testResult)
					},
				},
				{
					waitBefore: 5 * time.Second,
					prepare: func(t *testing.T, redisClient *redis.Client, keyName string, ctx context.Context) {
						ttlData, err := redisClient.TTL(ctx, keyName).Result()
						require.NoError(t, err)
						assert.Greater(t, ttlData, 1*time.Second)
					},
					validate: func(t *testing.T, redisClient *redis.Client, keyName string, callCount int, ctx context.Context) {
						assert.Equal(t, 1, callCount, "cache should still be valid, callCount should be 1")
					},
				},
				{
					waitBefore: 5 * time.Second,
					prepare: func(t *testing.T, redisClient *redis.Client, keyName string, ctx context.Context) {
						// manually expire the cache by deleting it
						// This simulates cache expiration and ensures next call will be a cache miss
						_ = redisClient.Del(ctx, keyName).Err()
					},
					validate: func(t *testing.T, redisClient *redis.Client, keyName string, callCount int, ctx context.Context) {
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
			client := testutil.CreateTestRedisClient(t)
			defer client.Close()

			c, err := cache.NewCache(client, cache.Options{
				KeyPrefix:                 "e2e_test_expiration",
				ExpireDefault:             1 * time.Hour,
				VersionExpire:             24 * time.Hour,
				LockDuration:              5 * time.Second,
				LockInterval:              100 * time.Millisecond,
				LogProvider:               func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
				Semaphore:                 adapter.NewSemaphore(10),
				KeyGenerator:              key.KeyGenerator,
				KeyVersionGenerator:       key.KeyVersionGenerator,
				KeyHashedVersionGenerator: key.KeyHashedVersionGenerator,
				VersionGenerator:          key.VersionGenerator,
				LockGenerator:             key.LockGenerator,
				StartSpan:                 adapter.NoOpStartSpan,
				StartChildSpan:            adapter.NoOpStartChildSpan,
			})
			require.NoError(t, err)
			ctx := context.Background()
			var callCount int64

			cachedWrapper := c.Cached(tt.keyName, tt.ttl, false, "")
			cachedFunc := cachedWrapper(
				func(ctx context.Context, args ...interface{}) (interface{}, error) {
					atomic.AddInt64(&callCount, 1)
					time.Sleep(tt.fnDelay)
					result := "result"
					return &result, nil
				}, nil, true,
			)
			cacheKey, err := c.KeyHashedVersionGenerator(map[string]string{
				"prefix":    c.KeyPrefix,
				"namespace": tt.keyName,
				"version":   "0", // versioning is disabled
				"key":       hash.CacheKey(tt.keyName, []interface{}{"arg1"}),
			})
			require.NoError(t, err)

			for stepIndex, step := range tt.steps {
				if step.waitBefore > 0 {
					time.Sleep(step.waitBefore)
				}
				var resultType string
				var resultString string
				if step.prepare != nil {
					step.prepare(t, client, cacheKey, ctx)
				}
				result, err := cachedFunc(&resultType, ctx, "arg1")
				require.NoError(t, err)
				fmt.Println("step index: ", stepIndex)
				fmt.Println("result value: ", result)
				fmt.Println("result type:", fmt.Sprintf("%T", result))
				resultString = *result.(*string)
				assert.Equal(t, "result", resultString)
				if step.validate != nil {
					step.validate(t, client, cacheKey, int(callCount), ctx)
				}
			}
			tt.validate(t, int(callCount))

			// Wait for all operations to complete before cleanup
			time.Sleep(200 * time.Millisecond)

			// Cleanup: remove test data to ensure test isolation
			_ = client.Del(ctx, cacheKey).Err()
		})
	}
}
