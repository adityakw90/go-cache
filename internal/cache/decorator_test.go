package cache

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/hash"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/adityakw90/go-cache/internal/serialize"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_Cached_KeyRegistration(t *testing.T) {
	tests := []struct {
		name          string
		keyName       string
		prefix        string
		defaultPrefix string
		expectedIn    string
	}{
		{
			name:          "registers custom prefix",
			keyName:       "testFunc",
			prefix:        "customPrefix",
			defaultPrefix: "test",
			expectedIn:    "customPrefix",
		},
		{
			name:          "uses default prefix when empty",
			keyName:       "testFunc",
			prefix:        "",
			defaultPrefix: "default",
			expectedIn:    "default",
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
				KeyPrefix:           tt.defaultPrefix,
				ExpireDefault:       5 * time.Minute,
				KeyVersionGenerator: key.KeyVersionGenerator,
				VersionGenerator:    key.VersionGenerator,
				LockGenerator:       key.LockGenerator,
			})
			require.NoError(t, err)

			_ = cache.Cached(
				tt.keyName,
				5*time.Minute,
				false,
				tt.prefix,
			)

			prefixes := cache.getCacheKeyUsage(tt.keyName)
			assert.Contains(t, prefixes, tt.expectedIn)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCache_Cached_CustomKey(t *testing.T) {
	tests := []struct {
		name        string
		keyName     string
		prefix      string
		setupCustom func(t *testing.T) key.CustomKeyFunction
		args        []interface{}
	}{
		{
			name:    "registers custom key function",
			keyName: "getUser",
			prefix:  "user",
			setupCustom: func(t *testing.T) key.CustomKeyFunction {
				customKey, err := key.NewCustomKeyFunction(
					"getUser",
					func(args ...interface{}) string {
						return "user:" + args[0].(string)
					},
					[]string{"uid"},
				)
				require.NoError(t, err)
				return customKey
			},
			args: []interface{}{"123"},
		},
		{
			name:    "handles empty args with custom key",
			keyName: "getUser",
			prefix:  "user",
			setupCustom: func(t *testing.T) key.CustomKeyFunction {
				customKey, err := key.NewCustomKeyFunction(
					"getUser",
					func(args ...interface{}) string {
						if len(args) == 0 {
							return "default"
						}
						return "user:" + args[0].(string)
					},
					[]string{"uid"},
				)
				require.NoError(t, err)
				return customKey
			},
			args: []interface{}{},
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

			customKey := tt.setupCustom(t)

			fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
				return "result", nil
			}

			cachedFunc := cache.Cached(
				tt.keyName,
				5*time.Minute,
				false,
				tt.prefix,
			)(fn, customKey)

			var result string
			res, err := cachedFunc(&result, context.Background(), tt.args...)
			require.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, "result", res)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCache_Cached_Versioning(t *testing.T) {
	tests := []struct {
		name            string
		versioning      bool
		versionExpire   time.Duration
		setupVersionGen func() func(data map[string]string) (string, error)
		expectCallCount int
		expectError     bool
	}{
		{
			name:          "with versioning enabled",
			versioning:    true,
			versionExpire: 1 * time.Hour,
			setupVersionGen: func() func(data map[string]string) (string, error) {
				return key.VersionGenerator
			},
			expectCallCount: 1,
			expectError:     false,
		},
		{
			name:          "without versioning",
			versioning:    false,
			versionExpire: 0,
			setupVersionGen: func() func(data map[string]string) (string, error) {
				return key.VersionGenerator
			},
			expectCallCount: 1,
			expectError:     false,
		},
		{
			name:          "version error falls back to function",
			versioning:    true,
			versionExpire: 1 * time.Hour,
			setupVersionGen: func() func(data map[string]string) (string, error) {
				return func(data map[string]string) (string, error) {
					return "", assert.AnError
				}
			},
			expectCallCount: 1,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := redismock.NewClientMock()
			defer client.Close()

			opts := Options{
				Tracer:              adapter.NewNoOpTracer(),
				Logger:              adapter.NewNoOpLogger(),
				Semaphore:           adapter.NewSemaphore(10),
				KeyPrefix:           "test",
				ExpireDefault:       5 * time.Minute,
				KeyVersionGenerator: key.KeyVersionGenerator,
				VersionGenerator:    tt.setupVersionGen(),
				LockGenerator:       key.LockGenerator,
			}
			if tt.versioning {
				opts.VersionExpire = tt.versionExpire
			}

			cache, err := NewCache(client, opts)
			require.NoError(t, err)

			ctx := context.Background()
			callCount := 0

			fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
				callCount++
				return "result", nil
			}

			cachedFunc := cache.Cached(
				"testFunc",
				5*time.Minute,
				tt.versioning,
				"test",
			)(fn, nil)

			var result string
			_, err = cachedFunc(&result, ctx, "arg1")
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, callCount, tt.expectCallCount)
			}

			time.Sleep(300 * time.Millisecond)
		})
	}
}

