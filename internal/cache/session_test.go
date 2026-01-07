package cache

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/adapter"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_GetSession(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, session redis.Pipeliner, mock redismock.ClientMock)
	}{
		{
			name: "returns non-nil pipeline",
			checkFunc: func(t *testing.T, session redis.Pipeliner, mock redismock.ClientMock) {
				assert.NotNil(t, session)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name: "returns pipeline instance",
			checkFunc: func(t *testing.T, session redis.Pipeliner, mock redismock.ClientMock) {
				assert.NotNil(t, session)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name: "multiple calls return different instances",
			checkFunc: func(t *testing.T, session redis.Pipeliner, mock redismock.ClientMock) {
				client, mock2 := redismock.NewClientMock()
				cache, err := NewCache(client, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           adapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				})
				require.NoError(t, err)

				session1 := cache.getSession()
				session2 := cache.getSession()

				// Should be different instances
				assert.NotSame(t, session1, session2)
				assert.NoError(t, mock.ExpectationsWereMet())
				assert.NoError(t, mock2.ExpectationsWereMet())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			defer client.Close()

			cache, err := NewCache(client, Options{
				ExpireDefault:       1 * time.Minute,
				VersionExpire:       1 * time.Hour,
				LockDuration:        5 * time.Second,
				LockInterval:        100 * time.Millisecond,
				StartSpan:           adapter.NoOpStartSpan,
				StartChildSpan:      adapter.NoOpStartChildSpan,
				LogProvider:         adapter.GetNoOpLogger,
				Semaphore:           adapter.NewSemaphore(10),
				KeyGenerator:        key.KeyGenerator,
				KeyVersionGenerator: key.KeyVersionGenerator,
				VersionGenerator:    key.VersionGenerator,
				LockGenerator:       key.LockGenerator,
			})
			require.NoError(t, err)

			session := cache.getSession()
			if tt.checkFunc != nil {
				tt.checkFunc(t, session, mock)
			}
		})
	}
}

func TestCache_GetSession_CanBeUsed(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		ExpireDefault:       1 * time.Minute,
		VersionExpire:       1 * time.Hour,
		LockDuration:        5 * time.Second,
		LockInterval:        100 * time.Millisecond,
		StartSpan:           adapter.NoOpStartSpan,
		StartChildSpan:      adapter.NoOpStartChildSpan,
		LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
		Semaphore:           adapter.NewSemaphore(10),
		KeyGenerator:        key.KeyGenerator,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	session := cache.getSession()
	assert.NotNil(t, session)

	// Set up mock expectations for pipeline operations
	ctx := context.Background()
	mock.ExpectSet("test:key", "value", 0).SetVal("OK")
	mock.ExpectGet("test:key").SetVal("value")

	// Verify we can add commands to the pipeline
	session.Set(ctx, "test:key", "value", 0)
	session.Get(ctx, "test:key")

	// Pipeline should be executable
	cmds, err := session.Exec(ctx)
	require.NoError(t, err)
	assert.NotNil(t, cmds)

	assert.NoError(t, mock.ExpectationsWereMet())
}
