//go:build integration
// +build integration

// Integration tests for lock functionality.
// These tests require a running Redis instance.
package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adityakw90/go-cache/internal/lock"
	testutil "github.com/adityakw90/go-cache/test/util"
)

func TestCache_acquireLock(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		duration    time.Duration
		interval    time.Duration
		wait        bool
		timeout     time.Duration
		prepareFunc func(t *testing.T, cache *Cache, ctx context.Context, key string) func() // cleanup function
		checkFunc   func(t *testing.T, lock *lock.LockData, key string)
	}{
		{
			name:     "acquire lock successfully",
			key:      "test:lock:1",
			duration: time.Minute,
			interval: 100 * time.Millisecond,
			wait:     false,
			timeout:  5 * time.Second,
			prepareFunc: func(t *testing.T, cache *Cache, ctx context.Context, key string) func() {
				// No preparation needed
				return func() {}
			},
			checkFunc: func(t *testing.T, lock *lock.LockData, key string) {
				assert.True(t, lock.Acquired)
				assert.NotEmpty(t, lock.Token)
				assert.Equal(t, key, lock.Key)
				assert.NoError(t, lock.Error)
			},
		},
		{
			name:     "acquire lock when already locked",
			key:      "test:lock:2",
			duration: time.Minute,
			interval: 100 * time.Millisecond,
			wait:     false,
			timeout:  5 * time.Second,
			prepareFunc: func(t *testing.T, cache *Cache, ctx context.Context, key string) func() {
				// Acquire first lock
				lock1 := cache.acquireLock(ctx, key, time.Minute, 100*time.Millisecond, false, 5*time.Second)
				require.True(t, lock1.Acquired)
				// Return cleanup function
				return func() {
					cache.releaseLock(ctx, lock1)
				}
			},
			checkFunc: func(t *testing.T, lock *lock.LockData, key string) {
				assert.False(t, lock.Acquired)
				assert.Error(t, lock.Error)
				assert.Equal(t, ErrLockAcquireFailed, lock.Error)
			},
		},
		{
			name:     "acquire lock with wait",
			key:      "test:lock:3",
			duration: time.Minute,
			interval: 50 * time.Millisecond,
			wait:     true,
			timeout:  5 * time.Second,
			prepareFunc: func(t *testing.T, cache *Cache, ctx context.Context, key string) func() {
				// Acquire first lock
				lock1 := cache.acquireLock(ctx, key, 2*time.Second, 100*time.Millisecond, false, 5*time.Second)
				require.True(t, lock1.Acquired)
				// Start goroutine to release lock after a short delay
				go func() {
					time.Sleep(200 * time.Millisecond)
					cache.releaseLock(ctx, lock1)
				}()
				// Return no-op cleanup (lock1 will be released by goroutine)
				return func() {}
			},
			checkFunc: func(t *testing.T, lock *lock.LockData, key string) {
				assert.True(t, lock.Acquired)
				assert.NoError(t, lock.Error)
			},
		},
		{
			name:     "acquire lock with wait timeout",
			key:      "test:lock:4",
			duration: time.Minute,
			interval: 50 * time.Millisecond,
			wait:     true,
			timeout:  200 * time.Millisecond,
			prepareFunc: func(t *testing.T, cache *Cache, ctx context.Context, key string) func() {
				// Acquire first lock
				lock1 := cache.acquireLock(ctx, key, time.Minute, 100*time.Millisecond, false, 5*time.Second)
				require.True(t, lock1.Acquired)
				// Return cleanup function
				return func() {
					cache.releaseLock(ctx, lock1)
				}
			},
			checkFunc: func(t *testing.T, lock *lock.LockData, key string) {
				assert.False(t, lock.Acquired)
				assert.Error(t, lock.Error)
				assert.Contains(t, lock.Error.Error(), "timeout")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient := testutil.CreateTestRedisClient(t)
			cacheInstance, err := NewCache(redisClient)
			require.NoError(t, err)

			ctx := context.Background()

			// Prepare test state
			cleanup := tt.prepareFunc(t, cacheInstance, ctx, tt.key)
			defer cleanup()

			// Acquire lock
			lock := cacheInstance.acquireLock(ctx, tt.key, tt.duration, tt.interval, tt.wait, tt.timeout)

			// Check results
			if tt.checkFunc != nil {
				tt.checkFunc(t, lock, tt.key)
			}

			// Clean up acquired lock if successful
			if lock.Acquired {
				cacheInstance.releaseLock(ctx, lock)
				assert.True(t, lock.Released)
			}
		})
	}
}

