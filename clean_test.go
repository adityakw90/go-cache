package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_CleanCache_StandardKey(t *testing.T) {
	redisClient := createTestRedisClient(t)
	defer redisClient.Close()

	cache, err := NewCache(
		redisClient,
		WithKeyPrefix("test"),
		WithExpireDefault(5*time.Minute),
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Register a cache key by using Cached
	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"testKey",
		5*time.Minute,
		true, // versioning enabled
		"testPrefix",
	)(fn, nil)

	// Call it once to register the key
	var result string
	_, err = cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)

	// Wait for async cache update
	time.Sleep(300 * time.Millisecond)

	// Get initial version
	initialVersion, err := cache.getCacheVersion(ctx, "testKey", "testPrefix")
	require.NoError(t, err)

	// Clean cache with standard key (no custom keys, no params)
	err = cache.CleanCache(ctx, "testKey", nil, nil, true, nil)
	require.NoError(t, err)

	// Wait for pipeline execution
	time.Sleep(200 * time.Millisecond)

	// Check that version was incremented
	newVersion, err := cache.getCacheVersion(ctx, "testKey", "testPrefix")
	require.NoError(t, err)
	assert.Greater(t, newVersion, initialVersion, "Version should be incremented after CleanCache")
}

func TestCache_CleanCache_WithCustomKey(t *testing.T) {
	redisClient := createTestRedisClient(t)
	defer redisClient.Close()

	cache, err := NewCache(
		redisClient,
		WithKeyPrefix("test"),
		WithExpireDefault(5*time.Minute),
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Create custom key function
	customKey, err := NewCustomKeyFunction(
		"getUser",
		func(args ...interface{}) string {
			return "user:" + args[0].(string)
		},
		[]string{"uid"},
	)
	require.NoError(t, err)

	// Register a cache key with custom key function
	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"getUser",
		5*time.Minute,
		true, // versioning enabled
		"user",
	)(fn, customKey)

	// Call it once to register the key and custom key function
	var result string
	_, err = cachedFunc(&result, ctx, "user-123")
	require.NoError(t, err)

	// Wait for async cache update
	time.Sleep(300 * time.Millisecond)

	// Get initial version for the custom namespace
	initialVersion, err := cache.getCacheVersion(ctx, "user:user-123", "user")
	require.NoError(t, err)

	// Clean cache with custom key and params
	params := map[string]interface{}{
		"uid": "user-123",
	}
	err = cache.CleanCache(ctx, "getUser", params, nil, true, nil)
	require.NoError(t, err)

	// Wait for pipeline execution
	time.Sleep(200 * time.Millisecond)

	// Check that version was incremented
	newVersion, err := cache.getCacheVersion(ctx, "user:user-123", "user")
	require.NoError(t, err)
	assert.Greater(t, newVersion, initialVersion, "Version should be incremented after CleanCache with custom key")
}

func TestCache_CleanCache_MultiplePrefixes(t *testing.T) {
	redisClient := createTestRedisClient(t)
	defer redisClient.Close()

	cache, err := NewCache(
		redisClient,
		WithKeyPrefix("test"),
		WithExpireDefault(5*time.Minute),
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Register the same key with multiple prefixes
	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	// Register with first prefix
	cachedFunc1 := cache.Cached(
		"testKey",
		5*time.Minute,
		true,
		"prefix1",
	)(fn, nil)

	// Register with second prefix
	cachedFunc2 := cache.Cached(
		"testKey",
		5*time.Minute,
		true,
		"prefix2",
	)(fn, nil)

	// Call both to register
	var result string
	_, err = cachedFunc1(&result, ctx, "arg1")
	require.NoError(t, err)
	_, err = cachedFunc2(&result, ctx, "arg1")
	require.NoError(t, err)

	// Wait for async cache updates
	time.Sleep(300 * time.Millisecond)

	// Get initial versions
	initialVersion1, err := cache.getCacheVersion(ctx, "testKey", "prefix1")
	require.NoError(t, err)
	initialVersion2, err := cache.getCacheVersion(ctx, "testKey", "prefix2")
	require.NoError(t, err)

	// Clean cache - should affect both prefixes
	err = cache.CleanCache(ctx, "testKey", nil, nil, true, nil)
	require.NoError(t, err)

	// Wait for pipeline execution
	time.Sleep(200 * time.Millisecond)

	// Check that both versions were incremented
	newVersion1, err := cache.getCacheVersion(ctx, "testKey", "prefix1")
	require.NoError(t, err)
	newVersion2, err := cache.getCacheVersion(ctx, "testKey", "prefix2")
	require.NoError(t, err)

	assert.Greater(t, newVersion1, initialVersion1, "Version 1 should be incremented")
	assert.Greater(t, newVersion2, initialVersion2, "Version 2 should be incremented")
}

func TestCache_CleanCache_WithoutExecute(t *testing.T) {
	redisClient := createTestRedisClient(t)
	defer redisClient.Close()

	cache, err := NewCache(
		redisClient,
		WithKeyPrefix("test"),
		WithExpireDefault(5*time.Minute),
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Register a cache key
	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"testKey",
		5*time.Minute,
		true,
		"testPrefix",
	)(fn, nil)

	// Call it once to register
	var result string
	_, err = cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)

	// Wait for async cache update
	time.Sleep(300 * time.Millisecond)

	// Get initial version
	initialVersion, err := cache.getCacheVersion(ctx, "testKey", "testPrefix")
	require.NoError(t, err)

	// Create a session and use it
	session := cache.GetSession()
	var lockKeys []string

	// Clean cache without executing (execute=false)
	err = cache.CleanCache(ctx, "testKey", nil, session, false, &lockKeys)
	require.NoError(t, err)

	// Version should not be incremented yet (pipeline not executed)
	currentVersion, err := cache.getCacheVersion(ctx, "testKey", "testPrefix")
	require.NoError(t, err)
	assert.Equal(t, initialVersion, currentVersion, "Version should not change until pipeline is executed")

	// Execute the pipeline
	_, err = session.Exec(ctx)
	require.NoError(t, err)

	// Wait a bit
	time.Sleep(200 * time.Millisecond)

	// Now version should be incremented
	newVersion, err := cache.getCacheVersion(ctx, "testKey", "testPrefix")
	require.NoError(t, err)
	assert.Greater(t, newVersion, initialVersion, "Version should be incremented after pipeline execution")
}

