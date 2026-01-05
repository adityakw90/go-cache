package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// AcquireLock attempts to acquire a distributed lock.
// It uses SETNX (SET if Not eXists) with expiration to create the lock.
// If wait is true, it will retry at the specified interval until waitTimeout.
func AcquireLock(
	ctx context.Context,
	redisClient *redis.Client,
	key string,
	timeout time.Duration,
	interval time.Duration,
	wait bool,
	waitTimeout time.Duration,
) *LockData {
	timeoutCtx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	for {
		lock := &LockData{
			Key:      key,
			Token:    uuid.New().String(),
			Acquired: false,
			Released: false,
		}

		// Try to acquire the lock using SETNX (SET if Not eXists) with expiration
		result, err := redisClient.SetNX(ctx, lock.Key, lock.Token, timeout).Result()
		if err != nil {
			lock.Error = fmt.Errorf("error while trying to acquire lock: %w", err)
			return lock
		}

		if result {
			// Lock acquired successfully
			lock.Acquired = true
			return lock
		}

		if !wait {
			lock.Error = ErrLockAcquireFailed
			return lock
		}

		// Wait for the retry interval before trying again
		select {
		case <-time.After(interval):
			// Retry after the retry interval
		case <-timeoutCtx.Done():
			// Timeout exceeded
			lock.Error = fmt.Errorf("failed to acquire lock within the timeout of %s", waitTimeout)
			return lock
		}
	}
}

// ReleaseLock releases a distributed lock using a Lua script for atomic operation.
// The Lua script ensures that only the lock owner (matching token) can release it.
func ReleaseLock(ctx context.Context, redisClient *redis.Client, lock *LockData) {
	if lock == nil {
		return
	}

	result, err := luaScriptUnlock.Run(ctx, redisClient, []string{lock.Key}, lock.Token).Result()
	if err != nil {
		lock.Error = fmt.Errorf("error releasing lock: %w", err)
		return
	}

	switch result {
	case int64(-1):
		lock.Error = ErrLockReleaseUnlocked // Lock doesn't exist
	case int64(0):
		lock.Error = ErrLockReleaseForbidden // Lock exists but is owned by someone else
	case int64(1):
		lock.Released = true
	default:
		lock.Error = fmt.Errorf("unexpected result from lock release script: %v", result)
	}
}
