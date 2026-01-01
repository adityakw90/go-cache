package adapter

import "context"

// No Operation Tracer implementation.
type NoOpTracer struct{}
type NoOpSpan struct{}
type NoOpSpanContext struct{}
type NoOpSpanAttribute struct{}

var noOpSpan = &NoOpSpan{}
var noOpSpanContext = &NoOpSpanContext{}
var noOpSpanAttribute = &NoOpSpanAttribute{}

// StartSpan starts a no-op span.
func (n *NoOpTracer) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	return ctx, noOpSpan
}

// NewSpanFromSpan creates a no-op child span.
func (n *NoOpTracer) NewSpanFromSpan(ctx context.Context, name string, parent Span) (context.Context, Span) {
	return ctx, noOpSpan
}

// NewStringAttribute creates a no-op string attribute.
func (n *NoOpTracer) NewStringAttribute(key string, value string) SpanAttribute {
	return noOpSpanAttribute
}

// End does nothing.
func (n *NoOpSpan) End() {}

// AddEvent does nothing.
func (n *NoOpSpan) AddEvent(name string, attrs ...SpanAttribute) {}

// SetAttributes does nothing.
func (n *NoOpSpan) SetAttributes(attrs ...SpanAttribute) {}

// SpanContext returns a no-op span context.
func (n *NoOpSpan) SpanContext() SpanContext {
	return noOpSpanContext
}

// NoOpSpanContext is a no-op span context implementation.

// TraceID returns an empty string.
func (n *NoOpSpanContext) TraceID() string {
	return ""
}

// SpanID returns an empty string.
func (n *NoOpSpanContext) SpanID() string {
	return ""
}
