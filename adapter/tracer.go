package adapter

import (
	"context"

	"github.com/adityakw90/go-cache/internal/adapter"
)

// Span represents a tracing span.
type Span interface {
	End() // End finishes the span.
}

// hooks for tracing
type StartSpan func(ctx context.Context, name string) (context.Context, Span)
type StartChildSpan func(ctx context.Context, name string, parent Span) (context.Context, Span)

var noOpSpan Span = &adapter.NoOpSpan{}

// NoOpStartSpan is a no-op implementation of StartSpan.
var NoOpStartSpan StartSpan = func(ctx context.Context, name string) (context.Context, Span) {
	return ctx, noOpSpan
}

// NoOpStartChildSpan is a no-op implementation of StartChildSpan.
var NoOpStartChildSpan StartChildSpan = func(ctx context.Context, name string, parent Span) (context.Context, Span) {
	return ctx, noOpSpan
}
