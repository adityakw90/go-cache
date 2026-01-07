package cache

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/adapter"
	internaladapter "github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/errs"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
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
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache)
				assert.NotNil(t, cache.RedisClient)
				assert.NotNil(t, cache.StartSpan)
				assert.NotNil(t, cache.StartChildSpan)
				assert.NotNil(t, cache.LogProvider)
				assert.NotNil(t, cache.Semaphore)
			},
		},
		{
			name:    "nil redis client",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				_, mock := redismock.NewClientMock()
				return nil, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
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
					KeyPrefix:           "test",
					ExpireDefault:       10 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(20),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
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
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.StartSpan)
				assert.NotNil(t, cache.StartChildSpan)
			},
		},
		{
			name:    "with custom logger",
			wantErr: false,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.NotNil(t, cache.LogProvider)
			},
		},
		{
			name:    "with custom semaphore",
			wantErr: false,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
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
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
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
			name:    "nil startSpan returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           nil,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "startSpan", invalidConfigErr.Field())
				assert.Equal(t, "cannot be nil", invalidConfigErr.Message())
			},
		},
		{
			name:    "nil logger returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         nil,
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "logProvider", invalidConfigErr.Field())
				assert.Equal(t, "cannot be nil", invalidConfigErr.Message())
			},
		},
		{
			name:    "nil semaphore returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           nil,
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
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
			name:    "zero expireDefault returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       0,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "expireDefault", invalidConfigErr.Field())
				assert.Equal(t, "must be greater than zero", invalidConfigErr.Message())
			},
		},
		{
			name:    "negative expireDefault returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       -1 * time.Second,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "expireDefault", invalidConfigErr.Field())
				assert.Equal(t, "must be greater than zero", invalidConfigErr.Message())
			},
		},
		{
			name:    "zero versionExpire returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       0,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "versionExpire", invalidConfigErr.Field())
				assert.Equal(t, "must be greater than zero", invalidConfigErr.Message())
			},
		},
		{
			name:    "negative versionExpire returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       -1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "versionExpire", invalidConfigErr.Field())
				assert.Equal(t, "must be greater than zero", invalidConfigErr.Message())
			},
		},
		{
			name:    "zero lockDuration returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        0,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "lockDuration", invalidConfigErr.Field())
				assert.Equal(t, "must be greater than zero", invalidConfigErr.Message())
			},
		},
		{
			name:    "negative lockDuration returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        -1 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "lockDuration", invalidConfigErr.Field())
				assert.Equal(t, "must be greater than zero", invalidConfigErr.Message())
			},
		},
		{
			name:    "zero lockInterval returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        0,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "lockInterval", invalidConfigErr.Field())
				assert.Equal(t, "must be greater than zero", invalidConfigErr.Message())
			},
		},
		{
			name:    "negative lockInterval returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        -1 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "lockInterval", invalidConfigErr.Field())
				assert.Equal(t, "must be greater than zero", invalidConfigErr.Message())
			},
		},
		{
			name:    "nil keyGenerator returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        nil,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "keyGenerator", invalidConfigErr.Field())
				assert.Equal(t, "cannot be nil", invalidConfigErr.Message())
			},
		},
		{
			name:    "nil keyVersionGenerator returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: nil,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "keyVersionGenerator", invalidConfigErr.Field())
				assert.Equal(t, "cannot be nil", invalidConfigErr.Message())
			},
		},
		{
			name:    "nil versionGenerator returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    nil,
					LockGenerator:       key.LockGenerator,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "versionGenerator", invalidConfigErr.Field())
				assert.Equal(t, "cannot be nil", invalidConfigErr.Message())
			},
		},
		{
			name:    "nil lockGenerator returns error",
			wantErr: true,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(10),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       nil,
				}
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "lockGenerator", invalidConfigErr.Field())
				assert.Equal(t, "cannot be nil", invalidConfigErr.Message())
			},
		},
		{
			name:    "zero semaphore size defaults to one",
			wantErr: false,
			setupFunc: func(t *testing.T) (*redis.Client, redismock.ClientMock, Options) {
				client, mock := redismock.NewClientMock()
				return client, mock, Options{
					ExpireDefault:       1 * time.Minute,
					VersionExpire:       1 * time.Hour,
					LockDuration:        5 * time.Second,
					LockInterval:        100 * time.Millisecond,
					StartSpan:           adapter.NoOpStartSpan,
					StartChildSpan:      adapter.NoOpStartChildSpan,
					LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
					Semaphore:           internaladapter.NewSemaphore(0),
					KeyGenerator:        key.KeyGenerator,
					KeyVersionGenerator: key.KeyVersionGenerator,
					VersionGenerator:    key.VersionGenerator,
					LockGenerator:       key.LockGenerator,
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
		ExpireDefault:       1 * time.Minute,
		VersionExpire:       1 * time.Hour,
		LockDuration:        5 * time.Second,
		LockInterval:        100 * time.Millisecond,
		StartSpan:           adapter.NoOpStartSpan,
		StartChildSpan:      adapter.NoOpStartChildSpan,
		LogProvider:         func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
		Semaphore:           internaladapter.NewSemaphore(10),
		KeyGenerator:        key.KeyGenerator,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	// Verify that keyUsage and customKeys maps are initialized
	prefixes := cache.getCacheKeyUsage("nonexistent")
	assert.Empty(t, prefixes)
	assert.NotNil(t, prefixes)

	assert.NoError(t, mock.ExpectationsWereMet())
}
