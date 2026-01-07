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
// Returns the lock data and an error if acquisition fails.
func AcquireLock(
	ctx context.Context,
	redisClient *redis.Client,
	key string,
	timeout time.Duration,
	interval time.Duration,
	wait bool,
	waitTimeout time.Duration,
) (*LockData, error) {
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
			return lock, fmt.Errorf("error while trying to acquire lock: %w", err)
		}

		if result {
			// Lock acquired successfully
			lock.Acquired = true
			return lock, nil
		}

		if !wait {
			return lock, ErrLockAcquireFailed
		}

		// Wait for the retry interval before trying again
		timer := time.NewTimer(interval)
		defer timer.Stop()
		select {
		case <-timer.C:
			// Retry after the retry interval
		case <-timeoutCtx.Done():
			// Timeout exceeded
			return lock, fmt.Errorf("failed to acquire lock within the timeout of %s", waitTimeout)
		}
	}
}

// ReleaseLock releases a distributed lock using a Lua script for atomic operation.
// The Lua script ensures that only the lock owner (matching token) can release it.
// Returns an error if the release fails.
func ReleaseLock(ctx context.Context, redisClient *redis.Client, lock *LockData) error {
	if lock == nil {
		return nil
	}

	result, err := luaScriptUnlock.Run(ctx, redisClient, []string{lock.Key}, lock.Token).Result()
	if err != nil {
		return fmt.Errorf("error releasing lock: %w", err)
	}

	switch result {
	case int64(-1):
		return ErrLockReleaseUnlocked // Lock doesn't exist
	case int64(0):
		return ErrLockReleaseForbidden // Lock exists but is owned by someone else
	case int64(1):
		lock.Released = true
		return nil
	default:
		return fmt.Errorf("unexpected result from lock release script: %v", result)
	}
}
