package adapter

import "context"

// Tracer interface for optional observability.
type Tracer interface {
	StartSpan(ctx context.Context, name string) (context.Context, Span)                    // StartSpan starts a new span with the given name. Returns a context with the span and the span itself.
	NewSpanFromSpan(ctx context.Context, name string, parent Span) (context.Context, Span) // NewSpanFromSpan creates a new span from an existing span (child span). Returns a context with the new span and the span itself.
}

// Span represents a tracing span.
type Span interface {
	End()                                         // End finishes the span.
	AddEvent(name string, attrs ...SpanAttribute) // AddEvent adds an event to the span.
	SetAttributes(attrs ...SpanAttribute)         // SetAttributes sets attributes on the span.
	SpanContext() SpanContext                     // SpanContext returns the span context for propagation.
}

// SpanContext represents span context for propagation.
type SpanContext interface {
	TraceID() string // TraceID returns the trace ID.
	SpanID() string  // SpanID returns the span ID.
}

// SpanAttribute represents a span attribute.
type SpanAttribute interface{}

// Logger interface for optional observability.
type Logger interface {
	Info(msg string, fields map[string]interface{})  // Info logs an informational message with optional fields.
	Error(msg string, fields map[string]interface{}) // Error logs an error message with optional fields.
	Debug(msg string, fields map[string]interface{}) // Debug logs a debug message with optional fields.
	WithSpanContext(spanContext SpanContext) Logger  // WithSpanContext returns a logger with span context for correlation.
}

// Semaphore interface for concurrency control.
type Semaphore interface {
	Acquire() // Acquire acquires a semaphore permit, blocking if necessary.
	Release() // Release releases a semaphore permit.
}
