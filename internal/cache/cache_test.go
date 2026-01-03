package cache

import (
	"testing"
	"time"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/errs"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_NewCache(t *testing.T) {
	tests := []struct {
		name      string
		wantErr   bool
		setupFunc func(t *testing.T) (*redis.Client, redismock.ClientMock, Options)
		checkFunc func(t *testing.T, cache *Cache, err error)
	}{
		{
			name:    "valid redis client with default options",
			wantErr: false,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					Tracer:    adapter.NewNoOpTracer(),
					Logger:    adapter.NewNoOpLogger(),
					Semaphore: adapter.NewSemaphore(10),
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache)
				assert.NotNil(t, cache.RedisClient)
				assert.NotNil(t, cache.Tracer)
				assert.NotNil(t, cache.Logger)
				assert.NotNil(t, cache.Semaphore)
			},
		},
		{
			name:    "nil redis client",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				_, mock := redismock.NewClientMock()
				return nil, mock, Options{
					Tracer:    adapter.NewNoOpTracer(),
					Logger:    adapter.NewNoOpLogger(),
					Semaphore: adapter.NewSemaphore(10),
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "redisClient", invalidConfigErr.Field())
				assert.Equal(t, "cannot be nil", invalidConfigErr.Message())
			},
		},
		{
			name:    "with custom options",
			wantErr: false,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					KeyPrefix:     "test",
					ExpireDefault: 10 * time.Minute,
					VersionExpire: 1 * time.Hour,
					LockDuration:  5 * time.Second,
					LockInterval:  100 * time.Millisecond,
					Tracer:        adapter.NewNoOpTracer(),
					Logger:        adapter.NewNoOpLogger(),
					Semaphore:     adapter.NewSemaphore(20),
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.Equal(t, "test", cache.KeyPrefix)
				assert.Equal(t, 10*time.Minute, cache.ExpireDefault)
				assert.Equal(t, 1*time.Hour, cache.VersionExpire)
				assert.Equal(t, 5*time.Second, cache.LockDuration)
				assert.Equal(t, 100*time.Millisecond, cache.LockInterval)
			},
		},
		{
			name:    "with custom tracer",
			wantErr: false,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					Tracer:    adapter.NewNoOpTracer(),
					Logger:    adapter.NewNoOpLogger(),
					Semaphore: adapter.NewSemaphore(10),
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.Tracer)
			},
		},
		{
			name:    "with custom logger",
			wantErr: false,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					Tracer:    adapter.NewNoOpTracer(),
					Logger:    adapter.NewNoOpLogger(),
					Semaphore: adapter.NewSemaphore(10),
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.Logger)
			},
		},
		{
			name:    "with custom semaphore",
			wantErr: false,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					Tracer:    adapter.NewNoOpTracer(),
					Logger:    adapter.NewNoOpLogger(),
					Semaphore: adapter.NewSemaphore(10),
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.Semaphore)
			},
		},
		{
			name:    "with key generators",
			wantErr: false,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
					Tracer:              adapter.NewNoOpTracer(),
					Logger:              adapter.NewNoOpLogger(),
					Semaphore:           adapter.NewSemaphore(10),
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.KeyGenerator)
				assert.NotNil(t, cache.KeyVersionGenerator)
				assert.NotNil(t, cache.VersionGenerator)
				assert.NotNil(t, cache.LockGenerator)
			},
		},
		{
			name:    "nil tracer returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					Tracer:    nil,
					Logger:    adapter.NewNoOpLogger(),
					Semaphore: adapter.NewSemaphore(10),
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "tracer", invalidConfigErr.Field())
				assert.Equal(t, "cannot be nil", invalidConfigErr.Message())
			},
		},
		{
			name:    "nil logger returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					Tracer:    adapter.NewNoOpTracer(),
					Logger:    nil,
					Semaphore: adapter.NewSemaphore(10),
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "logger", invalidConfigErr.Field())
				assert.Equal(t, "cannot be nil", invalidConfigErr.Message())
			},
		},
		{
			name:    "nil semaphore returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					Tracer:    adapter.NewNoOpTracer(),
					Logger:    adapter.NewNoOpLogger(),
					Semaphore: nil,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "semaphore", invalidConfigErr.Field())
				assert.Equal(t, "cannot be nil", invalidConfigErr.Message())
			},
		},
		{
			name:    "zero semaphore size defaults to one",
			wantErr: false,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					Tracer:    adapter.NewNoOpTracer(),
					Logger:    adapter.NewNoOpLogger(),
					Semaphore: adapter.NewSemaphore(0),
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.Semaphore)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock, opts := tt.setupFunc(t)
			if client != nil {
				defer client.Close()
			}
			cache, err := NewCache(client, opts)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if mock != nil {
					// Verify no unexpected Redis calls were made
					assert.NoError(t, mock.ExpectationsWereMet())
				}
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, cache, err)
			}
		})
	}
}

func TestCache_NewCache_InitializesMaps(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		Tracer:    adapter.NewNoOpTracer(),
		Logger:    adapter.NewNoOpLogger(),
		Semaphore: adapter.NewSemaphore(10),
	})
	require.NoError(t, err)

	// Verify that keyUsage and customKeys maps are initialized
	prefixes := cache.getCacheKeyUsage("nonexistent")
	assert.Empty(t, prefixes)
	assert.NotNil(t, prefixes)

	assert.NoError(t, mock.ExpectationsWereMet())
}