func TestCache_releaseLock(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		prepareFunc func(t *testing.T, cache *Cache, ctx context.Context, key string) (*lock.LockData, func()) // returns lock to release and cleanup function
		checkFunc   func(t *testing.T, cache *Cache, ctx context.Context, lock *lock.LockData, key string)
	}{
		{
			name: "release lock successfully",
			key:  "test:lock:5",
			prepareFunc: func(t *testing.T, cache *Cache, ctx context.Context, key string) (*lock.LockData, func()) {
				// Acquire lock
				lock := cache.acquireLock(ctx, key, time.Minute, 100*time.Millisecond, false, 5*time.Second)
				require.True(t, lock.Acquired)
				// Return lock to release and no-op cleanup
				return lock, func() {}
			},
			checkFunc: func(t *testing.T, cache *Cache, ctx context.Context, lock *lock.LockData, key string) {
				// Release lock
				cache.releaseLock(ctx, lock)
				assert.True(t, lock.Released)
				assert.NoError(t, lock.Error)

				// Verify lock is released by trying to acquire it again
				lock2 := cache.acquireLock(ctx, key, time.Minute, 100*time.Millisecond, false, 5*time.Second)
				assert.True(t, lock2.Acquired)
				cache.releaseLock(ctx, lock2)
			},
		},
		{
			name: "release lock with wrong token",
			key:  "test:lock:6",
			prepareFunc: func(t *testing.T, cache *Cache, ctx context.Context, key string) (*lock.LockData, func()) {
				// Acquire lock
				lock1 := cache.acquireLock(ctx, key, time.Minute, 100*time.Millisecond, false, 5*time.Second)
				require.True(t, lock1.Acquired)
				// Create lock with wrong token
				lock2 := &lock.LockData{
					Key:   key,
					Token: "wrong-token",
				}
				// Return lock2 to release and cleanup function for lock1
				return lock2, func() {
					cache.releaseLock(ctx, lock1)
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, ctx context.Context, lock *lock.LockData, key string) {
				// Try to release with wrong token
				cache.releaseLock(ctx, lock)
				assert.False(t, lock.Released)
				assert.Error(t, lock.Error)
				assert.Equal(t, ErrLockReleaseForbidden, lock.Error)
			},
		},
		{
			name: "release non-existent lock",
			key:  "test:lock:7",
			prepareFunc: func(t *testing.T, cache *Cache, ctx context.Context, key string) (*lock.LockData, func()) {
				// Create lock that doesn't exist
				lock := &lock.LockData{
					Key:   key,
					Token: "some-token",
				}
				// Return lock to release and no-op cleanup
				return lock, func() {}
			},
			checkFunc: func(t *testing.T, cache *Cache, ctx context.Context, lock *lock.LockData, key string) {
				// Try to release non-existent lock
				cache.releaseLock(ctx, lock)
				assert.False(t, lock.Released)
				assert.Error(t, lock.Error)
				assert.Equal(t, ErrLockReleaseUnlocked, lock.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient := testutil.CreateTestRedisClient(t)
			cacheInstance, err := NewCache(redisClient)
			require.NoError(t, err)

			ctx := context.Background()

			// Prepare lock to release
			lock, cleanup := tt.prepareFunc(t, cacheInstance, ctx, tt.key)
			defer cleanup()

			// Check results
			if tt.checkFunc != nil {
				tt.checkFunc(t, cacheInstance, ctx, lock, tt.key)
			}
		})
	}
}

func TestCache_acquireMultipleLock(t *testing.T) {
	tests := []struct {
		name        string
		keys        []string
		prepareFunc func(t *testing.T, cache *Cache, ctx context.Context) func() // cleanup function
		checkFunc   func(t *testing.T, locks []*lock.LockData, err error)
	}{
		{
			name: "acquire multiple locks successfully",
			keys: []string{"test:lock:8", "test:lock:9", "test:lock:10"},
			prepareFunc: func(t *testing.T, cache *Cache, ctx context.Context) func() {
				// No preparation needed
				return func() {}
			},
			checkFunc: func(t *testing.T, locks []*lock.LockData, err error) {
				require.NoError(t, err)
				require.NotNil(t, locks)
				require.Equal(t, 3, len(locks))
				// Verify all locks are acquired
				for _, lock := range locks {
					assert.True(t, lock.Acquired)
					assert.NotEmpty(t, lock.Token)
				}
			},
		},
		{
			name: "acquire multiple locks with partial failure",
			keys: []string{"test:lock:11", "test:lock:12"},
			prepareFunc: func(t *testing.T, cache *Cache, ctx context.Context) func() {
				// Acquire first lock manually to cause conflict
				key1 := "test:lock:11"
				lock1 := cache.acquireLock(ctx, key1, time.Minute, 100*time.Millisecond, false, 5*time.Second)
				require.True(t, lock1.Acquired)
				// Return cleanup function
				return func() {
					cache.releaseLock(ctx, lock1)
				}
			},
			checkFunc: func(t *testing.T, locks []*lock.LockData, err error) {
				assert.Error(t, err)
				assert.Equal(t, ErrLockAcquireFailed, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient := testutil.CreateTestRedisClient(t)
			cacheInstance, err := NewCache(redisClient)
			require.NoError(t, err)

			ctx := context.Background()

			// Prepare test state
			cleanup := tt.prepareFunc(t, cacheInstance, ctx)
			defer cleanup()

			// Acquire multiple locks
			locks, err := cacheInstance.acquireMultipleLock(ctx, tt.keys, time.Minute, 100*time.Millisecond, false, 5*time.Second)

			// Check results
			if tt.checkFunc != nil {
				tt.checkFunc(t, locks, err)
			}

			// Clean up acquired locks if successful
			if err == nil && locks != nil {
				err = cacheInstance.releaseMultipleLock(ctx, locks)
				assert.NoError(t, err)
				for _, lock := range locks {
					assert.True(t, lock.Released)
				}
			}
		})
	}
}

func TestCache_releaseMultipleLock(t *testing.T) {
	tests := []struct {
		name        string
		prepareFunc func(t *testing.T, cache *Cache, ctx context.Context) []*lock.LockData
		checkFunc   func(t *testing.T, locks []*lock.LockData, err error)
	}{
		{
			name: "release multiple locks successfully",
			prepareFunc: func(t *testing.T, cache *Cache, ctx context.Context) []*lock.LockData {
				keys := []string{"test:lock:13", "test:lock:14"}
				// Acquire multiple locks
				locks, err := cache.acquireMultipleLock(ctx, keys, time.Minute, 100*time.Millisecond, false, 5*time.Second)
				require.NoError(t, err)
				return locks
			},
			checkFunc: func(t *testing.T, locks []*lock.LockData, err error) {
				assert.NoError(t, err)
				// Verify all locks are released
				for _, lock := range locks {
					assert.True(t, lock.Released)
					assert.NoError(t, lock.Error)
				}
			},
		},
		{
			name: "release empty locks",
			prepareFunc: func(t *testing.T, cache *Cache, ctx context.Context) []*lock.LockData {
				// Return empty slice
				return []*lock.LockData{}
			},
			checkFunc: func(t *testing.T, locks []*lock.LockData, err error) {
				// Should not error with empty locks
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient := testutil.CreateTestRedisClient(t)
			cacheInstance, err := NewCache(redisClient)
			require.NoError(t, err)

			ctx := context.Background()

			// Prepare locks
			locks := tt.prepareFunc(t, cacheInstance, ctx)

			// Release all locks
			err = cacheInstance.releaseMultipleLock(ctx, locks)

			// Check results
			if tt.checkFunc != nil {
				tt.checkFunc(t, locks, err)
			}
		})
	}
}

func TestLockData_Fields(t *testing.T) {
	tests := []struct {
		name      string
		lock      *lock.LockData
		checkFunc func(t *testing.T, lock *lock.LockData)
	}{
		{
			name: "lock data fields",
			lock: &lock.LockData{
				Key:      "test:key",
				Token:    "test-token",
				Acquired: true,
				Released: false,
				Error:    nil,
			},
			checkFunc: func(t *testing.T, lock *lock.LockData) {
				assert.Equal(t, "test:key", lock.Key)
				assert.Equal(t, "test-token", lock.Token)
				assert.True(t, lock.Acquired)
				assert.False(t, lock.Released)
				assert.NoError(t, lock.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.checkFunc != nil {
				tt.checkFunc(t, tt.lock)
			}
		})
	}
}

// Integration Tests

func TestCache_ConcurrentLockAcquisition(t *testing.T) {
	redisClient := testutil.CreateTestRedisClient(t)
	cacheInstance, err := NewCache(redisClient)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:lock:concurrent:1"
	numGoroutines := 10
	acquiredCount := 0
	var mu sync.Mutex

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines trying to acquire the same lock
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			lock := cacheInstance.acquireLock(ctx, key, time.Minute, 50*time.Millisecond, false, 2*time.Second)
			if lock.Acquired {
				mu.Lock()
				acquiredCount++
				mu.Unlock()
				// Hold the lock for a short time
				time.Sleep(100 * time.Millisecond)
				cacheInstance.releaseLock(ctx, lock)
			}
		}(i)
	}

	wg.Wait()

	// Only one goroutine should have acquired the lock
	assert.Equal(t, 1, acquiredCount)
}

func TestCache_ConcurrentLockAcquisitionWithWait(t *testing.T) {
	redisClient := testutil.CreateTestRedisClient(t)
	cacheInstance, err := NewCache(redisClient)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:lock:concurrent:wait:1"
	numGoroutines := 5
	acquiredLocks := make([]*lock.LockData, 0, numGoroutines)
	var mu sync.Mutex

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines trying to acquire the same lock with wait
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			lock := cacheInstance.acquireLock(ctx, key, 500*time.Millisecond, 50*time.Millisecond, true, 5*time.Second)
			if lock.Acquired {
				mu.Lock()
				acquiredLocks = append(acquiredLocks, lock)
				mu.Unlock()
				// Hold the lock for a short time
				time.Sleep(200 * time.Millisecond)
				cacheInstance.releaseLock(ctx, lock)
			}
		}(i)
	}

	wg.Wait()

	// All goroutines should eventually acquire the lock (sequentially)
	assert.Equal(t, numGoroutines, len(acquiredLocks))
	for _, lock := range acquiredLocks {
		assert.True(t, lock.Acquired)
		assert.True(t, lock.Released)
	}
}

