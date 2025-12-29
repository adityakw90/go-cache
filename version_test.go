package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_getCacheVersion(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T, cache *Cache, redisClient *redis.Client) func()
		namespace     string
		prefix        string
		options       []Option
		wantVersion   int
		wantErr       bool
		errContains   string
		verifyRedis   func(t *testing.T, cache *Cache, redisClient *redis.Client, expectedVersion int)
		skipIfNoRedis bool
	}{
		{
			name:      "success existing version",
			namespace: "testNamespace",
			prefix:    "testPrefix",
			setup: func(t *testing.T, cache *Cache, redisClient *redis.Client) func() {
				ctx := context.Background()
				key, err := cache.versionGenerator(map[string]string{
					"prefix":    "testPrefix",
					"namespace": "testNamespace",
				})
				require.NoError(t, err)
				err = redisClient.Set(ctx, key, "5", 0).Err()
				require.NoError(t, err)
				return func() {}
			},
			wantVersion: 5,
			wantErr:     false,
		},
		{
			name:      "success new version initialization",
			namespace: "testNamespaceNew",
			prefix:    "testPrefix",
			setup: func(t *testing.T, cache *Cache, redisClient *redis.Client) func() {
				return func() {}
			},
			wantVersion: 1,
			wantErr:     false,
			verifyRedis: func(t *testing.T, cache *Cache, redisClient *redis.Client, expectedVersion int) {
				ctx := context.Background()
				key, err := cache.versionGenerator(map[string]string{
					"prefix":    "testPrefix",
					"namespace": "testNamespaceNew",
				})
				require.NoError(t, err)
				val, err := redisClient.Get(ctx, key).Result()
				require.NoError(t, err)
				assert.Equal(t, "1", val)
			},
		},
		{
			name:      "error version generator fails",
			namespace: "test",
			prefix:    "test",
			options: []Option{
				WithVersionGenerator(func(data map[string]string) (string, error) {
					return "", errors.New("generator error")
				}),
			},
			wantVersion: 0,
			wantErr:     true,
			errContains: "failed to generate version key",
		},
		{
			name:      "error parse version fails",
			namespace: "testNamespaceInvalid",
			prefix:    "testPrefix",
			setup: func(t *testing.T, cache *Cache, redisClient *redis.Client) func() {
				ctx := context.Background()
				key, err := cache.versionGenerator(map[string]string{
					"prefix":    "testPrefix",
					"namespace": "testNamespaceInvalid",
				})
				require.NoError(t, err)
				err = redisClient.Set(ctx, key, "not-a-number", 0).Err()
				require.NoError(t, err)
				return func() {}
			},
			wantVersion: 0,
			wantErr:     true,
			errContains: "failed to parse version",
		},
		{
			name:      "error redis get fails",
			namespace: "testNamespace",
			prefix:    "testPrefix",
			setup: func(t *testing.T, cache *Cache, redisClient *redis.Client) func() {
				redisClient.Close()
				return func() {}
			},
			wantVersion:   0,
			wantErr:       true,
			skipIfNoRedis: true,
		},
		{
			name:      "error redis incr fails",
			namespace: "testNamespaceIncrFail",
			prefix:    "testPrefix",
			setup: func(t *testing.T, cache *Cache, redisClient *redis.Client) func() {
				redisClient.Close()
				return func() {}
			},
			wantVersion:   0,
			wantErr:       true,
			skipIfNoRedis: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient := createTestRedisClient(t)
			defer redisClient.Close()

			cache, err := NewCache(redisClient, tt.options...)
			require.NoError(t, err)

			var cleanup func()
			if tt.setup != nil {
				cleanup = tt.setup(t, cache, redisClient)
				if cleanup != nil {
					defer cleanup()
				}
			}

			ctx := context.Background()
			version, err := cache.getCacheVersion(ctx, tt.namespace, tt.prefix)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantVersion, version)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantVersion, version)
				if tt.verifyRedis != nil {
					tt.verifyRedis(t, cache, redisClient, tt.wantVersion)
				}
			}
		})
	}
}

