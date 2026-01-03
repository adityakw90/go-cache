package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/internal/lock"
	testutil "github.com/adityakw90/go-cache/test/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLock_AcquireRelease_Single(t *testing.T) {
	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:lock:single"
	timeout := 5 * time.Second
	interval := 100 * time.Millisecond

	t.Run("acquire and release lock", func(t *testing.T) {
		lockData := lock.AcquireLock(ctx, client, key, timeout, interval, false, 1*time.Second)
		require.NotNil(t, lockData)
		assert.True(t, lockData.Acquired)
		assert.Empty(t, lockData.Error)
		assert.NotEmpty(t, lockData.Token)

		lock.ReleaseLock(ctx, client, lockData)
		assert.True(t, lockData.Released)
		assert.Empty(t, lockData.Error)
	})

	t.Run("acquire lock twice without wait", func(t *testing.T) {
		lockData1 := lock.AcquireLock(ctx, client, key, timeout, interval, false, 1*time.Second)
		require.NotNil(t, lockData1)
		assert.True(t, lockData1.Acquired)

		lockData2 := lock.AcquireLock(ctx, client, key, timeout, interval, false, 1*time.Second)
		require.NotNil(t, lockData2)
		assert.False(t, lockData2.Acquired)
		assert.Equal(t, lock.ErrLockAcquireFailed, lockData2.Error)

		lock.ReleaseLock(ctx, client, lockData1)
		assert.True(t, lockData1.Released)
	})

	t.Run("acquire lock with wait", func(t *testing.T) {
		lockData1 := lock.AcquireLock(ctx, client, key, timeout, interval, false, 1*time.Second)
		require.NotNil(t, lockData1)
		assert.True(t, lockData1.Acquired)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(200 * time.Millisecond)
			lock.ReleaseLock(ctx, client, lockData1)
		}()

		lockData2 := lock.AcquireLock(ctx, client, key, timeout, interval, true, 2*time.Second)
		require.NotNil(t, lockData2)
		assert.True(t, lockData2.Acquired)
		assert.Empty(t, lockData2.Error)

		lock.ReleaseLock(ctx, client, lockData2)
		assert.True(t, lockData2.Released)

		wg.Wait()
	})

	t.Run("release unlocked lock", func(t *testing.T) {
		lockData := &lock.LockData{
			Key:      key,
			Token:    "invalid-token",
			Acquired: false,
		}

		lock.ReleaseLock(ctx, client, lockData)
		assert.Equal(t, lock.ErrLockReleaseUnlocked, lockData.Error)
	})

	t.Run("release lock with wrong token", func(t *testing.T) {
		lockData1 := lock.AcquireLock(ctx, client, key, timeout, interval, false, 1*time.Second)
		require.NotNil(t, lockData1)
		assert.True(t, lockData1.Acquired)

		lockData2 := &lock.LockData{
			Key:      key,
			Token:    "wrong-token",
			Acquired: false,
		}

		lock.ReleaseLock(ctx, client, lockData2)
		assert.Equal(t, lock.ErrLockReleaseForbidden, lockData2.Error)

		lock.ReleaseLock(ctx, client, lockData1)
		assert.True(t, lockData1.Released)
	})
}

