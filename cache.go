package cache

import (
	"github.com/go-redis/redis/v8"

	"github.com/adityakw90/go-cache/internal/adapter"
	internalcache "github.com/adityakw90/go-cache/internal/cache"
	"github.com/adityakw90/go-cache/internal/key"
)

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
func NewCache(redisClient *redis.Client, opts ...Option) (Cache, error) {
	// Build options from functional options
	rootOpts := defaultOptions()
	for _, opt := range opts {
		opt(rootOpts)
	}

	// Apply defaults for optional dependencies
	tracer := rootOpts.tracer
	if tracer == nil {
		tracer = adapter.NewNoOpTracer()
	}

	logger := rootOpts.logger
	if logger == nil {
		logger = adapter.NewNoOpLogger()
	}

	semaphore := rootOpts.semaphore
	if semaphore == nil {
		semaphore = adapter.NewSemaphore(rootOpts.semaphoreSize)
	}

	// Convert to internal options, converting KeyGeneratorFunc types
	internalOpts := internalcache.Options{
		KeyPrefix:           rootOpts.keyPrefix,
		ExpireDefault:       rootOpts.expireDefault,
		VersionExpire:       rootOpts.versionExpire,
		LockDuration:        rootOpts.lockDuration,
		LockInterval:        rootOpts.lockInterval,
		Tracer:              tracer,
		Logger:              logger,
		Semaphore:           semaphore,
		KeyGenerator:        convertKeyGeneratorFunc(rootOpts.keyGenerator),
		KeyVersionGenerator: convertKeyGeneratorFunc(rootOpts.keyVersionGenerator),
		VersionGenerator:    convertKeyGeneratorFunc(rootOpts.versionGenerator),
		LockGenerator:       convertKeyGeneratorFunc(rootOpts.lockGenerator),
	}

	impl, err := internalcache.NewCache(redisClient, internalOpts)
	if err != nil {
		return nil, err
	}

	return impl, nil
}

// convertKeyGeneratorFunc converts root KeyGeneratorFunc to key.KeyGeneratorFunc.
func convertKeyGeneratorFunc(fn KeyGeneratorFunc) key.KeyGeneratorFunc {
	if fn == nil {
		return nil
	}
	return key.KeyGeneratorFunc(fn)
}
