package cache

import (
	"context"
)

// NoOpTracer is a no-op tracer implementation that does nothing.
// Used as a default when no tracer is provided.
type NoOpTracer struct{}

// StartSpan starts a no-op span.
func (n *NoOpTracer) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	return ctx, &NoOpSpan{}
}

// NewSpanFromSpan creates a no-op child span.
func (n *NoOpTracer) NewSpanFromSpan(ctx context.Context, name string, parent Span) (context.Context, Span) {
	return ctx, &NoOpSpan{}
}

// NoOpSpan is a no-op span implementation.
type NoOpSpan struct{}

// End does nothing.
func (n *NoOpSpan) End() {}

// AddEvent does nothing.
func (n *NoOpSpan) AddEvent(name string, attrs ...SpanAttribute) {}

// SetAttributes does nothing.
func (n *NoOpSpan) SetAttributes(attrs ...SpanAttribute) {}

// SpanContext returns a no-op span context.
func (n *NoOpSpan) SpanContext() SpanContext {
	return &NoOpSpanContext{}
}

// NoOpSpanContext is a no-op span context implementation.
type NoOpSpanContext struct{}

// TraceID returns an empty string.
func (n *NoOpSpanContext) TraceID() string {
	return ""
}

// SpanID returns an empty string.
func (n *NoOpSpanContext) SpanID() string {
	return ""
}

// NoOpLogger is a no-op logger implementation that does nothing.
// Used as a default when no logger is provided.
type NoOpLogger struct{}

// Info does nothing.
func (n *NoOpLogger) Info(msg string, fields map[string]interface{}) {}

// Error does nothing.
func (n *NoOpLogger) Error(msg string, fields map[string]interface{}) {}

// Debug does nothing.
func (n *NoOpLogger) Debug(msg string, fields map[string]interface{}) {}

// WithSpanContext returns the same no-op logger.
func (n *NoOpLogger) WithSpanContext(spanContext SpanContext) Logger {
	return n
}

// DefaultSemaphore is a channel-based semaphore implementation.
// This is the default semaphore used when none is provided.
type DefaultSemaphore struct {
	sem chan struct{}
}

// NewDefaultSemaphore creates a new default semaphore with the given size.
func NewDefaultSemaphore(size int) *DefaultSemaphore {
	if size <= 0 {
		size = 1
	}
	return &DefaultSemaphore{
		sem: make(chan struct{}, size),
	}
}

// Acquire acquires a semaphore permit, blocking if necessary.
func (d *DefaultSemaphore) Acquire() {
	d.sem <- struct{}{}
}

// Release releases a semaphore permit.
func (d *DefaultSemaphore) Release() {
	<-d.sem
}
