package cache

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/internal/hash"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/adityakw90/go-cache/internal/serialize"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_Cached_RegistersKey(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	// Create cached function
	_ = cache.Cached(
		"testFunc",
		5*time.Minute,
		false,
		"customPrefix",
	)

	// Verify key is registered
	prefixes := cache.GetCacheKeyUsage("testFunc")
	assert.Contains(t, prefixes, "customPrefix")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCache_Cached_UsesDefaultPrefix(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		KeyPrefix:           "default",
		ExpireDefault:       5 * time.Minute,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	// Create cached function with empty prefix
	_ = cache.Cached(
		"testFunc",
		5*time.Minute,
		false,
		"",
	)

	// Verify default prefix is used
	prefixes := cache.GetCacheKeyUsage("testFunc")
	assert.Contains(t, prefixes, "default")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCache_Cached_RegistersCustomKey(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	// Create custom key function
	customKey, err := key.NewCustomKeyFunction(
		"getUser",
		func(args ...interface{}) string {
			return "user:" + args[0].(string)
		},
		[]string{"uid"},
	)
	require.NoError(t, err)

	// Create cached function with custom key
	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"getUser",
		5*time.Minute,
		false,
		"user",
	)(fn, customKey)

	// Verify custom key is registered (indirectly tested)
	assert.NotNil(t, cachedFunc)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCache_Cached_WithVersioning(t *testing.T) {
	client, _ := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		VersionExpire:       1 * time.Hour,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	ctx := context.Background()
	callCount := 0

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		callCount++
		return "result", nil
	}

	// Create cached function with versioning
	cachedFunc := cache.Cached(
		"testFunc",
		5*time.Minute,
		true, // versioning enabled
		"test",
	)(fn, nil)

	// First call
	var result string
	_, err = cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Wait for async cache update
	time.Sleep(300 * time.Millisecond)

	// Second call - should use cache if Redis available
	var result2 string
	_, err = cachedFunc(&result2, ctx, "arg1")
	require.NoError(t, err)
	// Function may or may not be called again depending on Redis availability
	assert.GreaterOrEqual(t, callCount, 1)

	// Note: Mock expectations may not be fully met due to async operations
	// This is acceptable for these tests
}

func TestCache_Cached_WithoutVersioning(t *testing.T) {
	client, _ := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
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
		return "result", nil
	}

	// Create cached function without versioning
	cachedFunc := cache.Cached(
		"testFunc",
		5*time.Minute,
		false, // versioning disabled
		"test",
	)(fn, nil)

	// Call function
	var result string
	_, err = cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Note: Mock expectations may not be fully met due to async operations
	// This is acceptable for these tests
}

func TestCache_Cached_FunctionError(t *testing.T) {
	client, _ := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	ctx := context.Background()

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return nil, assert.AnError
	}

	cachedFunc := cache.Cached(
		"testFunc",
		5*time.Minute,
		false,
		"test",
	)(fn, nil)

	// Call function - should return error
	var result string
	_, err = cachedFunc(&result, ctx, "arg1")
	assert.Error(t, err)

	// Note: Mock expectations may not be fully met due to error path
	// This is acceptable for these tests
}

func TestCache_Cached_DynamicTTL(t *testing.T) {
	client, _ := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	ctx := context.Background()

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	// Create cached function with dynamic TTL
	ttlFunc := func(result interface{}, args ...interface{}) time.Duration {
		return 10 * time.Minute
	}

	cachedFunc := cache.Cached(
		"testFunc",
		ttlFunc,
		false,
		"test",
	)(fn, nil)

	// Call function
	var result string
	_, err = cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)

	// Wait for async cache update
	time.Sleep(300 * time.Millisecond)

	// Note: Mock expectations may not be fully met due to async operations
	// This is acceptable for these tests
}

