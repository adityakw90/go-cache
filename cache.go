package cache

import (
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cache is the main cache structure.
type Cache struct {
	redisClient         *redis.Client
	tracer              Tracer
	logger              Logger
	semaphore           Semaphore
	keyPrefix           string
	keyGenerator        KeyGeneratorFunc
	keyVersionGenerator KeyGeneratorFunc
	versionGenerator    KeyGeneratorFunc
	versionExpire       time.Duration
	lockGenerator       KeyGeneratorFunc
	lockDuration        time.Duration
	lockInterval        time.Duration
	expireDefault       time.Duration
	keyUsage            map[string][]string                      // To track cache key usage
	customKeys          map[string]map[string]*CustomKeyFunction // To track custom key functions
	keyMutex            sync.Mutex                               // Mutex to handle concurrent map access
}

// NewCache creates a new cache instance with functional options.
//
// Example:
//
//	cache, err := cache.NewCache(
//	    redisClient,
//	    cache.WithKeyPrefix("myapp"),
//	    cache.WithExpireDefault(5 * time.Minute),
//	    cache.WithTracer(myTracer),
//	    cache.WithLogger(myLogger),
//	)
func NewCache(redisClient *redis.Client, opts ...Option) (*Cache, error) {
	if redisClient == nil {
		return nil, &ErrInvalidConfig{Message: "redis client cannot be nil"}
	}

	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Use provided semaphore or create default
	semaphore := options.semaphore
	if semaphore == nil {
		semaphore = NewDefaultSemaphore(options.semaphoreSize)
	}

	// Use provided tracer or default to no-op
	tracer := options.tracer
	if tracer == nil {
		tracer = &NoOpTracer{}
	}

	// Use provided logger or default to no-op
	logger := options.logger
	if logger == nil {
		logger = &NoOpLogger{}
	}

	c := &Cache{
		redisClient:         redisClient,
		tracer:              tracer,
		logger:              logger,
		semaphore:           semaphore,
		keyPrefix:           options.keyPrefix,
		keyGenerator:        options.keyGenerator,
		keyVersionGenerator: options.keyVersionGenerator,
		versionGenerator:    options.versionGenerator,
		versionExpire:       options.versionExpire,
		lockGenerator:       options.lockGenerator,
		lockDuration:        options.lockDuration,
		lockInterval:        options.lockInterval,
		expireDefault:       options.expireDefault,
		keyUsage:            make(map[string][]string),
		customKeys:          make(map[string]map[string]*CustomKeyFunction),
	}

	return c, nil
}

// GetSession returns a Redis pipeline session.
func (c *Cache) GetSession() redis.Pipeliner {
	return c.redisClient.Pipeline()
}

// registerCacheKey registers a cache key for tracking.
func (c *Cache) registerCacheKey(keyName string, prefix string) {
	c.keyMutex.Lock()
	defer c.keyMutex.Unlock()

	// Check if the key is already registered
	if _, exists := c.keyUsage[prefix]; !exists {
		c.keyUsage[prefix] = []string{}
	}

	// Check if keyName already exists in the list
	found := false
	for _, key := range c.keyUsage[prefix] {
		if key == keyName {
			found = true
			break
		}
	}

	if !found {
		c.keyUsage[prefix] = append(c.keyUsage[prefix], keyName)
	}
}

// registerCustomKey registers a custom key function.
func (c *Cache) registerCustomKey(keyName string, customKeyFunc *CustomKeyFunction) {
	if customKeyFunc == nil {
		return
	}

	c.keyMutex.Lock()
	defer c.keyMutex.Unlock()

	if _, exists := c.customKeys[keyName]; !exists {
		c.customKeys[keyName] = make(map[string]*CustomKeyFunction)
	}

	c.customKeys[keyName][customKeyFunc.Name] = customKeyFunc
}

// getCacheKeyUsage returns the list of key names for a given prefix.
func (c *Cache) getCacheKeyUsage(prefix string) []string {
	c.keyMutex.Lock()
	defer c.keyMutex.Unlock()

	if keys, exists := c.keyUsage[prefix]; exists {
		// Return a copy to prevent external modification
		result := make([]string, len(keys))
		copy(result, keys)
		return result
	}

	return []string{}
}

// ErrInvalidConfig represents an invalid configuration error.
type ErrInvalidConfig struct {
	Message string
}

func (e *ErrInvalidConfig) Error() string {
	return e.Message
}
