package cache

import (
	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/key"
)

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