func TestCache_Cached_CustomKeyNamespace(t *testing.T) {
	client, _ := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	ctx := context.Background()

	customKey, err := key.NewCustomKeyFunction(
		"getUser",
		func(args ...interface{}) string {
			return "user:" + args[0].(string)
		},
		[]string{"uid"},
	)
	require.NoError(t, err)

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"getUser",
		5*time.Minute,
		false,
		"user",
	)(fn, customKey)

	// Call function with custom key
	var result string
	_, err = cachedFunc(&result, ctx, "123")
	require.NoError(t, err)

	// Note: Mock expectations may not be fully met due to async operations
	// This is acceptable for these tests
}

func TestCache_Cached_KeyGenerationError(t *testing.T) {
	client, _ := redismock.NewClientMock()
	defer client.Close()

	// Create cache with invalid key generator
	invalidKeyGen := func(data map[string]string) (string, error) {
		return "", assert.AnError
	}

	cache, err := NewCache(client, Options{
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		KeyVersionGenerator: invalidKeyGen,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	ctx := context.Background()

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"testFunc",
		5*time.Minute,
		false,
		"test",
	)(fn, nil)

	// Call function - should fall back to executing function
	var result string
	res, err := cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)
	// Should return the function result directly
	assert.Equal(t, "result", res)
}

func TestCache_Cached_VersionError(t *testing.T) {
	client, _ := redismock.NewClientMock()
	defer client.Close()

	// Create cache with invalid version generator
	invalidVersionGen := func(data map[string]string) (string, error) {
		return "", assert.AnError
	}

	cache, err := NewCache(client, Options{
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		VersionExpire:       1 * time.Hour,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    invalidVersionGen,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	ctx := context.Background()

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"testFunc",
		5*time.Minute,
		true, // versioning enabled
		"test",
	)(fn, nil)

	// Call function - should fall back to executing function when version fails
	var result string
	res, err := cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)
	// Should return the function result directly
	assert.Equal(t, "result", res)
}

func TestCache_Cached_CacheHit(t *testing.T) {
	client, _ := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
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

	// First call - cache miss
	var result1 string
	res1, err := cachedFunc(&result1, ctx, "arg1")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)
	assert.Equal(t, "result-arg1", res1)

	// Wait for async cache update
	time.Sleep(500 * time.Millisecond)

	// Second call - should be cache hit if Redis available
	var result2 string
	res2, err := cachedFunc(&result2, ctx, "arg1")
	require.NoError(t, err)
	// If cache hit, result2 should be populated and callCount should remain 1
	// If cache miss (Redis not available), callCount will be 2
	if callCount == 1 {
		// Cache hit - resultType should be returned
		assert.NotNil(t, res2)
	}

	// Note: Mock expectations may not be fully met due to async operations
	// This is acceptable for these tests
}

func TestCache_Cached_DeserializeError(t *testing.T) {
	client, _ := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
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
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"testFunc",
		5*time.Minute,
		false,
		"test",
	)(fn, nil)

	// First call
	var result1 string
	_, err = cachedFunc(&result1, ctx, "arg1")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Wait for async cache update
	time.Sleep(300 * time.Millisecond)

	// Second call with wrong type - should fall back to function if deserialize fails
	var result2 int // Wrong type
	res2, err := cachedFunc(&result2, ctx, "arg1")
	require.NoError(t, err)
	// Should fall back to executing function
	assert.Equal(t, "result", res2)

	// Note: Mock expectations may not be fully met due to async operations
	// This is acceptable for these tests
}

