package cache

import (
	"testing"

	"github.com/adityakw90/go-cache/internal/errs"
	"github.com/go-redis/redis/v8"
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
		checkFunc func(t *testing.T, cache *Cache, err error)
	}{
		{
			name:    "valid redis client",
			client:  redisClient,
			wantErr: false,
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				assert.NotNil(t, cache)
				assert.NotNil(t, cache.redisClient)
				assert.NotNil(t, cache.tracer)
				assert.NotNil(t, cache.logger)
				assert.NotNil(t, cache.semaphore)
				assert.Equal(t, "CACHE", cache.keyPrefix)
				assert.NotNil(t, cache.keyUsage)
				assert.NotNil(t, cache.customKeys)
			},
		},
		{
			name:    "nil redis client",
			client:  nil,
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

	tracer := &NoOpTracer{}
	logger := &NoOpLogger{}

	tests := []struct {
		name      string
		options   []Option
		checkFunc func(t *testing.T, cache *Cache, err error)
	}{
		{
			name: "with custom options",
			options: []Option{
				WithKeyPrefix("myapp"),
				WithTracer(tracer),
				WithLogger(logger),
				WithSemaphoreSize(20),
			},
			checkFunc: func(t *testing.T, cache *Cache, err error) {
				require.NoError(t, err)
				assert.Equal(t, "myapp", cache.keyPrefix)
				assert.Equal(t, tracer, cache.tracer)
				assert.Equal(t, logger, cache.logger)
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

func TestCache_GetSession(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	tests := []struct {
		name      string
		checkFunc func(t *testing.T, session interface{})
	}{
		{
			name: "get session returns non-nil",
			checkFunc: func(t *testing.T, session interface{}) {
				assert.NotNil(t, session)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewCache(redisClient)
			require.NoError(t, err)

			session := cache.GetSession()
			if tt.checkFunc != nil {
				tt.checkFunc(t, session)
			}
		})
	}
}

func TestCache_registerCacheKey(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	tests := []struct {
		name     string
		keyName  string
		register []struct {
			keyName string
			prefix  string
		}
		checkFunc func(t *testing.T, prefixes []string)
	}{
		{
			name:    "register multiple prefixes for same key",
			keyName: "getUser",
			register: []struct {
				keyName string
				prefix  string
			}{
				{"getUser", "user"},
				{"getUser", "admin"},
			},
			checkFunc: func(t *testing.T, prefixes []string) {
				assert.Contains(t, prefixes, "user")
				assert.Contains(t, prefixes, "admin")
				assert.Equal(t, 2, len(prefixes))
			},
		},
		{
			name:    "duplicate prefix registration",
			keyName: "getUser",
			register: []struct {
				keyName string
				prefix  string
			}{
				{"getUser", "user"},
				{"getUser", "user"}, // Duplicate
			},
			checkFunc: func(t *testing.T, prefixes []string) {
				assert.Equal(t, 1, len(prefixes))
				assert.Contains(t, prefixes, "user")
			},
		},
		{
			name:    "register different keys with same prefix",
			keyName: "getUser",
			register: []struct {
				keyName string
				prefix  string
			}{
				{"getUser", "user"},
				{"listUser", "user"},
			},
			checkFunc: func(t *testing.T, prefixes []string) {
				// getUser should only have "user" prefix
				assert.Equal(t, 1, len(prefixes))
				assert.Contains(t, prefixes, "user")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewCache(redisClient)
			require.NoError(t, err)

			for _, reg := range tt.register {
				cache.registerCacheKey(reg.keyName, reg.prefix)
			}
			prefixes := cache.GetCacheKeyUsage(tt.keyName)
			if tt.checkFunc != nil {
				tt.checkFunc(t, prefixes)
			}
		})
	}
}

func TestCache_registerCustomKey(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	tests := []struct {
		name      string
		funcName  string
		customKey map[string]interface{}
		checkFunc func(t *testing.T, exists bool, customKeys map[string]CustomKeyFunction)
	}{
		{
			name:     "register valid custom key",
			funcName: "getUser",
			customKey: map[string]interface{}{
				"name": "getUser",
				"callable": func(args ...interface{}) string {
					return "user:" + args[0].(string)
				},
				"params": []string{"uid"},
			},
			checkFunc: func(t *testing.T, exists bool, customKeys map[string]CustomKeyFunction) {
				assert.True(t, exists)
				assert.NotNil(t, customKeys["getUser"])
			},
		},
		{
			name:      "register nil custom key",
			funcName:  "getUser2",
			customKey: nil,
			checkFunc: func(t *testing.T, exists bool, customKeys map[string]CustomKeyFunction) {
				assert.False(t, exists)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewCache(redisClient)
			require.NoError(t, err)

			var customKey CustomKeyFunction
			if tt.customKey != nil {
				var err error
				customKey, err = NewCustomKeyFunction(
					tt.funcName,
					tt.customKey["callable"].(func(args ...interface{}) string),
					tt.customKey["params"].([]string),
				)
				require.NoError(t, err)
			}
			cache.registerCustomKey(tt.funcName, customKey)

			cache.keyMutex.Lock()
			customKeys, exists := cache.customKeys[tt.funcName]
			cache.keyMutex.Unlock()

			if tt.checkFunc != nil {
				tt.checkFunc(t, exists, customKeys)
			}
		})
	}
}

func TestCache_GetCacheKeyUsage(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	cache, err := NewCache(redisClient)
	require.NoError(t, err)

	tests := []struct {
		name      string
		keyName   string
		setup     func()
		checkFunc func(t *testing.T, prefixes []string)
	}{
		{
			name:    "non-existent key",
			keyName: "nonexistent",
			setup:   func() {},
			checkFunc: func(t *testing.T, prefixes []string) {
				assert.Empty(t, prefixes)
			},
		},
		{
			name:    "key with single prefix",
			keyName: "getUser",
			setup: func() {
				cache.registerCacheKey("getUser", "user")
			},
			checkFunc: func(t *testing.T, prefixes []string) {
				assert.Equal(t, 1, len(prefixes))
				assert.Contains(t, prefixes, "user")
			},
		},
		{
			name:    "key with multiple prefixes",
			keyName: "getUser",
			setup: func() {
				cache.registerCacheKey("getUser", "user")
				cache.registerCacheKey("getUser", "admin")
				cache.registerCacheKey("getUser", "public")
			},
			checkFunc: func(t *testing.T, prefixes []string) {
				assert.Equal(t, 3, len(prefixes))
				assert.Contains(t, prefixes, "user")
				assert.Contains(t, prefixes, "admin")
				assert.Contains(t, prefixes, "public")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset cache for each test
			cache, err = NewCache(redisClient)
			require.NoError(t, err)

			if tt.setup != nil {
				tt.setup()
			}

			prefixes := cache.GetCacheKeyUsage(tt.keyName)
			if tt.checkFunc != nil {
				tt.checkFunc(t, prefixes)
			}
		})
	}
}

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
