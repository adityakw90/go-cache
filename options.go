package cache

import (
	"context"
	"time"

	"github.com/adityakw90/go-cache/adapter"
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
	semaphore           adapter.Semaphore
	keyGenerator        KeyGeneratorFunc
	keyVersionGenerator KeyGeneratorFunc
	versionGenerator    KeyGeneratorFunc
	lockGenerator       KeyGeneratorFunc
	getLogger           adapter.GetLogger
	startSpan           adapter.StartSpan
	startChildSpan      adapter.StartChildSpan
}

// defaultOptions returns default cache options.
func defaultOptions() *options {
	return &options{
		keyPrefix:           "CACHE",
		expireDefault:       time.Minute,
		versionExpire:       30 * 24 * time.Hour, // 30 days
		lockDuration:        time.Minute,
		lockInterval:        100 * time.Millisecond,
		semaphoreSize:       10,
		semaphore:           nil, // Will be created from semaphoreSize
		keyGenerator:        key.KeyGenerator,
		keyVersionGenerator: key.KeyVersionGenerator,
		versionGenerator:    key.VersionGenerator,
		lockGenerator:       key.LockGenerator,
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

// WithSemaphore sets an optional semaphore (defaults to channel-based).
func WithSemaphore(sem adapter.Semaphore) Option {
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

// WithLogProvider sets an optional log provider function.
func WithLogProvider(fn func(ctx context.Context) adapter.Logger) Option {
	return func(o *options) {
		if fn != nil {
			o.getLogger = fn
		}
	}
}

// WithTraceProvider sets an optional trace provider function.
func WithTraceProvider(
	fnStartSpan adapter.StartSpan,
	fnStartChildSpan adapter.StartChildSpan,
) Option {
	return func(o *options) {
		if fnStartSpan != nil {
			o.startSpan = fnStartSpan
		}
		if fnStartChildSpan != nil {
			o.startChildSpan = fnStartChildSpan
		}
	}
}