func TestCache_incrementCacheVersion(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T, cache *Cache, redisClient *redis.Client) func()
		namespace     string
		prefix        string
		ttl           time.Duration
		options       []Option
		wantErr       bool
		errContains   string
		verifyRedis   func(t *testing.T, cache *Cache, redisClient *redis.Client, initialVersion int)
		skipIfNoRedis bool
	}{
		{
			name:      "success increment version",
			namespace: "testNamespace",
			prefix:    "testPrefix",
			ttl:       time.Hour,
			setup: func(t *testing.T, cache *Cache, redisClient *redis.Client) func() {
				ctx := context.Background()
				key, err := cache.versionGenerator(map[string]string{
					"prefix":    "testPrefix",
					"namespace": "testNamespace",
				})
				require.NoError(t, err)
				err = redisClient.Set(ctx, key, "10", 0).Err()
				require.NoError(t, err)
				return func() {}
			},
			wantErr: false,
			verifyRedis: func(t *testing.T, cache *Cache, redisClient *redis.Client, initialVersion int) {
				ctx := context.Background()
				key, err := cache.versionGenerator(map[string]string{
					"prefix":    "testPrefix",
					"namespace": "testNamespace",
				})
				require.NoError(t, err)
				val, err := redisClient.Get(ctx, key).Result()
				require.NoError(t, err)
				assert.Equal(t, "11", val)
			},
		},
		{
			name:      "error version generator fails",
			namespace: "test",
			prefix:    "test",
			ttl:       time.Hour,
			options: []Option{
				WithVersionGenerator(func(data map[string]string) (string, error) {
					return "", errors.New("generator error")
				}),
			},
			wantErr:     true,
			errContains: "failed to generate version key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient := createTestRedisClient(t)
			defer redisClient.Close()

			cache, err := NewCache(redisClient, tt.options...)
			require.NoError(t, err)

			var cleanup func()
			if tt.setup != nil {
				cleanup = tt.setup(t, cache, redisClient)
				if cleanup != nil {
					defer cleanup()
				}
			}

			ctx := context.Background()
			session := cache.GetSession()
			_, err = cache.incrementCacheVersion(ctx, session, tt.prefix, tt.namespace, tt.ttl)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				_, err = session.Exec(ctx)
				require.NoError(t, err)
				if tt.verifyRedis != nil {
					tt.verifyRedis(t, cache, redisClient, 10)
				}
			}
		})
	}
}