func TestCache_Cached_ErrorHandling(t *testing.T) {
	tests := []struct {
		name            string
		setupKeyGen     func() func(data map[string]string) (string, error)
		setupFn         func() func(ctx context.Context, args ...interface{}) (interface{}, error)
		expectCallCount int
		expectResult    interface{}
		expectError     bool
	}{
		{
			name: "function error propagates",
			setupKeyGen: func() func(data map[string]string) (string, error) {
				return key.KeyVersionGenerator
			},
			setupFn: func() func(ctx context.Context, args ...interface{}) (interface{}, error) {
				return func(ctx context.Context, args ...interface{}) (interface{}, error) {
					return nil, assert.AnError
				}
			},
			expectCallCount: 1,
			expectResult:    nil,
			expectError:     true,
		},
		{
			name: "key generation error falls back to function",
			setupKeyGen: func() func(data map[string]string) (string, error) {
				return func(data map[string]string) (string, error) {
					return "", assert.AnError
				}
			},
			setupFn: func() func(ctx context.Context, args ...interface{}) (interface{}, error) {
				return func(ctx context.Context, args ...interface{}) (interface{}, error) {
					return "result", nil
				}
			},
			expectCallCount: 1,
			expectResult:    "result",
			expectError:     false,
		},
		{
			name: "redis get error falls back to function",
			setupKeyGen: func() func(data map[string]string) (string, error) {
				return key.KeyVersionGenerator
			},
			setupFn: func() func(ctx context.Context, args ...interface{}) (interface{}, error) {
				return func(ctx context.Context, args ...interface{}) (interface{}, error) {
					return "result", nil
				}
			},
			expectCallCount: 1,
			expectResult:    "result",
			expectError:     false,
		},
		{
			name: "deserialize error falls back to function",
			setupKeyGen: func() func(data map[string]string) (string, error) {
				return key.KeyVersionGenerator
			},
			setupFn: func() func(ctx context.Context, args ...interface{}) (interface{}, error) {
				return func(ctx context.Context, args ...interface{}) (interface{}, error) {
					return "result", nil
				}
			},
			expectCallCount: 1,
			expectResult:    "result",
			expectError:     false,
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
				KeyVersionGenerator: tt.setupKeyGen(),
				VersionGenerator:    key.VersionGenerator,
				LockGenerator:       key.LockGenerator,
			})
			require.NoError(t, err)

			ctx := context.Background()
			callCount := 0

			fn := tt.setupFn()
			wrappedFn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
				callCount++
				return fn(ctx, args...)
			}

			cachedFunc := cache.Cached(
				"testFunc",
				5*time.Minute,
				false,
				"test",
			)(wrappedFn, nil)

			// Setup mocks based on test case
			if tt.name == "redis get error falls back to function" {
				namespace := "testFunc"
				hashKey := hash.CacheKey(namespace, []interface{}{"arg1"})
				cacheKey, err := key.KeyVersionGenerator(map[string]string{
					"prefix":    "test",
					"namespace": namespace,
					"version":   "0",
					"key":       hashKey,
				})
				require.NoError(t, err)
				mock.ExpectGet(cacheKey).SetErr(assert.AnError)
			} else if tt.name == "deserialize error falls back to function" {
				namespace := "testFunc"
				hashKey := hash.CacheKey(namespace, []interface{}{"arg1"})
				cacheKey, err := key.KeyVersionGenerator(map[string]string{
					"prefix":    "test",
					"namespace": namespace,
					"version":   "0",
					"key":       hashKey,
				})
				require.NoError(t, err)
				// First call - cache miss, then async update
				mock.ExpectGet(cacheKey).RedisNil()
				time.Sleep(300 * time.Millisecond)
			}

			var result string
			if tt.name == "deserialize error falls back to function" {
				// Second call with wrong type
				var result2 int
				res, err := cachedFunc(&result2, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, tt.expectResult, res)
			} else {
				res, err := cachedFunc(&result, ctx, "arg1")
				if tt.expectError {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.expectResult, res)
				}
			}

			assert.Equal(t, tt.expectCallCount, callCount)

			if tt.name != "deserialize error falls back to function" {
				time.Sleep(300 * time.Millisecond)
			}
		})
	}
}

