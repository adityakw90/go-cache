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
		opts      Options
		wantErr   bool
		checkFunc func(t *testing.T, cache *Cache, err error)
	}{
		{
			name:    "valid redis client with default options",
			opts:    Options{},
			wantErr: false,
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
			opts:    Options{},
			wantErr: true,
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
			name: "with custom options",
			opts: Options{
				KeyPrefix:     "test",
				ExpireDefault: 10 * time.Minute,
				VersionExpire: 1 * time.Hour,
				LockDuration:  5 * time.Second,
				LockInterval:  100 * time.Millisecond,
				SemaphoreSize: 20,
			},
			wantErr: false,
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
			name: "with custom tracer",
			opts: Options{
				Tracer: adapter.NewNoOpTracer(),
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.Tracer)
			},
		},
		{
			name: "with custom logger",
			opts: Options{
				Logger: adapter.NewNoOpLogger(),
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.Logger)
			},
		},
		{
			name: "with custom semaphore",
			opts: Options{
				Semaphore: adapter.NewSemaphore(10),
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.Semaphore)
			},
		},
		{
			name: "with key generators",
			opts: Options{
				KeyGenerator:        key.KeyGenerator,
				KeyVersionGenerator: key.KeyVersionGenerator,
				VersionGenerator:    key.VersionGenerator,
				LockGenerator:       key.LockGenerator,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.KeyGenerator)
				assert.NotNil(t, cache.KeyVersionGenerator)
				assert.NotNil(t, cache.VersionGenerator)
				assert.NotNil(t, cache.LockGenerator)
			},
		},
		{
			name: "nil semaphore creates default",
			opts: Options{
				SemaphoreSize: 5,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.Semaphore)
			},
		},
		{
			name: "nil tracer creates no-op",
			opts: Options{
				Tracer: nil,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.Tracer)
			},
		},
		{
			name: "nil logger creates no-op",
			opts: Options{
				Logger: nil,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.Logger)
			},
		},
		{
			name: "zero semaphore size defaults to one",
			opts: Options{
				SemaphoreSize: 0,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.Semaphore)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *redis.Client
			var mock redismock.ClientMock
			client, mock = redismock.NewClientMock()
			defer client.Close()

			if tt.wantErr {
				if tt.name == "nil redis client" {
					client = nil
				}
			}

			cache, err := NewCache(client, tt.opts)
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

	cache, err := NewCache(client, Options{})
	require.NoError(t, err)

	// Verify that keyUsage and customKeys maps are initialized
	prefixes := cache.getCacheKeyUsage("nonexistent")
	assert.Empty(t, prefixes)
	assert.NotNil(t, prefixes)

	assert.NoError(t, mock.ExpectationsWereMet())
}
