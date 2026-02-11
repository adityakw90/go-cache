package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Constants for lock configuration.
const (
	// maxRetryIntervalFactor determines the maximum retry interval as a fraction
	// of the total wait timeout. This prevents excessively long retry intervals
	// during lock acquisition attempts.
	maxRetryIntervalFactor = 2
)

// AcquireMultipleLock attempts to acquire multiple locks atomically.
// It uses a pipeline to batch SETNX operations for efficient lock acquisition.
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - redisClient: Redis client instance
//   - keys: list of lock keys to acquire
//   - timeout: individual lock timeout duration
//   - interval: retry interval between acquisition attempts
//   - wait: if true, retry until waitTimeout; if false, attempt once
//   - waitTimeout: maximum time to wait for all locks to be acquired
//
// Returns:
//   - slice of acquired LockData for successful locks
//   - error if acquisition fails or times out
func AcquireMultipleLock(
	ctx context.Context,
	redisClient *redis.Client,
	keys []string,
	timeout time.Duration,
	interval time.Duration,
	wait bool,
	waitTimeout time.Duration,
) ([]*LockData, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	var acquiredLocks []*LockData

	// Create initial lock data
	locks := make([]*LockData, len(keys))
	for i, key := range keys {
		locks[i] = &LockData{
			Key:      key,
			Token:    uuid.New().String(),
			Acquired: false,
			Released: false,
		}
	}

	// Retry mechanism with waitTimeout
	retryInterval := interval
	maxInterval := waitTimeout / maxRetryIntervalFactor
	for {
		pipe := redisClient.Pipeline()
		var results []*redis.BoolCmd
		var remainingKeys []string

		// Add SETNX commands for unacquired locks to the pipeline
		for _, lock := range locks {
			if !lock.Acquired {
				result := pipe.SetNX(timeoutCtx, lock.Key, lock.Token, timeout)
				results = append(results, result)
				remainingKeys = append(remainingKeys, lock.Key)
			}
		}

		// If all locks are acquired, return all acquired locks
		if len(remainingKeys) == 0 {
			// Build the return value from all locks that are acquired
			allAcquired := make([]*LockData, 0, len(locks))
			for _, lock := range locks {
				if lock.Acquired {
					allAcquired = append(allAcquired, lock)
				}
			}
			return allAcquired, nil
		}

		// Execute the pipeline
		_, err := pipe.Exec(timeoutCtx)
		if err != nil {
			// Handle any pipeline execution errors
			if len(acquiredLocks) > 0 {
				ReleaseMultipleLock(ctx, redisClient, acquiredLocks)
			}
			return nil, fmt.Errorf("pipeline execution error: %w", err)
		}

		// Process the results of each SETNX command
		resultIndex := 0
		for i := range locks {
			if !locks[i].Acquired {
				if resultIndex >= len(results) {
					break
				}
				acquired, err := results[resultIndex].Result()
				resultIndex++
				if err != nil {
					// Individual lock acquisition failed, continue to next
					continue
				}

				// If the lock was acquired (SETNX returned true), mark it as acquired
				if acquired {
					locks[i].Acquired = true
					acquiredLocks = append(acquiredLocks, locks[i])
				}
			}
		}

		// Check if all locks are now acquired after processing results
		allAcquired := true
		for _, lock := range locks {
			if !lock.Acquired {
				allAcquired = false
				break
			}
		}
		if allAcquired {
			return acquiredLocks, nil
		}

		// If not waiting, exit immediately after the first attempt
		if !wait {
			ReleaseMultipleLock(ctx, redisClient, acquiredLocks)
			return nil, ErrLockAcquireFailed
		}

		// If we're out of time, release acquired locks and return
		select {
		case <-time.After(retryInterval):
			// Retry after the retry interval
			// Increase the interval (exponential backoff)
			retryInterval = time.Duration(float64(retryInterval) * 1.5)
			if retryInterval > maxInterval {
				retryInterval = maxInterval
			}
			continue
		case <-timeoutCtx.Done():
			ReleaseMultipleLock(ctx, redisClient, acquiredLocks)
			return nil, fmt.Errorf("failed to acquire locks within waitTimeout of %s", waitTimeout)
		}

	}
}

// ReleaseMultipleLock releases multiple locks atomically using a pipeline.
// Each lock release uses a Lua script to ensure only the lock owner can release it.
// Parameters:
//   - ctx: context for cancellation and timeouts
//   - redisClient: Redis client instance
//   - locks: slice of LockData to release (nil entries are skipped)
//
// Returns:
//   - error if any lock release fails
//   - nil if all locks released successfully or locks slice is empty
func ReleaseMultipleLock(ctx context.Context, redisClient *redis.Client, locks []*LockData) error {
	if len(locks) == 0 {
		return nil
	}

	// Create a Redis pipeline
	pipe := redisClient.Pipeline()

	// Store the results of each Lua script execution
	var results []*redis.Cmd

	for _, lock := range locks {
		if lock == nil {
			continue
		}
		// Add each Lua script execution to the pipeline
		result := luaScriptUnlock.Run(ctx, pipe, []string{lock.Key}, lock.Token)
		results = append(results, result)
	}

	// Execute the pipeline (send all commands to Redis at once)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("pipeline execution error: %w", err)
	}

	// Process each result after the pipeline execution
	resultIndex := 0
	for _, lock := range locks {
		if lock == nil || resultIndex >= len(results) {
			continue
		}
		result, err := results[resultIndex].Result()
		resultIndex++
		if err != nil {
			return fmt.Errorf("error releasing lock for key %s: %w", lock.Key, err)
		}

		// Process the result from the Lua script
		switch result {
		case int64(-1):
			return ErrLockReleaseUnlocked
		case int64(0):
			return ErrLockReleaseForbidden
		case int64(1):
			lock.Released = true
		default:
			return fmt.Errorf("unexpected result from lock release script for key %s: %v", lock.Key, result)
		}
	}

	return nil
}
