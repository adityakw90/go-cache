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
	timeout := 5 * time.Second
	interval := 100 * time.Millisecond

	t.Run("acquire and release lock", func(t *testing.T) {
		key := "test:lock:single:" + fmt.Sprintf("%d", time.Now().UnixNano())
		lockData, err := lock.AcquireLock(ctx, client, key, timeout, interval, false, 1*time.Second)
		require.NoError(t, err)
		require.NotNil(t, lockData)
		assert.True(t, lockData.Acquired)
		assert.NotEmpty(t, lockData.Token)

		err = lock.ReleaseLock(ctx, client, lockData)
		require.NoError(t, err)
		assert.True(t, lockData.Released)
	})

	t.Run("acquire lock twice without wait", func(t *testing.T) {
		key := "test:lock:twice:" + fmt.Sprintf("%d", time.Now().UnixNano())
		lockData1, err := lock.AcquireLock(ctx, client, key, timeout, interval, false, 1*time.Second)
		require.NoError(t, err)
		require.NotNil(t, lockData1)
		assert.True(t, lockData1.Acquired)

		// Second attempt should fail to acquire (wait=false)
		lockData2, err := lock.AcquireLock(ctx, client, key, timeout, interval, false, 1*time.Second)
		require.Error(t, err) // Expect error when lock already exists
		assert.Equal(t, lock.ErrLockAcquireFailed, err)
		require.NotNil(t, lockData2) // LockData is still returned with Acquired=false
		assert.False(t, lockData2.Acquired)

		err = lock.ReleaseLock(ctx, client, lockData1)
		require.NoError(t, err)
		assert.True(t, lockData1.Released)
	})

	t.Run("acquire lock with wait", func(t *testing.T) {
		key := "test:lock:wait:" + fmt.Sprintf("%d", time.Now().UnixNano())
		lockData1, err := lock.AcquireLock(ctx, client, key, timeout, interval, false, 1*time.Second)
		require.NoError(t, err)
		require.NotNil(t, lockData1)
		assert.True(t, lockData1.Acquired)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(200 * time.Millisecond)
			err := lock.ReleaseLock(ctx, client, lockData1)
			require.NoError(t, err)
		}()

		// Second attempt should succeed (wait=true and lock is released within timeout)
		lockData2, err := lock.AcquireLock(ctx, client, key, timeout, interval, true, 2*time.Second)
		require.NoError(t, err)
		require.NotNil(t, lockData2)
		assert.True(t, lockData2.Acquired)

		err = lock.ReleaseLock(ctx, client, lockData2)
		require.NoError(t, err)
		assert.True(t, lockData2.Released)

		wg.Wait()
	})

	t.Run("release unlocked lock", func(t *testing.T) {
		// This test uses a key that should not exist in Redis
		uniqueKey := "test:lock:nonexistent:" + fmt.Sprintf("%d", time.Now().UnixNano())
		lockData := &lock.LockData{
			Key:      uniqueKey,
			Token:    "invalid-token",
			Acquired: false,
		}

		err := lock.ReleaseLock(ctx, client, lockData)
		assert.Equal(t, lock.ErrLockReleaseUnlocked, err)
	})

	t.Run("release lock with wrong token", func(t *testing.T) {
		key := "test:lock:wrongtoken:" + fmt.Sprintf("%d", time.Now().UnixNano())
		lockData1, err := lock.AcquireLock(ctx, client, key, timeout, interval, false, 1*time.Second)
		require.NoError(t, err)
		require.NotNil(t, lockData1)
		assert.True(t, lockData1.Acquired)

		// Try to release with wrong token
		lockData2 := &lock.LockData{
			Key:      key,
			Token:    "wrong-token",
			Acquired: false,
		}
		err = lock.ReleaseLock(ctx, client, lockData2)
		assert.Error(t, err)
		assert.Equal(t, lock.ErrLockReleaseForbidden, err)

		// Release with correct token should succeed
		err = lock.ReleaseLock(ctx, client, lockData1)
		require.NoError(t, err)
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
		prefix := "test:lock:multi:" + fmt.Sprintf("%d", time.Now().UnixNano())
		keys := []string{prefix + ":1", prefix + ":2", prefix + ":3"}

		locks, err := lock.AcquireMultipleLock(ctx, client, keys, timeout, interval, false, waitTimeout)
		require.NoError(t, err)
		require.Len(t, locks, 3)

		for _, lockData := range locks {
			assert.True(t, lockData.Acquired)
			assert.NotEmpty(t, lockData.Token)
		}

		err = lock.ReleaseMultipleLock(ctx, client, locks)
		require.NoError(t, err)

		for _, lockData := range locks {
			assert.True(t, lockData.Released)
		}
	})

	t.Run("acquire multiple locks with some already locked", func(t *testing.T) {
		prefix := "test:lock:multimixed:" + fmt.Sprintf("%d", time.Now().UnixNano())
		keys := []string{prefix + ":4", prefix + ":5"}

		lockData1, err := lock.AcquireLock(ctx, client, keys[0], timeout, interval, false, 1*time.Second)
		require.NoError(t, err)
		require.NotNil(t, lockData1)
		assert.True(t, lockData1.Acquired)

		_, err = lock.AcquireMultipleLock(ctx, client, keys, timeout, interval, false, waitTimeout)
		assert.Error(t, err)
		assert.Equal(t, lock.ErrLockAcquireFailed, err)

		err = lock.ReleaseLock(ctx, client, lockData1)
		require.NoError(t, err)
	})

	t.Run("acquire multiple locks with wait", func(t *testing.T) {
		prefix := "test:lock:multiwait:" + fmt.Sprintf("%d", time.Now().UnixNano())
		keys := []string{prefix + ":6", prefix + ":7"}

		lockData1, err := lock.AcquireLock(ctx, client, keys[0], timeout, interval, false, 1*time.Second)
		require.NoError(t, err)
		require.NotNil(t, lockData1)
		assert.True(t, lockData1.Acquired)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(200 * time.Millisecond)
			err := lock.ReleaseLock(ctx, client, lockData1)
			require.NoError(t, err)
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
		prefix := "test:lock:multitimeout:" + fmt.Sprintf("%d", time.Now().UnixNano())
		keys := []string{prefix + ":8"}

		lockData1, err := lock.AcquireLock(ctx, client, keys[0], 10*time.Second, interval, false, 1*time.Second)
		require.NoError(t, err)
		require.NotNil(t, lockData1)
		assert.True(t, lockData1.Acquired)

		_, err = lock.AcquireMultipleLock(ctx, client, keys, timeout, interval, true, 500*time.Millisecond)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to acquire locks within waitTimeout")

		err = lock.ReleaseLock(ctx, client, lockData1)
		require.NoError(t, err)
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
	timeout := 5 * time.Second
	interval := 100 * time.Millisecond

	t.Run("concurrent lock acquisition", func(t *testing.T) {
		const numGoroutines = 10
		prefix := "test:lock:concurrent:" + fmt.Sprintf("%d", time.Now().UnixNano())
		acquired := make(chan *lock.LockData, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				lockKey := prefix + "_" + fmt.Sprintf("%d", id)
				lockData, err := lock.AcquireLock(ctx, client, lockKey, timeout, interval, true, 2*time.Second)
				if err == nil && lockData.Acquired {
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
			err := lock.ReleaseLock(ctx, client, lockData)
			require.NoError(t, err)
			assert.True(t, lockData.Released)
		}
	})
}