func TestCache_checkVersionTtlAsync(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T, cache *Cache, redisClient *redis.Client) func()
		key           string
		ttl           time.Duration
		wantErr       bool
		verifyTTL     func(t *testing.T, redisClient *redis.Client, key string)
		skipIfNoRedis bool
	}{
		{
			name: "success key with TTL",
			key:  "test:key:with:ttl",
			ttl:  time.Hour,
			setup: func(t *testing.T, cache *Cache, redisClient *redis.Client) func() {
				ctx := context.Background()
				err := redisClient.Set(ctx, "test:key:with:ttl", "value", time.Hour).Err()
				require.NoError(t, err)
				return func() {}
			},
			wantErr: false,
			verifyTTL: func(t *testing.T, redisClient *redis.Client, key string) {
				ctx := context.Background()
				ttl, err := redisClient.TTL(ctx, key).Result()
				require.NoError(t, err)
				assert.Greater(t, ttl, time.Hour-time.Minute)
				assert.LessOrEqual(t, ttl, time.Hour)
			},
		},
		{
			name: "success key without TTL",
			key:  "test:key:without:ttl",
			ttl:  time.Hour,
			setup: func(t *testing.T, cache *Cache, redisClient *redis.Client) func() {
				ctx := context.Background()
				err := redisClient.Set(ctx, "test:key:without:ttl", "value", 0).Err()
				require.NoError(t, err)
				return func() {}
			},
			wantErr: false,
			verifyTTL: func(t *testing.T, redisClient *redis.Client, key string) {
				ctx := context.Background()
				ttl, err := redisClient.TTL(ctx, key).Result()
				require.NoError(t, err)
				assert.Greater(t, ttl, time.Hour-time.Minute)
				assert.LessOrEqual(t, ttl, time.Hour)
			},
		},
		{
			name: "error TTL fails",
			key:  "test:key:ttl:error",
			ttl:  time.Hour,
			setup: func(t *testing.T, cache *Cache, redisClient *redis.Client) func() {
				redisClient.Close()
				return func() {}
			},
			wantErr:       false, // Function handles error gracefully, doesn't return error
			skipIfNoRedis: true,
		},
		{
			name: "error expire fails",
			key:  "test:key:expire:error",
			ttl:  time.Hour,
			setup: func(t *testing.T, cache *Cache, redisClient *redis.Client) func() {
				ctx := context.Background()
				err := redisClient.Set(ctx, "test:key:expire:error", "value", 0).Err()
				require.NoError(t, err)
				redisClient.Close()
				return func() {}
			},
			wantErr:       false, // Function handles error gracefully, doesn't return error
			skipIfNoRedis: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient := createTestRedisClient(t)
			defer redisClient.Close()

			cache, err := NewCache(redisClient)
			require.NoError(t, err)

			var cleanup func()
			if tt.setup != nil {
				cleanup = tt.setup(t, cache, redisClient)
				if cleanup != nil {
					defer cleanup()
				}
			}

			ctx := context.Background()
			_, span := cache.tracer.StartSpan(ctx, "test")
			cache.checkVersionTtlAsync(span, tt.key, tt.ttl)

			// Wait for async operation to complete
			time.Sleep(200 * time.Millisecond)

			if tt.verifyTTL != nil {
				tt.verifyTTL(t, redisClient, tt.key)
			}
		})
	}
}

func TestCache_InvalidateVersion(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T, cache *Cache, redisClient *redis.Client) func()
		namespace     string
		prefix        string
		options       []Option
		wantErr       bool
		errContains   string
		verifyRedis   func(t *testing.T, cache *Cache, redisClient *redis.Client, initialVersion int)
		skipIfNoRedis bool
	}{
		{
			name:      "success invalidate version",
			namespace: "testNamespace",
			prefix:    "testPrefix",
			setup: func(t *testing.T, cache *Cache, redisClient *redis.Client) func() {
				ctx := context.Background()
				key, err := cache.versionGenerator(map[string]string{
					"prefix":    "testPrefix",
					"namespace": "testNamespace",
				})
				require.NoError(t, err)
				err = redisClient.Set(ctx, key, "5", 0).Err()
				require.NoError(t, err)
				return func() {}
			},
			wantErr: false,
			verifyRedis: func(t *testing.T, cache *Cache, redisClient *redis.Client, initialVersion int) {
				ctx := context.Background()
				key, err := cache.versionGenerator(map[string]string{
					"prefix":    "testPrefix",
					"namespace": "testNamespace",
				})
				require.NoError(t, err)
				val, err := redisClient.Get(ctx, key).Result()
				require.NoError(t, err)
				assert.Equal(t, "6", val)
			},
		},
		{
			name:      "success default prefix",
			namespace: "testNamespace",
			prefix:    "",
			setup: func(t *testing.T, cache *Cache, redisClient *redis.Client) func() {
				ctx := context.Background()
				key, err := cache.versionGenerator(map[string]string{
					"prefix":    cache.keyPrefix,
					"namespace": "testNamespace",
				})
				require.NoError(t, err)
				err = redisClient.Set(ctx, key, "3", 0).Err()
				require.NoError(t, err)
				return func() {}
			},
			wantErr: false,
			verifyRedis: func(t *testing.T, cache *Cache, redisClient *redis.Client, initialVersion int) {
				ctx := context.Background()
				key, err := cache.versionGenerator(map[string]string{
					"prefix":    cache.keyPrefix,
					"namespace": "testNamespace",
				})
				require.NoError(t, err)
				val, err := redisClient.Get(ctx, key).Result()
				require.NoError(t, err)
				assert.Equal(t, "4", val)
			},
		},
		{
			name:      "error increment fails",
			namespace: "test",
			prefix:    "test",
			options: []Option{
				WithVersionGenerator(func(data map[string]string) (string, error) {
					return "", errors.New("generator error")
				}),
			},
			wantErr:     true,
			errContains: "failed to increment version",
		},
		{
			name:      "error pipeline exec fails",
			namespace: "testNamespace",
			prefix:    "testPrefix",
			setup: func(t *testing.T, cache *Cache, redisClient *redis.Client) func() {
				redisClient.Close()
				return func() {}
			},
			wantErr:       true,
			errContains:   "failed to execute pipeline",
			skipIfNoRedis: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisClient := createTestRedisClient(t)
			defer redisClient.Close()

			cache, err := NewCache(redisClient, tt.options...)
			require.NoError(t, err)

			var cleanup func()
			if tt.setup != nil {
				cleanup = tt.setup(t, cache, redisClient)
				if cleanup != nil {
					defer cleanup()
				}
			}

			ctx := context.Background()
			err = cache.InvalidateVersion(ctx, tt.namespace, tt.prefix)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				if tt.verifyRedis != nil {
					tt.verifyRedis(t, cache, redisClient, 0)
				}
			}
		})
	}
}

