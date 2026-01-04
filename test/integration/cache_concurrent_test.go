package integration

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/cache"
	"github.com/adityakw90/go-cache/internal/key"
	testutil "github.com/adityakw90/go-cache/test/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_Cached_Concurrent_WriteLock(t *testing.T) {
	tests := []struct {
		name            string
		numGoroutines   int
		fnDelay         time.Duration
		expectedResult  string
		expectCallCount int
	}{
		{
			name:            "10 concurrent requests with write lock",
			numGoroutines:   10,
			fnDelay:         50 * time.Millisecond,
			expectedResult:  "result-arg1",
			expectCallCount: 1,
		},
		{
			name:            "20 concurrent requests with write lock",
			numGoroutines:   20,
			fnDelay:         30 * time.Millisecond,
			expectedResult:  "result-arg1",
			expectCallCount: 1,
		},
		{
			name:            "5 concurrent requests with write lock",
			numGoroutines:   5,
			fnDelay:         100 * time.Millisecond,
			expectedResult:  "result-arg1",
			expectCallCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := testutil.CreateTestRedisClient(t)
			defer client.Close()

			c, err := cache.NewCache(client, cache.Options{
				KeyPrefix:           "test",
				ExpireDefault:       5 * time.Minute,
				VersionExpire:       1 * time.Hour,
				LockDuration:        5 * time.Second,
				LockInterval:        50 * time.Millisecond,
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
			var callCount int64

			fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
				atomic.AddInt64(&callCount, 1)
				time.Sleep(tt.fnDelay)
				return "result-" + args[0].(string), nil
			}

			cachedFunc := c.Cached(
				"testFunc",
				5*time.Minute,
				false,
				"test",
			)(fn, nil)

			results := make(chan string, tt.numGoroutines)
			errors := make(chan error, tt.numGoroutines)
			var wg sync.WaitGroup

			// Launch all concurrent requests
			for i := 0; i < tt.numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					var result string
					res, err := cachedFunc(&result, ctx, "arg1")
					if err != nil {
						errors <- err
						return
					}
					// Use return value since cache miss returns the value directly
					if res != nil {
						if str, ok := res.(string); ok {
							results <- str
						} else {
							results <- result
						}
					} else {
						results <- result
					}
				}()
			}

			wg.Wait()
			close(results)
			close(errors)

			// Check for errors
			for err := range errors {
				require.NoError(t, err, "Unexpected error in concurrent request")
			}

			// Collect results
			collectedResults := make([]string, 0, tt.numGoroutines)
			for result := range results {
				collectedResults = append(collectedResults, result)
			}

			// Wait a bit to ensure all operations complete
			time.Sleep(200 * time.Millisecond)

			// Verify only one function call happened (write lock prevents stampede)
			assert.Equal(t, tt.expectCallCount, int(callCount), "Expected only one function execution with write lock")

			// Verify all requests got results
			assert.Equal(t, tt.numGoroutines, len(collectedResults), "All requests should return results")

			// Verify all results are the same (from cache)
			for _, result := range collectedResults {
				assert.Equal(t, tt.expectedResult, result, "All results should match the cached value")
			}
		})
	}
}

func TestCache_Cached_Concurrent_WriteLock_Timeout(t *testing.T) {
	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	c, err := cache.NewCache(client, cache.Options{
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		VersionExpire:       1 * time.Hour,
		LockDuration:        100 * time.Millisecond,
		LockInterval:        10 * time.Millisecond,
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
	var callCount int64

	// Function that takes longer than lock timeout
	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		atomic.AddInt64(&callCount, 1)
		time.Sleep(500 * time.Millisecond) // Longer than LockDuration * 2
		return "result-" + args[0].(string), nil
	}

	cachedFunc := c.Cached(
		"testFunc",
		5*time.Minute,
		false,
		"test",
	)(fn, nil)

	results := make(chan string, 5)
	errors := make(chan error, 5)
	var wg sync.WaitGroup

	// Launch concurrent requests
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var result string
			res, err := cachedFunc(&result, ctx, "arg1")
			if err != nil {
				errors <- err
				return
			}
			if res != nil {
				if str, ok := res.(string); ok {
					results <- str
				} else {
					results <- result
				}
			} else {
				results <- result
			}
		}()
	}

	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	for err := range errors {
		require.NoError(t, err, "Unexpected error in concurrent request")
	}

	// Collect results
	collectedResults := make([]string, 0, 5)
	for result := range results {
		collectedResults = append(collectedResults, result)
	}

	// Wait for operations to complete
	time.Sleep(1 * time.Second)

	// With timeout, some requests might execute function as fallback
	// But at least one should succeed with cache
	assert.GreaterOrEqual(t, len(collectedResults), 1, "At least some requests should return results")
	assert.GreaterOrEqual(t, int(callCount), 1, "At least one function execution should happen")
}

func TestCache_Cached_Concurrent_WriteLock_SequentialCacheHit(t *testing.T) {
	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	c, err := cache.NewCache(client, cache.Options{
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		VersionExpire:       1 * time.Hour,
		LockDuration:        5 * time.Second,
		LockInterval:        50 * time.Millisecond,
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
	var callCount int64

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		atomic.AddInt64(&callCount, 1)
		return "result-" + args[0].(string), nil
	}

	cachedFunc := c.Cached(
		"testFunc",
		5*time.Minute,
		false,
		"test",
	)(fn, nil)

	// First request - cache miss, should execute function
	var result1 string
	res1, err := cachedFunc(&result1, ctx, "arg1")
	require.NoError(t, err)
	assert.Equal(t, 1, int(callCount), "First request should execute function")
	assert.Equal(t, "result-arg1", res1.(string))

	// Wait for cache to be populated
	time.Sleep(100 * time.Millisecond)

	// Second request - cache hit, should NOT execute function
	var result2 string
	_, err = cachedFunc(&result2, ctx, "arg1")
	require.NoError(t, err)
	assert.Equal(t, 1, int(callCount), "Second request should NOT execute function (cache hit)")
	assert.Equal(t, "result-arg1", result2)

	// Third request - cache hit, should NOT execute function
	var result3 string
	_, err = cachedFunc(&result3, ctx, "arg1")
	require.NoError(t, err)
	assert.Equal(t, 1, int(callCount), "Third request should NOT execute function (cache hit)")
	assert.Equal(t, "result-arg1", result3)
}
