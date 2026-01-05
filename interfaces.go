package cache

import (
	"context"
	"time"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string, resultType interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	CleanCache(ctx context.Context, key string, params map[string]interface{}, session redis.Pipeliner, execute bool, listLockedKey *[]string) error
	Cached(keyName string, ttl interface{}, versioning bool, prefix string) func(fn func(ctx context.Context, args ...interface{}) (interface{}, error), customKeyFunc key.CustomKeyFunction) func(resultType interface{}, ctx context.Context, args ...interface{}) (interface{}, error)
}

// re-export
type Tracer = adapter.Tracer
type Span = adapter.Span
type SpanContext = adapter.SpanContext
type SpanAttribute = adapter.SpanAttribute
type Logger = adapter.Logger
type Semaphore = adapter.Semaphore

// re export custom key function
type CustomKeyFunction = key.CustomKeyFunction

// NewCustomKeyFunction creates a new custom key function.
func NewCustomKeyFunction(
	name string,
	callable func(args ...interface{}) string,
	params []string,
) (CustomKeyFunction, error) {
	return key.NewCustomKeyFunction(name, callable, params)
}