func TestCache_LockExpiration(t *testing.T) {
	redisClient := testutil.CreateTestRedisClient(t)
	cacheInstance, err := NewCache(redisClient)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:lock:expiration:1"

	// Acquire lock with short expiration
	lock1 := cacheInstance.acquireLock(ctx, key, 500*time.Millisecond, 100*time.Millisecond, false, 5*time.Second)
	require.True(t, lock1.Acquired)

	// Wait for lock to expire
	time.Sleep(600 * time.Millisecond)

	// Should be able to acquire the lock again after expiration
	lock2 := cacheInstance.acquireLock(ctx, key, time.Minute, 100*time.Millisecond, false, 5*time.Second)
	assert.True(t, lock2.Acquired)
	assert.NotEqual(t, lock1.Token, lock2.Token)

	// Clean up
	cacheInstance.releaseLock(ctx, lock2)
}

func TestCache_ConcurrentMultipleKeys(t *testing.T) {
	redisClient := testutil.CreateTestRedisClient(t)
	cacheInstance, err := NewCache(redisClient)
	require.NoError(t, err)

	ctx := context.Background()
	numKeys := 5
	numGoroutinesPerKey := 3
	keys := make([]string, numKeys)
	for i := 0; i < numKeys; i++ {
		keys[i] = fmt.Sprintf("test:lock:multi:%d", i)
	}

	var wg sync.WaitGroup
	acquiredCounts := make(map[string]int)
	var mu sync.Mutex

	// Start goroutines for each key
	for _, key := range keys {
		for i := 0; i < numGoroutinesPerKey; i++ {
			wg.Add(1)
			go func(k string) {
				defer wg.Done()
				lock := cacheInstance.acquireLock(ctx, k, time.Minute, 50*time.Millisecond, false, 2*time.Second)
				if lock.Acquired {
					mu.Lock()
					acquiredCounts[k]++
					mu.Unlock()
					time.Sleep(50 * time.Millisecond)
					cacheInstance.releaseLock(ctx, lock)
				}
			}(key)
		}
	}

	wg.Wait()

	// Each key should have been acquired exactly once
	for _, key := range keys {
		assert.Equal(t, 1, acquiredCounts[key], "key %s should be acquired once", key)
	}
}