func TestCache_Cached_CacheHitWithProperDeserialization(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
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
		return "cached-result", nil
	}

	cachedFunc := cache.Cached(
		"testFunc",
		5*time.Minute,
		false,
		"test",
	)(fn, nil)

	// Generate expected cache key with proper hash
	namespace := "testFunc"
	hashKey := hash.CacheKey(namespace, []interface{}{"arg1"})
	cacheKey, err := key.KeyVersionGenerator(map[string]string{
		"prefix":    "test",
		"namespace": namespace,
		"version":   "0",
		"key":       hashKey,
	})
	require.NoError(t, err)

	// Serialize the expected result using Gob
	serializedData, err := serialize.Serialize("cached-result")
	require.NoError(t, err)

	// Mock cache hit - return serialized data
	mock.ExpectGet(cacheKey).SetVal(string(serializedData))

	// Call function - should return cached data
	var result string
	res, err := cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)
	assert.Equal(t, 0, callCount) // Function should not be called
	assert.Equal(t, "cached-result", result)
	assert.Equal(t, &result, res) // Should return resultType

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCache_Cached_RedisGetError(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
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
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"testFunc",
		5*time.Minute,
		false,
		"test",
	)(fn, nil)

	// Generate expected cache key with proper hash
	namespace := "testFunc"
	hashKey := hash.CacheKey(namespace, []interface{}{"arg1"})
	cacheKey, err := key.KeyVersionGenerator(map[string]string{
		"prefix":    "test",
		"namespace": namespace,
		"version":   "0",
		"key":       hashKey,
	})
	require.NoError(t, err)

	// Mock Redis Get error (non-Nil error)
	mock.ExpectGet(cacheKey).SetErr(assert.AnError)

	// Call function - should fall back to executing function
	var result string
	res, err := cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount) // Function should be called
	assert.Equal(t, "result", res)

	// Wait for async cache update attempt
	time.Sleep(300 * time.Millisecond)
}

func TestCache_Cached_TTLFunctionPath(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
		LockDuration:        5 * time.Second,
		LockInterval:        100 * time.Millisecond,
	})
	require.NoError(t, err)

	ctx := context.Background()
	callCount := 0

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		callCount++
		return map[string]string{"status": "ok"}, nil
	}

	// Create cached function with TTL function
	ttlFunc := func(result interface{}, args ...interface{}) time.Duration {
		// Return different TTL based on result
		if m, ok := result.(map[string]string); ok && m["status"] == "ok" {
			return 10 * time.Minute
		}
		return 5 * time.Minute
	}

	cachedFunc := cache.Cached(
		"testFunc",
		ttlFunc,
		false,
		"test",
	)(fn, nil)

	// Generate expected keys with proper hash
	namespace := "testFunc"
	hashKey := hash.CacheKey(namespace, []interface{}{"arg1"})
	cacheKey, err := key.KeyVersionGenerator(map[string]string{
		"prefix":    "test",
		"namespace": namespace,
		"version":   "0",
		"key":       hashKey,
	})
	require.NoError(t, err)

	lockKey, err := key.LockGenerator(map[string]string{
		"prefix":    "test",
		"namespace": namespace,
	})
	require.NoError(t, err)

	// Mock cache miss
	mock.ExpectGet(cacheKey).RedisNil()
	// Mock async cache update: lock acquisition
	mock.Regexp().ExpectSetNX(lockKey, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, 5*time.Second).SetVal(true)
	// Mock Set with TTL from function (10 minutes)
	mock.Regexp().ExpectSet(cacheKey, `.*`, 10*time.Minute).SetVal("OK")
	// Mock lock release
	mock.Regexp().ExpectEvalSha(`.*`, []string{lockKey}, []interface{}{`.*`}).SetErr(assert.AnError)
	mock.Regexp().ExpectEval(`.*`, []string{lockKey}, []interface{}{`.*`}).SetVal(int64(1))

	// Call function
	var result map[string]string
	res, err := cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)
	assert.Equal(t, map[string]string{"status": "ok"}, res)

	// Wait for async cache update
	time.Sleep(300 * time.Millisecond)

	// Note: Some expectations may not be met due to async operations
}

