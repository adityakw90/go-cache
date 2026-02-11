package cache

import (
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/adityakw90/go-cache/adapter"
	"github.com/adityakw90/go-cache/internal/errs"
	"github.com/adityakw90/go-cache/internal/key"
)

// Cache is the internal cache implementation.
type Cache struct {
	RedisClient               *redis.Client
	LogProvider               adapter.GetLogger
	StartSpan                 adapter.StartSpan
	StartChildSpan            adapter.StartChildSpan
	Semaphore                 adapter.Semaphore
	KeyPrefix                 string
	KeyGenerator              key.KeyGeneratorFunc
	KeyVersionGenerator       key.KeyGeneratorFunc
	KeyHashedVersionGenerator key.KeyGeneratorFunc
	VersionGenerator          key.KeyGeneratorFunc
	VersionExpire             time.Duration
	LockGenerator             key.KeyGeneratorFunc
	LockDuration              time.Duration
	LockInterval              time.Duration
	ExpireDefault             time.Duration
	keyUsage                  map[string][]string                         // To track cache key usage: keyUsage[keyName] = []prefix
	customKeys                map[string]map[string]key.CustomKeyFunction // To track custom key functions
	keyMutex                  sync.Mutex                                  // Mutex to handle concurrent map access
}

// Options holds all cache configuration.
type Options struct {
	KeyPrefix                 string
	ExpireDefault             time.Duration
	VersionExpire             time.Duration
	LockDuration              time.Duration
	LockInterval              time.Duration
	LogProvider               adapter.GetLogger
	StartSpan                 adapter.StartSpan
	StartChildSpan            adapter.StartChildSpan
	Semaphore                 adapter.Semaphore
	KeyGenerator              key.KeyGeneratorFunc
	KeyVersionGenerator       key.KeyGeneratorFunc
	KeyHashedVersionGenerator key.KeyGeneratorFunc
	VersionGenerator          key.KeyGeneratorFunc
	LockGenerator             key.KeyGeneratorFunc
}

// NewCache creates a new cache instance with options.
func NewCache(redisClient *redis.Client, opts Options) (*Cache, error) {
	if redisClient == nil {
		return nil, errs.NewInvalidConfigError("redisClient", "cannot be nil")
	}
	if opts.ExpireDefault <= 0 {
		return nil, errs.NewInvalidConfigError("expireDefault", "must be greater than zero")
	}
	if opts.VersionExpire <= 0 {
		return nil, errs.NewInvalidConfigError("versionExpire", "must be greater than zero")
	}
	if opts.LockDuration <= 0 {
		return nil, errs.NewInvalidConfigError("lockDuration", "must be greater than zero")
	}
	if opts.LockInterval <= 0 {
		return nil, errs.NewInvalidConfigError("lockInterval", "must be greater than zero")
	}

	// validate hooks
	if opts.StartSpan == nil {
		return nil, errs.NewInvalidConfigError("startSpan", "cannot be nil")
	}
	if opts.StartChildSpan == nil {
		return nil, errs.NewInvalidConfigError("startChildSpan", "cannot be nil")
	}
	if opts.Semaphore == nil {
		return nil, errs.NewInvalidConfigError("semaphore", "cannot be nil")
	}

	// validate generators
	if opts.KeyGenerator == nil {
		return nil, errs.NewInvalidConfigError("keyGenerator", "cannot be nil")
	}
	if opts.KeyVersionGenerator == nil {
		return nil, errs.NewInvalidConfigError("keyVersionGenerator", "cannot be nil")
	}
	if opts.KeyHashedVersionGenerator == nil {
		return nil, errs.NewInvalidConfigError("keyHashedVersionGenerator", "cannot be nil")
	}
	if opts.VersionGenerator == nil {
		return nil, errs.NewInvalidConfigError("versionGenerator", "cannot be nil")
	}
	if opts.LockGenerator == nil {
		return nil, errs.NewInvalidConfigError("lockGenerator", "cannot be nil")
	}

	// hooks
	if opts.LogProvider == nil {
		return nil, errs.NewInvalidConfigError("logProvider", "cannot be nil")
	}

	c := &Cache{
		RedisClient:               redisClient,
		LogProvider:               opts.LogProvider,
		StartSpan:                 opts.StartSpan,
		StartChildSpan:            opts.StartChildSpan,
		Semaphore:                 opts.Semaphore,
		KeyPrefix:                 opts.KeyPrefix,
		KeyGenerator:              opts.KeyGenerator,
		KeyVersionGenerator:       opts.KeyVersionGenerator,
		KeyHashedVersionGenerator: opts.KeyHashedVersionGenerator,
		VersionGenerator:          opts.VersionGenerator,
		VersionExpire:             opts.VersionExpire,
		LockGenerator:             opts.LockGenerator,
		LockDuration:              opts.LockDuration,
		LockInterval:              opts.LockInterval,
		ExpireDefault:             opts.ExpireDefault,
		keyUsage:                  make(map[string][]string),
		customKeys:                make(map[string]map[string]key.CustomKeyFunction),
	}

	return c, nil
}
