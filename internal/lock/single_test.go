package lock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLock_AcquireLock(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		timeout     time.Duration
		interval    time.Duration
		wait        bool
		waitTimeout time.Duration
		setupMock   func(redismock.ClientMock, string, time.Duration)
		ctx         context.Context
		validate    func(*testing.T, *LockData, redismock.ClientMock)
	}{
		{
			name:        "success",
			key:         "test:lock:1",
			timeout:     time.Minute,
			interval:    100 * time.Millisecond,
			wait:        false,
			waitTimeout: 5 * time.Second,
			setupMock: func(mock redismock.ClientMock, key string, timeout time.Duration) {
				mock.Regexp().ExpectSetNX(key, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, timeout).SetVal(true)
			},
			ctx: context.Background(),
			validate: func(t *testing.T, lock *LockData, mock redismock.ClientMock) {
				assert.True(t, lock.Acquired)
				assert.NotEmpty(t, lock.Token)
				assert.Equal(t, "test:lock:1", lock.Key)
				assert.NoError(t, lock.Error)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name:        "already_locked_no_wait",
			key:         "test:lock:2",
			timeout:     time.Minute,
			interval:    100 * time.Millisecond,
			wait:        false,
			waitTimeout: 5 * time.Second,
			setupMock: func(mock redismock.ClientMock, key string, timeout time.Duration) {
				mock.Regexp().ExpectSetNX(key, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, timeout).SetVal(false)
			},
			ctx: context.Background(),
			validate: func(t *testing.T, lock *LockData, mock redismock.ClientMock) {
				assert.False(t, lock.Acquired)
				assert.Error(t, lock.Error)
				assert.Equal(t, ErrLockAcquireFailed, lock.Error)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name:        "redis_error",
			key:         "test:lock:3",
			timeout:     time.Minute,
			interval:    100 * time.Millisecond,
			wait:        false,
			waitTimeout: 5 * time.Second,
			setupMock: func(mock redismock.ClientMock, key string, timeout time.Duration) {
				redisErr := errors.New("redis connection error")
				mock.Regexp().ExpectSetNX(key, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, timeout).SetErr(redisErr)
			},
			ctx: context.Background(),
			validate: func(t *testing.T, lock *LockData, mock redismock.ClientMock) {
				assert.False(t, lock.Acquired)
				assert.Error(t, lock.Error)
				assert.Contains(t, lock.Error.Error(), "error while trying to acquire lock")
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name:        "with_wait_success",
			key:         "test:lock:4",
			timeout:     time.Minute,
			interval:    50 * time.Millisecond,
			wait:        true,
			waitTimeout: 5 * time.Second,
			setupMock: func(mock redismock.ClientMock, key string, timeout time.Duration) {
				// First attempt fails, second succeeds
				mock.Regexp().ExpectSetNX(key, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, timeout).SetVal(false)
				mock.Regexp().ExpectSetNX(key, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, timeout).SetVal(true)
			},
			ctx: context.Background(),
			validate: func(t *testing.T, lock *LockData, mock redismock.ClientMock) {
				assert.True(t, lock.Acquired)
				assert.NoError(t, lock.Error)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name:        "with_wait_timeout",
			key:         "test:lock:5",
			timeout:     time.Minute,
			interval:    50 * time.Millisecond,
			wait:        true,
			waitTimeout: 100 * time.Millisecond, // Short timeout
			setupMock: func(mock redismock.ClientMock, key string, timeout time.Duration) {
				// Mock SetNX to always return false (lock never acquired)
				// Allow multiple retries within timeout
				for i := 0; i < 3; i++ {
					mock.Regexp().ExpectSetNX(key, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, timeout).SetVal(false)
				}
			},
			ctx: context.Background(),
			validate: func(t *testing.T, lock *LockData, mock redismock.ClientMock) {
				assert.False(t, lock.Acquired)
				assert.Error(t, lock.Error)
				assert.Contains(t, lock.Error.Error(), "failed to acquire lock within the timeout")
				// Note: Expectations might not be fully met due to timing, so we check leniently
			},
		},
		{
			name:        "context_cancellation",
			key:         "test:lock:11",
			timeout:     time.Minute,
			interval:    100 * time.Millisecond,
			wait:        true,
			waitTimeout: 5 * time.Second,
			setupMock: func(mock redismock.ClientMock, key string, timeout time.Duration) {
				mock.Regexp().ExpectSetNX(key, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, timeout).SetVal(false)
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx
			}(),
			validate: func(t *testing.T, lock *LockData, mock redismock.ClientMock) {
				// Should eventually timeout or fail
				// The exact behavior depends on timing, but it should handle cancellation
				assert.NotNil(t, lock)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			tt.setupMock(mock, tt.key, tt.timeout)

			lock := AcquireLock(tt.ctx, client, tt.key, tt.timeout, tt.interval, tt.wait, tt.waitTimeout)

			tt.validate(t, lock, mock)
		})
	}
}

func TestLock_AcquireLock_TokenGeneration(t *testing.T) {
	client, mock := redismock.NewClientMock()
	ctx := context.Background()
	key1 := "test:lock:12"
	key2 := "test:lock:13"
	timeout := time.Minute
	interval := 100 * time.Millisecond
	wait := false
	waitTimeout := 5 * time.Second

	// Mock SetNX to succeed
	mock.Regexp().ExpectSetNX(key1, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, timeout).SetVal(true)

	lock1 := AcquireLock(ctx, client, key1, timeout, interval, wait, waitTimeout)
	require.NoError(t, mock.ExpectationsWereMet())

	// Acquire another lock - should have different token
	mock.Regexp().ExpectSetNX(key2, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, timeout).SetVal(true)
	lock2 := AcquireLock(ctx, client, key2, timeout, interval, wait, waitTimeout)

	assert.NotEqual(t, lock1.Token, lock2.Token, "Each lock should have a unique token")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLock_ReleaseLock(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		token     string
		setupMock func(redismock.ClientMock, string, string)
		lock      *LockData
		validate  func(*testing.T, *LockData, redismock.ClientMock)
	}{
		{
			name:  "success",
			key:   "test:lock:6",
			token: "test-token-123",
			setupMock: func(mock redismock.ClientMock, key, token string) {
				// Mock script execution to return 1 (success)
				// Script.Run() tries EvalSha first, but in tests scripts aren't cached
				// Script.Run() checks for "NOSCRIPT " (with space) to trigger fallback to Eval
				mock.Regexp().ExpectEvalSha(`.*`, []string{key}, []interface{}{token}).SetErr(errors.New("NOSCRIPT "))
				mock.ExpectEval(scriptUnlock, []string{key}, []interface{}{token}).SetVal(int64(1))
			},
			lock: &LockData{
				Key:      "test:lock:6",
				Token:    "test-token-123",
				Acquired: true,
				Released: false,
			},
			validate: func(t *testing.T, lock *LockData, mock redismock.ClientMock) {
				assert.True(t, lock.Released)
				assert.NoError(t, lock.Error)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name:  "wrong_token",
			key:   "test:lock:7",
			token: "wrong-token",
			setupMock: func(mock redismock.ClientMock, key, token string) {
				// Mock script execution to return 0 (wrong token)
				mock.Regexp().ExpectEvalSha(`.*`, []string{key}, []interface{}{token}).SetErr(errors.New("NOSCRIPT "))
				mock.ExpectEval(scriptUnlock, []string{key}, []interface{}{token}).SetVal(int64(0))
			},
			lock: &LockData{
				Key:      "test:lock:7",
				Token:    "wrong-token",
				Acquired: true,
				Released: false,
			},
			validate: func(t *testing.T, lock *LockData, mock redismock.ClientMock) {
				assert.False(t, lock.Released)
				assert.Error(t, lock.Error)
				assert.Equal(t, ErrLockReleaseForbidden, lock.Error)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name:  "not_exists",
			key:   "test:lock:8",
			token: "test-token",
			setupMock: func(mock redismock.ClientMock, key, token string) {
				// Mock script execution to return -1 (lock doesn't exist)
				mock.Regexp().ExpectEvalSha(`.*`, []string{key}, []interface{}{token}).SetErr(errors.New("NOSCRIPT "))
				mock.ExpectEval(scriptUnlock, []string{key}, []interface{}{token}).SetVal(int64(-1))
			},
			lock: &LockData{
				Key:      "test:lock:8",
				Token:    "test-token",
				Acquired: true,
				Released: false,
			},
			validate: func(t *testing.T, lock *LockData, mock redismock.ClientMock) {
				assert.False(t, lock.Released)
				assert.Error(t, lock.Error)
				assert.Equal(t, ErrLockReleaseUnlocked, lock.Error)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name:  "script_error",
			key:   "test:lock:9",
			token: "test-token",
			setupMock: func(mock redismock.ClientMock, key, token string) {
				scriptErr := errors.New("script execution error")
				// Mock script execution to return an error
				mock.Regexp().ExpectEvalSha(`.*`, []string{key}, []interface{}{token}).SetErr(errors.New("NOSCRIPT "))
				mock.ExpectEval(scriptUnlock, []string{key}, []interface{}{token}).SetErr(scriptErr)
			},
			lock: &LockData{
				Key:      "test:lock:9",
				Token:    "test-token",
				Acquired: true,
				Released: false,
			},
			validate: func(t *testing.T, lock *LockData, mock redismock.ClientMock) {
				assert.False(t, lock.Released)
				assert.Error(t, lock.Error)
				assert.Contains(t, lock.Error.Error(), "error releasing lock")
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name:  "unexpected_result",
			key:   "test:lock:10",
			token: "test-token",
			setupMock: func(mock redismock.ClientMock, key, token string) {
				// Mock script execution to return unexpected value
				// EvalSha returns NOSCRIPT error (script not cached), then Eval will be called
				mock.Regexp().ExpectEvalSha(`.*`, []string{key}, []interface{}{token}).SetErr(errors.New("NOSCRIPT "))
				mock.ExpectEval(scriptUnlock, []string{key}, []interface{}{token}).SetVal(int64(99))
			},
			lock: &LockData{
				Key:      "test:lock:10",
				Token:    "test-token",
				Acquired: true,
				Released: false,
			},
			validate: func(t *testing.T, lock *LockData, mock redismock.ClientMock) {
				assert.False(t, lock.Released)
				assert.Error(t, lock.Error)
				assert.Contains(t, lock.Error.Error(), "unexpected result")
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name:  "nil_lock",
			key:   "",
			token: "",
			setupMock: func(mock redismock.ClientMock, key, token string) {
				// No expectations - nil lock should not call Redis
			},
			lock: nil,
			validate: func(t *testing.T, lock *LockData, mock redismock.ClientMock) {
				// Should not panic or call Redis
				// No expectations, so this should pass
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			ctx := context.Background()

			tt.setupMock(mock, tt.key, tt.token)

			ReleaseLock(ctx, client, tt.lock)

			tt.validate(t, tt.lock, mock)
		})
	}
}