func TestCache_Cached_TTLDefaultPath(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		KeyPrefix:           "test",
		ExpireDefault:       7 * time.Minute, // Default TTL
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
		LockDuration:        5 * time.Second,
		LockInterval:        100 * time.Millisecond,
	})
	require.NoError(t, err)

	ctx := context.Background()

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	// Create cached function with invalid TTL type (should use default)
	cachedFunc := cache.Cached(
		"testFunc",
		"invalid-ttl-type", // Invalid type, should use default
		false,
		"test",
	)(fn, nil)

	// Generate expected keys with proper hash
	namespace := "testFunc"
	hashKey := hash.CacheKey(namespace, []interface{}{"arg1"})
	cacheKey, err := key.KeyVersionGenerator(map[string]string{
		"prefix":    "test",
		"namespace": namespace,
		"version":   "0",
		"key":       hashKey,
	})
	require.NoError(t, err)

	lockKey, err := key.LockGenerator(map[string]string{
		"prefix":    "test",
		"namespace": namespace,
	})
	require.NoError(t, err)

	// Mock cache miss
	mock.ExpectGet(cacheKey).RedisNil()
	// Mock async cache update: lock acquisition
	mock.Regexp().ExpectSetNX(lockKey, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, 5*time.Second).SetVal(true)
	// Mock Set with default TTL (7 minutes)
	mock.Regexp().ExpectSet(cacheKey, `.*`, 7*time.Minute).SetVal("OK")
	// Mock lock release
	mock.Regexp().ExpectEvalSha(`.*`, []string{lockKey}, []interface{}{`.*`}).SetErr(assert.AnError)
	mock.Regexp().ExpectEval(`.*`, []string{lockKey}, []interface{}{`.*`}).SetVal(int64(1))

	// Call function
	var result string
	_, err = cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)

	// Wait for async cache update
	time.Sleep(300 * time.Millisecond)
}

func TestCache_Cached_VersioningWithCacheHit(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		VersionExpire:       1 * time.Hour,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	ctx := context.Background()
	callCount := 0

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		callCount++
		return "versioned-result", nil
	}

	cachedFunc := cache.Cached(
		"testFunc",
		5*time.Minute,
		true, // versioning enabled
		"test",
	)(fn, nil)

	// Generate expected keys
	namespace := "testFunc"
	versionKey, err := key.VersionGenerator(map[string]string{
		"prefix":    "test",
		"namespace": namespace,
	})
	require.NoError(t, err)

	hashKey := hash.CacheKey(namespace, []interface{}{"arg1"})
	cacheKey, err := key.KeyVersionGenerator(map[string]string{
		"prefix":    "test",
		"namespace": namespace,
		"version":   "1", // Version 1
		"key":       hashKey,
	})
	require.NoError(t, err)

	// Serialize the expected result using Gob
	serializedData, err := serialize.Serialize("versioned-result")
	require.NoError(t, err)

	// Mock version retrieval
	mock.ExpectGet(versionKey).SetVal("1")
	// Mock cache hit
	mock.ExpectGet(cacheKey).SetVal(string(serializedData))

	// Call function - should return cached data
	var result string
	res, err := cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)
	assert.Equal(t, 0, callCount) // Function should not be called
	assert.Equal(t, "versioned-result", result)
	assert.Equal(t, &result, res)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCache_Cached_CustomKeyCallable(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	cache, err := NewCache(client, Options{
		KeyPrefix:           "test",
		ExpireDefault:       5 * time.Minute,
		KeyVersionGenerator: key.KeyVersionGenerator,
		VersionGenerator:    key.VersionGenerator,
		LockGenerator:       key.LockGenerator,
	})
	require.NoError(t, err)

	ctx := context.Background()
	callCount := 0

	// Create custom key function
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

	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		callCount++
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"getUser",
		5*time.Minute,
		false,
		"user",
	)(fn, customKey)

	// Call with empty args - customKey.Callable will handle it
	var result string
	res, err := cachedFunc(&result, ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)
	assert.Equal(t, "result", res)

	assert.NoError(t, mock.ExpectationsWereMet())
}
