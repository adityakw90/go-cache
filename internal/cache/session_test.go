package cache

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
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
				// Verify it's a pipeline by checking it implements the interface
				assert.NotNil(t, session)
				// Pipeline should be ready to use
				assert.NotNil(t, session)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name: "multiple calls return different instances",
			checkFunc: func(t *testing.T, session redis.Pipeliner, mock redismock.ClientMock) {
				client, mock2 := redismock.NewClientMock()
				cache, err := NewCache(client, Options{})
				require.NoError(t, err)

				session1 := cache.GetSession()
				session2 := cache.GetSession()

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

			cache, err := NewCache(client, Options{})
			require.NoError(t, err)

			session := cache.GetSession()
			if tt.checkFunc != nil {
				tt.checkFunc(t, session, mock)
			}
		})
	}
}

func TestCache_GetSession_CanBeUsed(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{})
	require.NoError(t, err)

	session := cache.GetSession()
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
