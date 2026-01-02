package cache

import (
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/errs"
	"github.com/adityakw90/go-cache/internal/key"
)

// Cache is the internal cache implementation.
type Cache struct {
	RedisClient         *redis.Client
	Tracer              adapter.Tracer
	Logger              adapter.Logger
	Semaphore           adapter.Semaphore
	KeyPrefix           string
	KeyGenerator        key.KeyGeneratorFunc
	KeyVersionGenerator key.KeyGeneratorFunc
	VersionGenerator    key.KeyGeneratorFunc
	VersionExpire       time.Duration
	LockGenerator       key.KeyGeneratorFunc
	LockDuration        time.Duration
	LockInterval        time.Duration
	ExpireDefault       time.Duration
	keyUsage            map[string][]string                         // To track cache key usage: keyUsage[keyName] = []prefix
	customKeys          map[string]map[string]key.CustomKeyFunction // To track custom key functions
	keyMutex            sync.Mutex                                  // Mutex to handle concurrent map access
}

// Options holds all cache configuration.
type Options struct {
	KeyPrefix           string
	ExpireDefault       time.Duration
	VersionExpire       time.Duration
	LockDuration        time.Duration
	LockInterval        time.Duration
	SemaphoreSize       int
	Tracer              adapter.Tracer
	Logger              adapter.Logger
	Semaphore           adapter.Semaphore
	KeyGenerator        key.KeyGeneratorFunc
	KeyVersionGenerator key.KeyGeneratorFunc
	VersionGenerator    key.KeyGeneratorFunc
	LockGenerator       key.KeyGeneratorFunc
}

// NewCache creates a new cache instance with options.
func NewCache(redisClient *redis.Client, opts Options) (*Cache, error) {
	if redisClient == nil {
		return nil, errs.NewInvalidConfigError("redisClient", "cannot be nil")
	}

	// Use provided semaphore or create default
	semaphore := opts.Semaphore
	if semaphore == nil {
		semaphore = adapter.NewSemaphore(opts.SemaphoreSize)
	}

	// Use provided tracer or default to no-op
	tracer := opts.Tracer
	if tracer == nil {
		tracer = adapter.NewNoOpTracer()
	}

	// Use provided logger or default to no-op
	logger := opts.Logger
	if logger == nil {
		logger = adapter.NewNoOpLogger()
	}

	c := &Cache{
		RedisClient:         redisClient,
		Tracer:              tracer,
		Logger:              logger,
		Semaphore:           semaphore,
		KeyPrefix:           opts.KeyPrefix,
		KeyGenerator:        opts.KeyGenerator,
		KeyVersionGenerator: opts.KeyVersionGenerator,
		VersionGenerator:    opts.VersionGenerator,
		VersionExpire:       opts.VersionExpire,
		LockGenerator:       opts.LockGenerator,
		LockDuration:        opts.LockDuration,
		LockInterval:        opts.LockInterval,
		ExpireDefault:       opts.ExpireDefault,
		keyUsage:            make(map[string][]string),
		customKeys:          make(map[string]map[string]key.CustomKeyFunction),
	}

	return c, nil
}
