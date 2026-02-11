package integration

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/adapter"
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
				KeyPrefix:                 "test",
				ExpireDefault:             5 * time.Minute,
				VersionExpire:             1 * time.Hour,
				LockDuration:              5 * time.Second,
				LockInterval:              50 * time.Millisecond,
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
			)(fn, nil, true)

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
		KeyPrefix:                 "test",
		ExpireDefault:             5 * time.Minute,
		VersionExpire:             1 * time.Hour,
		LockDuration:              100 * time.Millisecond,
		LockInterval:              10 * time.Millisecond,
		StartSpan:                 adapter.NoOpStartSpan,
		StartChildSpan:            adapter.NoOpStartChildSpan,
		LogProvider:               func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
		Semaphore:                 adapter.NewSemaphore(10),
		KeyGenerator:              key.KeyGenerator,
		KeyVersionGenerator:       key.KeyVersionGenerator,
		KeyHashedVersionGenerator: key.KeyHashedVersionGenerator,
		VersionGenerator:          key.VersionGenerator,
		LockGenerator:             key.LockGenerator,
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
		"testFuncTimeout",
		5*time.Minute,
		false,
		"test",
	)(fn, nil, true)

	// Test with very short lock timeout - should fallback to direct execution
	results := make(chan string, 10)
	errors := make(chan error, 10)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
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

	// With very short lock timeout, we expect multiple calls due to lock timeout fallback
	// The exact count depends on lock acquisition timing
	t.Logf("Function called %d times with short lock timeout", callCount)
}

func TestCache_Cached_Concurrent_Versioning(t *testing.T) {

	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	c, err := cache.NewCache(client, cache.Options{
		KeyPrefix:                 "test",
		ExpireDefault:             5 * time.Minute,
		VersionExpire:             1 * time.Hour,
		LockDuration:              5 * time.Second,
		LockInterval:              50 * time.Millisecond,
		StartSpan:                 adapter.NoOpStartSpan,
		StartChildSpan:            adapter.NoOpStartChildSpan,
		LogProvider:               func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
		Semaphore:                 adapter.NewSemaphore(10),
		KeyGenerator:              key.KeyGenerator,
		KeyVersionGenerator:       key.KeyVersionGenerator,
		KeyHashedVersionGenerator: key.KeyHashedVersionGenerator,
		VersionGenerator:          key.VersionGenerator,
		LockGenerator:             key.LockGenerator,
	})
	require.NoError(t, err)

	ctx := context.Background()
	var callCount int64

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		atomic.AddInt64(&callCount, 1)
		time.Sleep(20 * time.Millisecond)
		return "result-" + args[0].(string), nil
	}

	cachedFunc := c.Cached(
		"testFuncVersion",
		5*time.Minute,
		true, // Enable versioning
		"test",
	)(fn, nil, true)

	const numGoroutines = 5
	results := make(chan string, numGoroutines)
	errors := make(chan error, numGoroutines)
	var wg sync.WaitGroup

	// First batch - should call function once
	for i := 0; i < numGoroutines; i++ {
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
		require.NoError(t, err)
	}

	// Verify only one function call for first batch
	assert.Equal(t, 1, int(callCount), "Expected exactly one function call for first batch")

	// Collect results
	collectedResults := make([]string, 0, numGoroutines)
	for result := range results {
		collectedResults = append(collectedResults, result)
	}

	assert.Equal(t, numGoroutines, len(collectedResults))

	// All results should be the same
	for _, result := range collectedResults {
		assert.Equal(t, "result-arg1", result)
	}
}
