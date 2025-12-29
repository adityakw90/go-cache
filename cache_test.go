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
		name      string
		funcName  string
		namespace string
		register  []struct {
			funcName  string
			namespace string
		}
		checkFunc func(t *testing.T, keys []string)
	}{
		{
			name:      "register multiple keys",
			funcName:  "getUser",
			namespace: "user",
			register: []struct {
				funcName  string
				namespace string
			}{
				{"getUser", "user"},
				{"listUser", "user"},
			},
			checkFunc: func(t *testing.T, keys []string) {
				assert.Contains(t, keys, "getUser")
				assert.Contains(t, keys, "listUser")
			},
		},
		{
			name:      "duplicate registration",
			funcName:  "getUser",
			namespace: "user2",
			register: []struct {
				funcName  string
				namespace string
			}{
				{"getUser", "user2"},
				{"getUser", "user2"}, // Duplicate
			},
			checkFunc: func(t *testing.T, keys []string) {
				assert.Equal(t, 1, len(keys))
				assert.Contains(t, keys, "getUser")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewCache(redisClient)
			require.NoError(t, err)

			for _, reg := range tt.register {
				cache.registerCacheKey(reg.funcName, reg.namespace)
			}
			keys := cache.getCacheKeyUsage(tt.namespace)
			if tt.checkFunc != nil {
				tt.checkFunc(t, keys)
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
		customKey *CustomKeyFunction
		checkFunc func(t *testing.T, exists bool, customKeys map[string]*CustomKeyFunction)
	}{
		{
			name:     "register valid custom key",
			funcName: "getUser",
			customKey: &CustomKeyFunction{
				Name: "getUser",
				Callable: func(args ...interface{}) string {
					return "user:" + args[0].(string)
				},
				Params: []string{"uid"},
			},
			checkFunc: func(t *testing.T, exists bool, customKeys map[string]*CustomKeyFunction) {
				assert.True(t, exists)
				assert.NotNil(t, customKeys["getUser"])
			},
		},
		{
			name:      "register nil custom key",
			funcName:  "getUser2",
			customKey: nil,
			checkFunc: func(t *testing.T, exists bool, customKeys map[string]*CustomKeyFunction) {
				assert.False(t, exists)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewCache(redisClient)
			require.NoError(t, err)

			cache.registerCustomKey(tt.funcName, tt.customKey)

			cache.keyMutex.Lock()
			customKeys, exists := cache.customKeys[tt.funcName]
			cache.keyMutex.Unlock()

			if tt.checkFunc != nil {
				tt.checkFunc(t, exists, customKeys)
			}
		})
	}
}

func TestCache_getCacheKeyUsage(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	cache, err := NewCache(redisClient)
	require.NoError(t, err)

	tests := []struct {
		name      string
		namespace string
		checkFunc func(t *testing.T, keys []string)
	}{
		{
			name:      "non-existent namespace",
			namespace: "nonexistent",
			checkFunc: func(t *testing.T, keys []string) {
				assert.Empty(t, keys)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := cache.getCacheKeyUsage(tt.namespace)
			if tt.checkFunc != nil {
				tt.checkFunc(t, keys)
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