func TestLock_AcquireRelease_Multiple(t *testing.T) {
	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	ctx := context.Background()
	timeout := 5 * time.Second
	interval := 100 * time.Millisecond
	waitTimeout := 2 * time.Second

	t.Run("acquire multiple locks", func(t *testing.T) {
		keys := []string{"test:lock:multi:1", "test:lock:multi:2", "test:lock:multi:3"}

		locks, err := lock.AcquireMultipleLock(ctx, client, keys, timeout, interval, false, waitTimeout)
		require.NoError(t, err)
		require.Len(t, locks, 3)

		for _, lockData := range locks {
			assert.True(t, lockData.Acquired)
			assert.Empty(t, lockData.Error)
			assert.NotEmpty(t, lockData.Token)
		}

		err = lock.ReleaseMultipleLock(ctx, client, locks)
		require.NoError(t, err)

		for _, lockData := range locks {
			assert.True(t, lockData.Released)
		}
	})

	t.Run("acquire multiple locks with some already locked", func(t *testing.T) {
		keys := []string{"test:lock:multi:4", "test:lock:multi:5"}

		lockData1 := lock.AcquireLock(ctx, client, keys[0], timeout, interval, false, 1*time.Second)
		require.NotNil(t, lockData1)
		assert.True(t, lockData1.Acquired)

		_, err := lock.AcquireMultipleLock(ctx, client, keys, timeout, interval, false, waitTimeout)
		assert.Error(t, err)
		assert.Equal(t, lock.ErrLockAcquireFailed, err)

		lock.ReleaseLock(ctx, client, lockData1)
	})

	t.Run("acquire multiple locks with wait", func(t *testing.T) {
		keys := []string{"test:lock:multi:6", "test:lock:multi:7"}

		lockData1 := lock.AcquireLock(ctx, client, keys[0], timeout, interval, false, 1*time.Second)
		require.NotNil(t, lockData1)
		assert.True(t, lockData1.Acquired)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(200 * time.Millisecond)
			lock.ReleaseLock(ctx, client, lockData1)
		}()

		locks, err := lock.AcquireMultipleLock(ctx, client, keys, timeout, interval, true, waitTimeout)
		require.NoError(t, err)
		require.Len(t, locks, 2)

		for _, lockData := range locks {
			assert.True(t, lockData.Acquired)
		}

		err = lock.ReleaseMultipleLock(ctx, client, locks)
		require.NoError(t, err)

		wg.Wait()
	})

	t.Run("acquire multiple locks timeout", func(t *testing.T) {
		keys := []string{"test:lock:multi:8"}

		lockData1 := lock.AcquireLock(ctx, client, keys[0], 10*time.Second, interval, false, 1*time.Second)
		require.NotNil(t, lockData1)
		assert.True(t, lockData1.Acquired)

		_, err := lock.AcquireMultipleLock(ctx, client, keys, timeout, interval, true, 500*time.Millisecond)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to acquire locks within waitTimeout")

		lock.ReleaseLock(ctx, client, lockData1)
	})

	t.Run("release empty locks", func(t *testing.T) {
		err := lock.ReleaseMultipleLock(ctx, client, []*lock.LockData{})
		require.NoError(t, err)
	})

	t.Run("release multiple locks with nil", func(t *testing.T) {
		locks := []*lock.LockData{nil, nil}
		err := lock.ReleaseMultipleLock(ctx, client, locks)
		require.NoError(t, err)
	})
}

func TestLock_Concurrent(t *testing.T) {
	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:lock:concurrent"
	timeout := 5 * time.Second
	interval := 100 * time.Millisecond

	t.Run("concurrent lock acquisition", func(t *testing.T) {
		const numGoroutines = 10
		acquired := make(chan *lock.LockData, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				lockKey := key + "_" + fmt.Sprintf("%d", id)
				lockData := lock.AcquireLock(ctx, client, lockKey, timeout, interval, true, 2*time.Second)
				if lockData.Acquired {
					acquired <- lockData
				}
			}(i)
		}

		var acquiredLocks []*lock.LockData
		for i := 0; i < numGoroutines; i++ {
			select {
			case lockData := <-acquired:
				acquiredLocks = append(acquiredLocks, lockData)
			case <-time.After(5 * time.Second):
				t.Fatal("timeout waiting for lock acquisition")
			}
		}

		wg.Wait()
		close(acquired)

		assert.Len(t, acquiredLocks, numGoroutines)

		for _, lockData := range acquiredLocks {
			lock.ReleaseLock(ctx, client, lockData)
			assert.True(t, lockData.Released)
		}
	})
}