func TestCache_StressTestMultipleLocks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	redisClient := testutil.CreateTestRedisClient(t)
	cacheInstance, err := NewCache(redisClient)
	require.NoError(t, err)

	ctx := context.Background()
	numLocks := 20
	keys := make([]string, numLocks)
	for i := 0; i < numLocks; i++ {
		keys[i] = fmt.Sprintf("test:lock:stress:%d", i)
	}

	// Acquire all locks
	locks, err := cacheInstance.acquireMultipleLock(ctx, keys, time.Minute, 100*time.Millisecond, false, 5*time.Second)
	require.NoError(t, err)
	require.Equal(t, numLocks, len(locks))

	// Verify all are acquired
	for _, lock := range locks {
		assert.True(t, lock.Acquired)
	}

	// Release all locks
	err = cacheInstance.releaseMultipleLock(ctx, locks)
	assert.NoError(t, err)

	// Verify all are released
	for _, lock := range locks {
		assert.True(t, lock.Released)
	}
}

func TestCache_SequentialLockAcquisition(t *testing.T) {
	redisClient := testutil.CreateTestRedisClient(t)
	cacheInstance, err := NewCache(redisClient)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:lock:sequential:1"
	numIterations := 10

	// Sequentially acquire and release the lock multiple times
	for i := 0; i < numIterations; i++ {
		lock := cacheInstance.acquireLock(ctx, key, time.Minute, 100*time.Millisecond, false, 5*time.Second)
		assert.True(t, lock.Acquired, "iteration %d should acquire lock", i)
		assert.NotEmpty(t, lock.Token, "iteration %d should have token", i)

		cacheInstance.releaseLock(ctx, lock)
		assert.True(t, lock.Released, "iteration %d should release lock", i)
	}
}