func TestCache_Cached_CacheHit(t *testing.T) {
	tests := []struct {
		name            string
		versioning      bool
		setupMock       func(t *testing.T, mock redismock.ClientMock, namespace string, hashKey string) string
		expectedResult  string
		expectCallCount int
	}{
		{
			name:       "cache hit without versioning",
			versioning: false,
			setupMock: func(t *testing.T, mock redismock.ClientMock, namespace string, hashKey string) string {
				cacheKey, err := key.KeyVersionGenerator(map[string]string{
					"prefix":    "test",
					"namespace": namespace,
					"version":   "0",
					"key":       hashKey,
				})
				require.NoError(t, err)

				serializedData, err := serialize.Serialize("cached-result")
				require.NoError(t, err)

				mock.ExpectGet(cacheKey).SetVal(string(serializedData))
				return cacheKey
			},
			expectedResult:  "cached-result",
			expectCallCount: 0,
		},
		{
			name:       "cache hit with versioning",
			versioning: true,
			setupMock: func(t *testing.T, mock redismock.ClientMock, namespace string, hashKey string) string {
				versionKey, err := key.VersionGenerator(map[string]string{
					"prefix":    "test",
					"namespace": namespace,
				})
				require.NoError(t, err)

				cacheKey, err := key.KeyVersionGenerator(map[string]string{
					"prefix":    "test",
					"namespace": namespace,
					"version":   "1",
					"key":       hashKey,
				})
				require.NoError(t, err)

				serializedData, err := serialize.Serialize("versioned-result")
				require.NoError(t, err)

				mock.ExpectGet(versionKey).SetVal("1")
				mock.ExpectGet(cacheKey).SetVal(string(serializedData))
				return cacheKey
			},
			expectedResult:  "versioned-result",
			expectCallCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			defer client.Close()

			opts := Options{
				Tracer:              adapter.NewNoOpTracer(),
				Logger:              adapter.NewNoOpLogger(),
				Semaphore:           adapter.NewSemaphore(10),
				KeyPrefix:           "test",
				ExpireDefault:       5 * time.Minute,
				KeyVersionGenerator: key.KeyVersionGenerator,
				VersionGenerator:    key.VersionGenerator,
				LockGenerator:       key.LockGenerator,
			}
			if tt.versioning {
				opts.VersionExpire = 1 * time.Hour
			}

			cache, err := NewCache(client, opts)
			require.NoError(t, err)

			ctx := context.Background()
			callCount := 0

			fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
				callCount++
				if tt.versioning {
					return "versioned-result", nil
				}
				return "cached-result", nil
			}

			cachedFunc := cache.Cached(
				"testFunc",
				5*time.Minute,
				tt.versioning,
				"test",
			)(fn, nil)

			namespace := "testFunc"
			hashKey := hash.CacheKey(namespace, []interface{}{"arg1"})
			_ = tt.setupMock(t, mock, namespace, hashKey)

			var result string
			res, err := cachedFunc(&result, ctx, "arg1")
			require.NoError(t, err)
			assert.Equal(t, tt.expectCallCount, callCount)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, &result, res)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCache_Cached_TTL(t *testing.T) {
	tests := []struct {
		name            string
		ttl             interface{}
		defaultTTL      time.Duration
		expectedTTL     time.Duration
		setupLock       bool
		expectCallCount int
	}{
		{
			name: "dynamic TTL function",
			ttl: func(result interface{}, args ...interface{}) time.Duration {
				if m, ok := result.(map[string]string); ok && m["status"] == "ok" {
					return 10 * time.Minute
				}
				return 5 * time.Minute
			},
			defaultTTL:      5 * time.Minute,
			expectedTTL:     10 * time.Minute,
			setupLock:       true,
			expectCallCount: 1,
		},
		{
			name:            "default TTL when invalid type",
			ttl:             "invalid-ttl-type",
			defaultTTL:      7 * time.Minute,
			expectedTTL:     7 * time.Minute,
			setupLock:       true,
			expectCallCount: 1,
		},
		{
			name:            "static TTL duration",
			ttl:             5 * time.Minute,
			defaultTTL:      5 * time.Minute,
			expectedTTL:     5 * time.Minute,
			setupLock:       false,
			expectCallCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			defer client.Close()

			opts := Options{
				Tracer:              adapter.NewNoOpTracer(),
				Logger:              adapter.NewNoOpLogger(),
				Semaphore:           adapter.NewSemaphore(10),
				KeyPrefix:           "test",
				ExpireDefault:       tt.defaultTTL,
				KeyVersionGenerator: key.KeyVersionGenerator,
				VersionGenerator:    key.VersionGenerator,
				LockGenerator:       key.LockGenerator,
			}
			if tt.setupLock {
				opts.LockDuration = 5 * time.Second
				opts.LockInterval = 100 * time.Millisecond
			}

			cache, err := NewCache(client, opts)
			require.NoError(t, err)

			ctx := context.Background()
			callCount := 0

			var fn func(ctx context.Context, args ...interface{}) (interface{}, error)
			if tt.name == "dynamic TTL function" {
				fn = func(ctx context.Context, args ...interface{}) (interface{}, error) {
					callCount++
					return map[string]string{"status": "ok"}, nil
				}
			} else {
				fn = func(ctx context.Context, args ...interface{}) (interface{}, error) {
					callCount++
					return "result", nil
				}
			}

			cachedFunc := cache.Cached(
				"testFunc",
				tt.ttl,
				false,
				"test",
			)(fn, nil)

			namespace := "testFunc"
			hashKey := hash.CacheKey(namespace, []interface{}{"arg1"})
			cacheKey, err := key.KeyVersionGenerator(map[string]string{
				"prefix":    "test",
				"namespace": namespace,
				"version":   "0",
				"key":       hashKey,
			})
			require.NoError(t, err)

			mock.ExpectGet(cacheKey).RedisNil()

			if tt.setupLock {
				lockKey, err := key.LockGenerator(map[string]string{
					"prefix":    "test",
					"namespace": namespace,
				})
				require.NoError(t, err)

				mock.Regexp().ExpectSetNX(lockKey, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, 5*time.Second).SetVal(true)
				mock.Regexp().ExpectSet(cacheKey, `.*`, tt.expectedTTL).SetVal("OK")
				mock.Regexp().ExpectEvalSha(`.*`, []string{lockKey}, []interface{}{`.*`}).SetErr(assert.AnError)
				mock.Regexp().ExpectEval(`.*`, []string{lockKey}, []interface{}{`.*`}).SetVal(int64(1))
			}

			if tt.name == "dynamic TTL function" {
				var resultMap map[string]string
				res, err := cachedFunc(&resultMap, ctx, "arg1")
				require.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, map[string]string{"status": "ok"}, res)
			} else {
				var resultStr string
				res, err := cachedFunc(&resultStr, ctx, "arg1")
				require.NoError(t, err)
				assert.Equal(t, "result", res)
			}
			assert.Equal(t, tt.expectCallCount, callCount)

			time.Sleep(300 * time.Millisecond)
		})
	}
}

func TestCache_Cached_CacheMiss(t *testing.T) {
	tests := []struct {
		name            string
		args            []interface{}
		expectedResult  string
		expectCallCount int
	}{
		{
			name:            "cache miss calls function",
			args:            []interface{}{"arg1"},
			expectedResult:  "result-arg1",
			expectCallCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := redismock.NewClientMock()
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

			ctx := context.Background()
			callCount := 0

			fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
				callCount++
				return "result-" + args[0].(string), nil
			}

			cachedFunc := cache.Cached(
				"testFunc",
				5*time.Minute,
				false,
				"test",
			)(fn, nil)

			var result1 string
			res1, err := cachedFunc(&result1, ctx, tt.args...)
			require.NoError(t, err)
			assert.Equal(t, tt.expectCallCount, callCount)
			assert.Equal(t, tt.expectedResult, res1)

			time.Sleep(500 * time.Millisecond)
		})
	}
}
