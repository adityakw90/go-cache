package cache

import (
	"time"

	"github.com/adityakw90/go-cache/internal/key"
)

// KeyGeneratorFunc generates cache keys from a data map.
// The data map contains keys like "prefix", "namespace", "version", "key", etc.
type KeyGeneratorFunc func(data map[string]string) (string, error)

// Option modifies cache configuration.
type Option func(*options)

// options holds all cache configuration.
type options struct {
	keyPrefix           string
	expireDefault       time.Duration
	versionExpire       time.Duration
	lockDuration        time.Duration
	lockInterval        time.Duration
	semaphoreSize       int
	tracer              Tracer
	logger              Logger
	semaphore           Semaphore
	keyGenerator        KeyGeneratorFunc
	keyVersionGenerator KeyGeneratorFunc
	versionGenerator    KeyGeneratorFunc
	lockGenerator       KeyGeneratorFunc
}

// re export key generator functions
var (
	defaultKeyGenerator        = key.KeyGenerator
	defaultKeyVersionGenerator = key.KeyVersionGenerator
	defaultVersionGenerator    = key.VersionGenerator
	defaultLockGenerator       = key.LockGenerator
)

// defaultOptions returns default cache options.
func defaultOptions() *options {
	return &options{
		keyPrefix:           "CACHE",
		expireDefault:       time.Minute,
		versionExpire:       30 * 24 * time.Hour, // 30 days
		lockDuration:        time.Minute,
		lockInterval:        100 * time.Millisecond,
		semaphoreSize:       10,
		tracer:              &NoOpTracer{},
		logger:              &NoOpLogger{},
		semaphore:           nil, // Will be created from semaphoreSize
		keyGenerator:        defaultKeyGenerator,
		keyVersionGenerator: defaultKeyVersionGenerator,
		versionGenerator:    defaultVersionGenerator,
		lockGenerator:       defaultLockGenerator,
	}
}

// WithKeyPrefix sets the cache key prefix.
func WithKeyPrefix(prefix string) Option {
	return func(o *options) {
		o.keyPrefix = prefix
	}
}

// WithExpireDefault sets the default TTL for cache entries.
func WithExpireDefault(duration time.Duration) Option {
	return func(o *options) {
		o.expireDefault = duration
	}
}

// WithVersionExpire sets the TTL for version keys.
func WithVersionExpire(duration time.Duration) Option {
	return func(o *options) {
		o.versionExpire = duration
	}
}

// WithLockDuration sets the lock timeout duration.
func WithLockDuration(duration time.Duration) Option {
	return func(o *options) {
		o.lockDuration = duration
	}
}

// WithLockInterval sets the lock retry interval.
func WithLockInterval(duration time.Duration) Option {
	return func(o *options) {
		o.lockInterval = duration
	}
}

// WithSemaphoreSize sets the semaphore size for concurrency control.
func WithSemaphoreSize(size int) Option {
	return func(o *options) {
		if size <= 0 {
			size = 1
		}
		o.semaphoreSize = size
	}
}

// WithTracer sets an optional tracer for observability.
func WithTracer(tracer Tracer) Option {
	return func(o *options) {
		if tracer != nil {
			o.tracer = tracer
		}
	}
}

// WithLogger sets an optional logger for observability.
func WithLogger(logger Logger) Option {
	return func(o *options) {
		if logger != nil {
			o.logger = logger
		}
	}
}

// WithSemaphore sets an optional semaphore (defaults to channel-based).
func WithSemaphore(sem Semaphore) Option {
	return func(o *options) {
		if sem != nil {
			o.semaphore = sem
		}
	}
}

// WithKeyGenerator sets a custom key generator function.
func WithKeyGenerator(fn KeyGeneratorFunc) Option {
	return func(o *options) {
		if fn != nil {
			o.keyGenerator = fn
		}
	}
}

// WithKeyVersionGenerator sets a custom versioned key generator function.
func WithKeyVersionGenerator(fn KeyGeneratorFunc) Option {
	return func(o *options) {
		if fn != nil {
			o.keyVersionGenerator = fn
		}
	}
}

// WithVersionGenerator sets a custom version key generator function.
func WithVersionGenerator(fn KeyGeneratorFunc) Option {
	return func(o *options) {
		if fn != nil {
			o.versionGenerator = fn
		}
	}
}

// WithLockGenerator sets a custom lock key generator function.
func WithLockGenerator(fn KeyGeneratorFunc) Option {
	return func(o *options) {
		if fn != nil {
			o.lockGenerator = fn
		}
	}
}