func TestCache_CleanCache_WithLockKeys(t *testing.T) {
	redisClient := createTestRedisClient(t)
	defer redisClient.Close()

	cache, err := NewCache(
		redisClient,
		WithKeyPrefix("test"),
		WithExpireDefault(5*time.Minute),
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Register a cache key
	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"testKey",
		5*time.Minute,
		true,
		"testPrefix",
	)(fn, nil)

	// Call it once to register
	var result string
	_, err = cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)

	// Wait for async cache update
	time.Sleep(300 * time.Millisecond)

	// Clean cache and collect lock keys
	var lockKeys []string
	err = cache.CleanCache(ctx, "testKey", nil, nil, true, &lockKeys)
	require.NoError(t, err)

	// Lock keys should be collected
	assert.NotEmpty(t, lockKeys, "Lock keys should be collected")
}

func TestCache_CleanCache_MissingParamsForCustomKey(t *testing.T) {
	redisClient := createTestRedisClient(t)
	defer redisClient.Close()

	cache, err := NewCache(
		redisClient,
		WithKeyPrefix("test"),
		WithExpireDefault(5*time.Minute),
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Create custom key function that requires "uid" param
	customKey, err := NewCustomKeyFunction(
		"getUser",
		func(args ...interface{}) string {
			return "user:" + args[0].(string)
		},
		[]string{"uid"},
	)
	require.NoError(t, err)

	// Register a cache key with custom key function
	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"getUser",
		5*time.Minute,
		true,
		"user",
	)(fn, customKey)

	// Call it once to register
	var result string
	_, err = cachedFunc(&result, ctx, "user-123")
	require.NoError(t, err)

	// Wait for async cache update
	time.Sleep(300 * time.Millisecond)

	// Try to clean cache with missing params - should return error
	params := map[string]interface{}{
		// "uid" is missing
	}
	err = cache.CleanCache(ctx, "getUser", params, nil, true, nil)
	assert.Error(t, err, "Should return error when required param is missing")
	assert.Contains(t, err.Error(), "required param", "Error should mention missing param")
}

