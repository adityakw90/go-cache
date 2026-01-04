package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/adityakw90/go-cache/internal/serialize"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_Operations_Get(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		setupMock      func(mock redismock.ClientMock, key string) string
		resultType     interface{}
		expectedValue  interface{}
		wantErr        bool
		wantErrIs      error
		wantErrMessage string
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
			resultType:     new(string),
			expectedValue:  "test-value",
			wantErr:        false,
			wantErrIs:      nil,
			wantErrMessage: "",
		},
		{
			name: "cache miss",
			key:  "test:key:2",
			setupMock: func(mock redismock.ClientMock, key string) string {
				mock.ExpectGet(key).RedisNil()
				return ""
			},
			resultType:     new(string),
			wantErr:        true,
			wantErrIs:      ErrGetCacheMiss,
			wantErrMessage: "",
		},
		{
			name: "redis error",
			key:  "test:key:3",
			setupMock: func(mock redismock.ClientMock, key string) string {
				mock.ExpectGet(key).SetErr(errors.New("redis connection error"))
				return ""
			},
			resultType:     new(string),
			wantErr:        true,
			wantErrIs:      nil,
			wantErrMessage: "failed to get cache: redis connection error",
		},
		{
			name: "deserialization failure",
			key:  "test:key:4",
			setupMock: func(mock redismock.ClientMock, key string) string {
				// Return invalid/corrupted data that cannot be deserialized
				invalidData := []byte{0xFF, 0xFF, 0xFF, 0xFF}
				mock.ExpectGet(key).SetVal(string(invalidData))
				return ""
			},
			resultType:     new(string),
			wantErr:        true,
			wantErrIs:      nil,
			wantErrMessage: "failed to deserialize cache data: failed to deserialize data: unexpected EOF",
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
				VersionExpire:       1 * time.Hour,
				LockDuration:        5 * time.Second,
				LockInterval:        100 * time.Millisecond,
				KeyGenerator:        key.KeyGenerator,
				KeyVersionGenerator: key.KeyVersionGenerator,
				VersionGenerator:    key.VersionGenerator,
				LockGenerator:       key.LockGenerator,
			})
			require.NoError(t, err)

			expectedValue := tt.setupMock(mock, tt.key)

			ctx := context.Background()
			err = cache.Get(ctx, tt.key, tt.resultType)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs), "expected error %v, got %v", tt.wantErrIs, err)
				}
				if tt.wantErrMessage != "" {
					assert.Equal(t, tt.wantErrMessage, err.Error())
				}
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

func TestCache_Operations_Set(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		value          interface{}
		ttl            time.Duration
		setupMock      func(mock redismock.ClientMock, key string, ttl time.Duration)
		wantErr        bool
		wantErrIs      error
		wantErrMessage string
	}{
		{
			name:  "set string value",
			key:   "test:key:1",
			value: "test-value",
			ttl:   5 * time.Minute,
			setupMock: func(mock redismock.ClientMock, key string, ttl time.Duration) {
				mock.Regexp().ExpectSet(key, `.*`, ttl).SetVal("OK")
			},
			wantErr:        false,
			wantErrIs:      nil,
			wantErrMessage: "",
		},
		{
			name:  "set map value",
			key:   "test:key:2",
			value: map[string]string{"key": "value"},
			ttl:   10 * time.Minute,
			setupMock: func(mock redismock.ClientMock, key string, ttl time.Duration) {
				mock.Regexp().ExpectSet(key, `.*`, ttl).SetVal("OK")
			},
			wantErr:        false,
			wantErrIs:      nil,
			wantErrMessage: "",
		},
		{
			name:  "redis set error",
			key:   "test:key:3",
			value: "test-value",
			ttl:   5 * time.Minute,
			setupMock: func(mock redismock.ClientMock, key string, ttl time.Duration) {
				mock.Regexp().ExpectSet(key, `.*`, ttl).SetErr(errors.New("redis set error"))
			},
			wantErr:        true,
			wantErrIs:      nil,
			wantErrMessage: "failed to set cache data: redis set error",
		},
		{
			name:  "serialization failure",
			key:   "test:key:4",
			value: nil,
			ttl:   5 * time.Minute,
			setupMock: func(mock redismock.ClientMock, key string, ttl time.Duration) {
				// No mock expectation needed as serialization fails before Redis call
			},
			wantErr:        true,
			wantErrIs:      nil,
			wantErrMessage: "failed to serialize value: cannot serialize nil value",
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
				VersionExpire:       1 * time.Hour,
				LockDuration:        5 * time.Second,
				LockInterval:        100 * time.Millisecond,
				KeyGenerator:        key.KeyGenerator,
				KeyVersionGenerator: key.KeyVersionGenerator,
				VersionGenerator:    key.VersionGenerator,
				LockGenerator:       key.LockGenerator,
			})
			require.NoError(t, err)

			tt.setupMock(mock, tt.key, tt.ttl)

			ctx := context.Background()
			err = cache.Set(ctx, tt.key, tt.value, tt.ttl)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrIs != nil {
					assert.True(t, errors.Is(err, tt.wantErrIs), "expected error %v, got %v", tt.wantErrIs, err)
				}
				if tt.wantErrMessage != "" {
					assert.Equal(t, tt.wantErrMessage, err.Error())
				}
			} else {
				require.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCache_Operations_GetSet_Integration(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		Tracer:              adapter.NewNoOpTracer(),
		Logger:              adapter.NewNoOpLogger(),
		Semaphore:           adapter.NewSemaphore(10),
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		VersionExpire:       1 * time.Hour,
		LockDuration:        5 * time.Second,
		LockInterval:        100 * time.Millisecond,
		KeyGenerator:        key.KeyGenerator,
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
