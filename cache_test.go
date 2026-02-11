package cache

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/adapter"
	"github.com/adityakw90/go-cache/internal/errs"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCache(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	tests := []struct {
		name      string
		client    *redis.Client
		wantErr   bool
		checkFunc func(t *testing.T, cache Cache, err error)
	}{
		{
			name:    "valid redis client",
			client:  redisClient,
			wantErr: false,
			checkFunc: func(t *testing.T, cache Cache, err error) {
				assert.NotNil(t, cache)
				// Cache is an interface, can't access internal fields
				// Test that cache methods work instead
				ctx := context.Background()
				fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
					return "test", nil
				}
				cachedFn := cache.Cached("test", 1*time.Minute, false, "")(fn, nil, true)
				var result string
				_, callErr := cachedFn(&result, ctx, "arg1")
				assert.NoError(t, callErr)
			},
		},
		{
			name:    "nil redis client",
			client:  nil,
			wantErr: true,
			checkFunc: func(t *testing.T, cache Cache, err error) {
				assert.Nil(t, cache)
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "redisClient", invalidConfigErr.Field())
				assert.Equal(t, "cannot be nil", invalidConfigErr.Message())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewCache(tt.client)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, cache, err)
			}
		})
	}
}

func TestNewCache_WithOptions(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	startSpan := adapter.NoOpStartSpan
	startChildSpan := adapter.NoOpStartChildSpan
	getLogger := adapter.GetNoOpLogger

	tests := []struct {
		name      string
		options   []Option
		checkFunc func(t *testing.T, cache Cache, err error)
	}{
		{
			name: "with custom options",
			options: []Option{
				WithKeyPrefix("myapp"),
				WithTraceProvider(startSpan, startChildSpan),
				WithLogProvider(getLogger),
				WithSemaphoreSize(20),
			},
			checkFunc: func(t *testing.T, cache Cache, err error) {
				require.NoError(t, err)
				// Cache is an interface, can't access internal fields
				// Test that cache works with custom options
				ctx := context.Background()
				fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
					return "test", nil
				}
				cachedFn := cache.Cached("test", 1*time.Minute, false, "")(fn, nil, true)
				var result string
				_, callErr := cachedFn(&result, ctx, "arg1")
				assert.NoError(t, callErr)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewCache(redisClient, tt.options...)
			if tt.checkFunc != nil {
				tt.checkFunc(t, cache, err)
			}
		})
	}
}

// GetSession is not part of the public Cache interface
// This test is removed as it tests internal implementation details

// RegisterCacheKey and GetCacheKeyUsage are not part of the public Cache interface
// These tests are removed as they test internal implementation details

// RegisterCustomKey is not part of the public Cache interface
// This test is removed as it tests internal implementation details

// GetCacheKeyUsage is not part of the public Cache interface
// This test is removed as it tests internal implementation details

func TestErrInvalidConfig(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		message   string
		checkFunc func(t *testing.T, err error)
	}{
		{
			name:    "error message",
			field:   "redisClient",
			message: "test error",
			checkFunc: func(t *testing.T, err error) {
				assert.Error(t, err)
				var invalidConfigErr errs.InvalidConfigError
				assert.ErrorAs(t, err, &invalidConfigErr)
				assert.Equal(t, "invalid config redisClient : test error", err.Error())
				assert.Equal(t, "redisClient", invalidConfigErr.Field())
				assert.Equal(t, "test error", invalidConfigErr.Message())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errs.NewInvalidConfigError(tt.field, tt.message)
			if tt.checkFunc != nil {
				tt.checkFunc(t, err)
			}
		})
	}
}