func TestCache_CleanCache_NoRegisteredKey(t *testing.T) {
	redisClient := createTestRedisClient(t)
	defer redisClient.Close()

	cache, err := NewCache(
		redisClient,
		WithKeyPrefix("test"),
		WithExpireDefault(5*time.Minute),
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Try to clean cache for a key that was never registered
	// Should return nil (no error, just no-op)
	err = cache.CleanCache(ctx, "nonExistentKey", nil, nil, true, nil)
	require.NoError(t, err, "Should not error for non-existent key, just return nil")
}

func TestCache_CleanCache_MultipleCustomKeys(t *testing.T) {
	redisClient := createTestRedisClient(t)
	defer redisClient.Close()

	cache, err := NewCache(
		redisClient,
		WithKeyPrefix("test"),
		WithExpireDefault(5*time.Minute),
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Create multiple custom key functions for the same key
	customKey1, err := NewCustomKeyFunction(
		"getUser1",
		func(args ...interface{}) string {
			return "user:" + args[0].(string)
		},
		[]string{"uid"},
	)
	require.NoError(t, err)

	customKey2, err := NewCustomKeyFunction(
		"getUser2",
		func(args ...interface{}) string {
			return "user:" + args[0].(string) + ":extra"
		},
		[]string{"uid"},
	)
	require.NoError(t, err)

	// Register cache key with first custom key
	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	cachedFunc1 := cache.Cached(
		"getUser",
		5*time.Minute,
		true,
		"user",
	)(fn, customKey1)

	// Register same key with second custom key (different name)
	cachedFunc2 := cache.Cached(
		"getUser",
		5*time.Minute,
		true,
		"user",
	)(fn, customKey2)

	// Call both to register
	var result string
	_, err = cachedFunc1(&result, ctx, "user-123")
	require.NoError(t, err)
	_, err = cachedFunc2(&result, ctx, "user-123")
	require.NoError(t, err)

	// Wait for async cache updates
	time.Sleep(300 * time.Millisecond)

	// Get initial versions for both namespaces
	initialVersion1, err := cache.getCacheVersion(ctx, "user:user-123", "user")
	require.NoError(t, err)
	initialVersion2, err := cache.getCacheVersion(ctx, "user:user-123:extra", "user")
	require.NoError(t, err)

	// Clean cache - should affect both custom key namespaces
	params := map[string]interface{}{
		"uid": "user-123",
	}
	err = cache.CleanCache(ctx, "getUser", params, nil, true, nil)
	require.NoError(t, err)

	// Wait for pipeline execution
	time.Sleep(200 * time.Millisecond)

	// Check that both versions were incremented
	newVersion1, err := cache.getCacheVersion(ctx, "user:user-123", "user")
	require.NoError(t, err)
	newVersion2, err := cache.getCacheVersion(ctx, "user:user-123:extra", "user")
	require.NoError(t, err)

	assert.Greater(t, newVersion1, initialVersion1, "Version 1 should be incremented")
	assert.Greater(t, newVersion2, initialVersion2, "Version 2 should be incremented")
}

func TestCache_CleanCache_WithExistingSession(t *testing.T) {
	redisClient := createTestRedisClient(t)
	defer redisClient.Close()

	cache, err := NewCache(
		redisClient,
		WithKeyPrefix("test"),
		WithExpireDefault(5*time.Minute),
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Register a cache key
	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"testKey",
		5*time.Minute,
		true,
		"testPrefix",
	)(fn, nil)

	// Call it once to register
	var result string
	_, err = cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)

	// Wait for async cache update
	time.Sleep(300 * time.Millisecond)

	// Get initial version
	initialVersion, err := cache.getCacheVersion(ctx, "testKey", "testPrefix")
	require.NoError(t, err)

	// Create a session and use it for CleanCache
	session := cache.GetSession()
	err = cache.CleanCache(ctx, "testKey", nil, session, true, nil)
	require.NoError(t, err)

	// Wait for pipeline execution
	time.Sleep(200 * time.Millisecond)

	// Check that version was incremented
	newVersion, err := cache.getCacheVersion(ctx, "testKey", "testPrefix")
	require.NoError(t, err)
	assert.Greater(t, newVersion, initialVersion, "Version should be incremented after CleanCache with existing session")
}

func TestCache_CleanCache_WithNilLockKeys(t *testing.T) {
	redisClient := createTestRedisClient(t)
	defer redisClient.Close()

	cache, err := NewCache(
		redisClient,
		WithKeyPrefix("test"),
		WithExpireDefault(5*time.Minute),
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Register a cache key
	fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
		return "result", nil
	}

	cachedFunc := cache.Cached(
		"testKey",
		5*time.Minute,
		true,
		"testPrefix",
	)(fn, nil)

	// Call it once to register
	var result string
	_, err = cachedFunc(&result, ctx, "arg1")
	require.NoError(t, err)

	// Wait for async cache update
	time.Sleep(300 * time.Millisecond)

	// Get initial version
	initialVersion, err := cache.getCacheVersion(ctx, "testKey", "testPrefix")
	require.NoError(t, err)

	// Clean cache with nil lockKeys - should initialize it internally
	err = cache.CleanCache(ctx, "testKey", nil, nil, true, nil)
	require.NoError(t, err)

	// Wait for pipeline execution
	time.Sleep(200 * time.Millisecond)

	// Check that version was incremented
	newVersion, err := cache.getCacheVersion(ctx, "testKey", "testPrefix")
	require.NoError(t, err)
	assert.Greater(t, newVersion, initialVersion, "Version should be incremented even with nil lockKeys")
}
