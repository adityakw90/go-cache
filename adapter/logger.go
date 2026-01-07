package adapter

import (
	"context"

	internaladapter "github.com/adityakw90/go-cache/internal/adapter"
)

// Logger interface for optional observability.
type Logger interface {
	Info(msg string, fields map[string]interface{})  // Info logs an informational message with optional fields.
	Error(msg string, fields map[string]interface{}) // Error logs an error message with optional fields.
	Debug(msg string, fields map[string]interface{}) // Debug logs a debug message with optional fields.
}

// GetLogger is a hook type that returns a Logger with trace/span correlation.
// Users can register their own implementation to customize logger behavior.
type GetLogger func(ctx context.Context) Logger

var noOpLogger Logger = &internaladapter.NoOpLogger{}

var GetNoOpLogger GetLogger = func(ctx context.Context) Logger {
	return noOpLogger
}

// NewNoOpLogger creates a new no-op logger.
func NewNoOpLogger() Logger {
	return noOpLogger
}
