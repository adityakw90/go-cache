package adapter

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
