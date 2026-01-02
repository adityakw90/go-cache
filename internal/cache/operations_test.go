package cache

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/adityakw90/go-cache/internal/serialize"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_Get(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		setupMock     func(mock redismock.ClientMock, key string) string
		resultType    interface{}
		expectedValue interface{}
		expectError   bool
	}{
		{
			name: "cache hit with string",
			key:  "test:key:1",
			setupMock: func(mock redismock.ClientMock, key string) string {
				serializedData, err := serialize.Serialize("test-value")
				require.NoError(t, err)
				mock.ExpectGet(key).SetVal(string(serializedData))
				return "test-value"
			},
			resultType:    new(string),
			expectedValue: "test-value",
			expectError:   false,
		},
		{
			name: "cache miss",
			key:  "test:key:2",
			setupMock: func(mock redismock.ClientMock, key string) string {
				mock.ExpectGet(key).RedisNil()
				return ""
			},
			resultType:  new(string),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			defer client.Close()

			cache, err := NewCache(client, Options{
				Tracer:              adapter.NewNoOpTracer(),
				Logger:              adapter.NewNoOpLogger(),
				Semaphore:           adapter.NewSemaphore(10),
				KeyPrefix:           "test",
				ExpireDefault:       5 * time.Minute,
				KeyVersionGenerator: key.KeyVersionGenerator,
				VersionGenerator:    key.VersionGenerator,
				LockGenerator:       key.LockGenerator,
			})
			require.NoError(t, err)

			expectedValue := tt.setupMock(mock, tt.key)

			ctx := context.Background()
			err = cache.Get(ctx, tt.key, tt.resultType)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if strPtr, ok := tt.resultType.(*string); ok {
					assert.Equal(t, expectedValue, *strPtr)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCache_Set(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       interface{}
		ttl         time.Duration
		expectError bool
	}{
		{
			name:        "set string value",
			key:         "test:key:1",
			value:       "test-value",
			ttl:         5 * time.Minute,
			expectError: false,
		},
		{
			name:        "set map value",
			key:         "test:key:2",
			value:       map[string]string{"key": "value"},
			ttl:         10 * time.Minute,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			defer client.Close()

			cache, err := NewCache(client, Options{
				Tracer:              adapter.NewNoOpTracer(),
				Logger:              adapter.NewNoOpLogger(),
				Semaphore:           adapter.NewSemaphore(10),
				KeyPrefix:           "test",
				ExpireDefault:       5 * time.Minute,
				KeyVersionGenerator: key.KeyVersionGenerator,
				VersionGenerator:    key.VersionGenerator,
				LockGenerator:       key.LockGenerator,
			})
			require.NoError(t, err)

			mock.Regexp().ExpectSet(tt.key, `.*`, tt.ttl).SetVal("OK")

			ctx := context.Background()
			err = cache.Set(ctx, tt.key, tt.value, tt.ttl)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCache_GetSet_Integration(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		Tracer:              adapter.NewNoOpTracer(),
		Logger:              adapter.NewNoOpLogger(),
		Semaphore:           adapter.NewSemaphore(10),
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	key := "test:integration:1"
	value := "integration-test-value"
	ttl := 5 * time.Minute

	// Set up mock for Set
	serializedData, err := serialize.Serialize(value)
	require.NoError(t, err)
	mock.Regexp().ExpectSet(key, `.*`, ttl).SetVal("OK")

	// Set up mock for Get
	mock.ExpectGet(key).SetVal(string(serializedData))

	ctx := context.Background()

	// Set value
	err = cache.Set(ctx, key, value, ttl)
	require.NoError(t, err)

	// Get value
	var result string
	err = cache.Get(ctx, key, &result)
	require.NoError(t, err)
	assert.Equal(t, value, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}