func TestCache_getCacheVersion_MultipleCalls(t *testing.T) {
	redisClient := createTestRedisClient(t)
	defer redisClient.Close()

	cache, err := NewCache(redisClient)
	require.NoError(t, err)

	ctx := context.Background()
	namespace := "testNamespaceMulti"
	prefix := "testPrefix"

	// First call - should initialize to 1
	version1, err := cache.getCacheVersion(ctx, namespace, prefix)
	require.NoError(t, err)
	assert.Equal(t, 1, version1)

	// Second call - should return 1 (same value)
	version2, err := cache.getCacheVersion(ctx, namespace, prefix)
	require.NoError(t, err)
	assert.Equal(t, 1, version2)

	// Increment manually
	key, err := cache.versionGenerator(map[string]string{
		"prefix":    prefix,
		"namespace": namespace,
	})
	require.NoError(t, err)

	err = redisClient.Incr(ctx, key).Err()
	require.NoError(t, err)

	// Third call - should return 2
	version3, err := cache.getCacheVersion(ctx, namespace, prefix)
	require.NoError(t, err)
	assert.Equal(t, 2, version3)
}

func TestCache_incrementCacheVersion_MultipleIncrements(t *testing.T) {
	redisClient := createTestRedisClient(t)
	defer redisClient.Close()

	cache, err := NewCache(redisClient)
	require.NoError(t, err)

	ctx := context.Background()
	namespace := "testNamespace"
	prefix := "testPrefix"

	// Set initial version
	key, err := cache.versionGenerator(map[string]string{
		"prefix":    prefix,
		"namespace": namespace,
	})
	require.NoError(t, err)

	err = redisClient.Set(ctx, key, "1", 0).Err()
	require.NoError(t, err)

	// Increment multiple times
	// Each increment creates a new session, so we need to execute each one
	// and verify the value increases sequentially
	for i := 0; i < 5; i++ {
		session := cache.GetSession()
		_, err := cache.incrementCacheVersion(ctx, session, prefix, namespace, time.Hour)
		require.NoError(t, err)

		// Execute pipeline to apply the increment
		_, err = session.Exec(ctx)
		require.NoError(t, err)

		// Verify the version in Redis increased
		val, err := redisClient.Get(ctx, key).Result()
		require.NoError(t, err)
		expectedVal := 2 + i // Starting from 1, after first increment it's 2
		assert.Equal(t, string(rune('0'+expectedVal)), val)
	}

	// Verify final version
	val, err := redisClient.Get(ctx, key).Result()
	require.NoError(t, err)
	assert.Equal(t, "6", val)
}