func TestCache_ConcurrentMultipleLockAcquisition(t *testing.T) {
	redisClient := testutil.CreateTestRedisClient(t)
	cacheInstance, err := NewCache(redisClient)
	require.NoError(t, err)

	ctx := context.Background()
	numGoroutines := 5
	keys := []string{"test:lock:multi:concurrent:1", "test:lock:multi:concurrent:2", "test:lock:multi:concurrent:3"}
	successCount := 0
	var mu sync.Mutex

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines trying to acquire the same set of locks
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			locks, err := cacheInstance.acquireMultipleLock(ctx, keys, time.Minute, 50*time.Millisecond, true, 5*time.Second)
			if err == nil && len(locks) == len(keys) {
				mu.Lock()
				successCount++
				mu.Unlock()
				// Hold locks for a short time
				time.Sleep(100 * time.Millisecond)
				cacheInstance.releaseMultipleLock(ctx, locks)
			}
		}(i)
	}

	wg.Wait()

	// All goroutines should eventually acquire all locks (sequentially)
	assert.Equal(t, numGoroutines, successCount)
}

func TestCache_LockContentionWithWait(t *testing.T) {
	redisClient := testutil.CreateTestRedisClient(t)
	cacheInstance, err := NewCache(redisClient)
	require.NoError(t, err)

	ctx := context.Background()
	key := "test:lock:contention:1"
	numContenders := 8
	acquiredOrder := make([]int, 0, numContenders)
	var mu sync.Mutex

	var wg sync.WaitGroup
	wg.Add(numContenders)

	// Start multiple goroutines competing for the same lock
	for i := 0; i < numContenders; i++ {
		go func(id int) {
			defer wg.Done()
			lock := cacheInstance.acquireLock(ctx, key, 300*time.Millisecond, 30*time.Millisecond, true, 10*time.Second)
			if lock.Acquired {
				mu.Lock()
				acquiredOrder = append(acquiredOrder, id)
				mu.Unlock()
				// Hold lock briefly
				time.Sleep(100 * time.Millisecond)
				cacheInstance.releaseLock(ctx, lock)
			}
		}(i)
	}

	wg.Wait()

	// All contenders should have acquired the lock
	assert.Equal(t, numContenders, len(acquiredOrder))
}

func TestCache_MixedSingleAndMultipleLocks(t *testing.T) {
	redisClient := testutil.CreateTestRedisClient(t)
	cacheInstance, err := NewCache(redisClient)
	require.NoError(t, err)

	ctx := context.Background()

	// Acquire a single lock
	key1 := "test:lock:mixed:1"
	lock1 := cacheInstance.acquireLock(ctx, key1, time.Minute, 100*time.Millisecond, false, 5*time.Second)
	require.True(t, lock1.Acquired)

	// Acquire multiple locks (different keys)
	keys := []string{"test:lock:mixed:2", "test:lock:mixed:3"}
	locks, err := cacheInstance.acquireMultipleLock(ctx, keys, time.Minute, 100*time.Millisecond, false, 5*time.Second)
	require.NoError(t, err)
	require.Equal(t, len(keys), len(locks))

	// Verify all locks are acquired
	assert.True(t, lock1.Acquired)
	for _, lock := range locks {
		assert.True(t, lock.Acquired)
	}

	// Release all locks
	cacheInstance.releaseLock(ctx, lock1)
	err = cacheInstance.releaseMultipleLock(ctx, locks)
	assert.NoError(t, err)

	// Verify all are released
	assert.True(t, lock1.Released)
	for _, lock := range locks {
		assert.True(t, lock.Released)
	}
}
