package cache

import (
	"fmt"
	"time"
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

// defaultKeyGenerator generates a simple cache key.
// Template: {prefix}:{key}
func defaultKeyGenerator(data map[string]string) (string, error) {
	prefix, ok := data["prefix"]
	if !ok {
		return "", fmt.Errorf("missing 'prefix' in key generator data")
	}
	key, ok := data["key"]
	if !ok {
		return "", fmt.Errorf("missing 'key' in key generator data")
	}
	return prefix + ":" + key, nil
}

// defaultKeyVersionGenerator generates a versioned cache key.
// Template: {prefix}:{namespace}:v{version}-{key}.gob
func defaultKeyVersionGenerator(data map[string]string) (string, error) {
	prefix, ok := data["prefix"]
	if !ok {
		return "", fmt.Errorf("missing 'prefix' in key version generator data")
	}
	namespace, ok := data["namespace"]
	if !ok {
		return "", fmt.Errorf("missing 'namespace' in key version generator data")
	}
	version, ok := data["version"]
	if !ok {
		return "", fmt.Errorf("missing 'version' in key version generator data")
	}
	key, ok := data["key"]
	if !ok {
		return "", fmt.Errorf("missing 'key' in key version generator data")
	}
	return fmt.Sprintf("%s:%s:v%s-%s.gob", prefix, namespace, version, key), nil
}

// defaultVersionGenerator generates a version key.
// Template: {prefix}:{namespace}:version
func defaultVersionGenerator(data map[string]string) (string, error) {
	prefix, ok := data["prefix"]
	if !ok {
		return "", fmt.Errorf("missing 'prefix' in version generator data")
	}
	namespace, ok := data["namespace"]
	if !ok {
		return "", fmt.Errorf("missing 'namespace' in version generator data")
	}
	return fmt.Sprintf("%s:%s:version", prefix, namespace), nil
}

// defaultLockGenerator generates a lock key.
// Template: {prefix}:{namespace}:lock
func defaultLockGenerator(data map[string]string) (string, error) {
	prefix, ok := data["prefix"]
	if !ok {
		return "", fmt.Errorf("missing 'prefix' in lock generator data")
	}
	namespace, ok := data["namespace"]
	if !ok {
		return "", fmt.Errorf("missing 'namespace' in lock generator data")
	}
	return fmt.Sprintf("%s:%s:lock", prefix, namespace), nil
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
