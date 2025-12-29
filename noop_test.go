package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoOpTracer_StartSpan(t *testing.T) {
	tracer := &NoOpTracer{}
	ctx := context.Background()

	tests := []struct {
		name      string
		spanName  string
		checkFunc func(t *testing.T, newCtx context.Context, span Span)
	}{
		{
			name:     "start span",
			spanName: "test-span",
			checkFunc: func(t *testing.T, newCtx context.Context, span Span) {
				assert.NotNil(t, newCtx)
				assert.NotNil(t, span)
				assert.IsType(t, &NoOpSpan{}, span)

				// Should not panic
				span.End()
				span.AddEvent("test-event")
				span.SetAttributes()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newCtx, span := tracer.StartSpan(ctx, tt.spanName)
			if tt.checkFunc != nil {
				tt.checkFunc(t, newCtx, span)
			}
		})
	}
}

func TestNoOpTracer_NewSpanFromSpan(t *testing.T) {
	tracer := &NoOpTracer{}
	ctx := context.Background()
	parentSpan := &NoOpSpan{}

	tests := []struct {
		name      string
		spanName  string
		checkFunc func(t *testing.T, newCtx context.Context, span Span)
	}{
		{
			name:     "new span from parent",
			spanName: "child-span",
			checkFunc: func(t *testing.T, newCtx context.Context, span Span) {
				assert.NotNil(t, newCtx)
				assert.NotNil(t, span)
				assert.IsType(t, &NoOpSpan{}, span)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newCtx, span := tracer.NewSpanFromSpan(ctx, tt.spanName, parentSpan)
			if tt.checkFunc != nil {
				tt.checkFunc(t, newCtx, span)
			}
		})
	}
}

func TestNoOpSpan_End(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, span *NoOpSpan)
	}{
		{
			name: "end span",
			checkFunc: func(t *testing.T, span *NoOpSpan) {
				// Should not panic
				span.End()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := &NoOpSpan{}
			if tt.checkFunc != nil {
				tt.checkFunc(t, span)
			}
		})
	}
}

func TestNoOpSpan_AddEvent(t *testing.T) {
	tests := []struct {
		name      string
		event     string
		attrs     map[string]interface{}
		checkFunc func(t *testing.T, span *NoOpSpan)
	}{
		{
			name:  "add event without attrs",
			event: "test-event",
			checkFunc: func(t *testing.T, span *NoOpSpan) {
				// Should not panic
				span.AddEvent("test-event")
			},
		},
		{
			name:  "add event with nil attrs",
			event: "test-event",
			attrs: nil,
			checkFunc: func(t *testing.T, span *NoOpSpan) {
				// Should not panic
				span.AddEvent("test-event", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := &NoOpSpan{}
			if tt.checkFunc != nil {
				tt.checkFunc(t, span)
			}
		})
	}
}

func TestNoOpSpan_SetAttributes(t *testing.T) {
	tests := []struct {
		name      string
		attrs     map[string]interface{}
		checkFunc func(t *testing.T, span *NoOpSpan)
	}{
		{
			name: "set attributes without attrs",
			checkFunc: func(t *testing.T, span *NoOpSpan) {
				// Should not panic
				span.SetAttributes()
			},
		},
		{
			name:  "set attributes with nil",
			attrs: nil,
			checkFunc: func(t *testing.T, span *NoOpSpan) {
				// Should not panic
				span.SetAttributes(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := &NoOpSpan{}
			if tt.checkFunc != nil {
				tt.checkFunc(t, span)
			}
		})
	}
}

func TestNoOpSpan_SpanContext(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, spanCtx SpanContext)
	}{
		{
			name: "get span context",
			checkFunc: func(t *testing.T, spanCtx SpanContext) {
				assert.NotNil(t, spanCtx)
				assert.IsType(t, &NoOpSpanContext{}, spanCtx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := &NoOpSpan{}
			spanCtx := span.SpanContext()
			if tt.checkFunc != nil {
				tt.checkFunc(t, spanCtx)
			}
		})
	}
}

func TestNoOpSpanContext_TraceID(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, traceID string)
	}{
		{
			name: "get trace ID",
			checkFunc: func(t *testing.T, traceID string) {
				assert.Equal(t, "", traceID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spanCtx := &NoOpSpanContext{}
			traceID := spanCtx.TraceID()
			if tt.checkFunc != nil {
				tt.checkFunc(t, traceID)
			}
		})
	}
}

func TestNoOpSpanContext_SpanID(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, spanID string)
	}{
		{
			name: "get span ID",
			checkFunc: func(t *testing.T, spanID string) {
				assert.Equal(t, "", spanID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spanCtx := &NoOpSpanContext{}
			spanID := spanCtx.SpanID()
			if tt.checkFunc != nil {
				tt.checkFunc(t, spanID)
			}
		})
	}
}

func TestNoOpLogger_Info(t *testing.T) {
	logger := &NoOpLogger{}

	tests := []struct {
		name      string
		message   string
		attrs     map[string]interface{}
		checkFunc func(t *testing.T, logger *NoOpLogger)
	}{
		{
			name:    "info with nil attrs",
			message: "test message",
			attrs:   nil,
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Info("test message", nil)
			},
		},
		{
			name:    "info with attrs",
			message: "test message",
			attrs:   map[string]interface{}{"key": "value"},
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Info("test message", map[string]interface{}{"key": "value"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.checkFunc != nil {
				tt.checkFunc(t, logger)
			}
		})
	}
}

func TestNoOpLogger_Error(t *testing.T) {
	logger := &NoOpLogger{}

	tests := []struct {
		name      string
		message   string
		attrs     map[string]interface{}
		checkFunc func(t *testing.T, logger *NoOpLogger)
	}{
		{
			name:    "error with nil attrs",
			message: "test error",
			attrs:   nil,
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Error("test error", nil)
			},
		},
		{
			name:    "error with attrs",
			message: "test error",
			attrs:   map[string]interface{}{"error": "something went wrong"},
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Error("test error", map[string]interface{}{"error": "something went wrong"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.checkFunc != nil {
				tt.checkFunc(t, logger)
			}
		})
	}
}

func TestNoOpLogger_Debug(t *testing.T) {
	logger := &NoOpLogger{}

	tests := []struct {
		name      string
		message   string
		attrs     map[string]interface{}
		checkFunc func(t *testing.T, logger *NoOpLogger)
	}{
		{
			name:    "debug with nil attrs",
			message: "test debug",
			attrs:   nil,
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Debug("test debug", nil)
			},
		},
		{
			name:    "debug with attrs",
			message: "test debug",
			attrs:   map[string]interface{}{"debug": "info"},
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Debug("test debug", map[string]interface{}{"debug": "info"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.checkFunc != nil {
				tt.checkFunc(t, logger)
			}
		})
	}
}

func TestNoOpLogger_WithSpanContext(t *testing.T) {
	logger := &NoOpLogger{}
	spanCtx := &NoOpSpanContext{}

	tests := []struct {
		name      string
		checkFunc func(t *testing.T, newLogger Logger)
	}{
		{
			name: "with span context",
			checkFunc: func(t *testing.T, newLogger Logger) {
				assert.NotNil(t, newLogger)
				assert.Equal(t, logger, newLogger) // Should return itself
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newLogger := logger.WithSpanContext(spanCtx)
			if tt.checkFunc != nil {
				tt.checkFunc(t, newLogger)
			}
		})
	}
}

func TestDefaultSemaphore_NewDefaultSemaphore(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		checkFunc func(t *testing.T, sem *DefaultSemaphore)
	}{
		{
			name: "valid size",
			size: 5,
			checkFunc: func(t *testing.T, sem *DefaultSemaphore) {
				assert.NotNil(t, sem)
				assert.NotNil(t, sem.sem)
			},
		},
		{
			name: "zero size defaults to 1",
			size: 0,
			checkFunc: func(t *testing.T, sem *DefaultSemaphore) {
				assert.NotNil(t, sem)
				// Should default to size 1
				assert.Equal(t, 1, cap(sem.sem))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sem := NewDefaultSemaphore(tt.size)
			if tt.checkFunc != nil {
				tt.checkFunc(t, sem)
			}
		})
	}
}

func TestDefaultSemaphore_AcquireRelease(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		checkFunc func(t *testing.T, sem *DefaultSemaphore)
	}{
		{
			name: "acquire and release",
			size: 2,
			checkFunc: func(t *testing.T, sem *DefaultSemaphore) {
				// Should not block when capacity available
				sem.Acquire()
				sem.Acquire()

				// Release should work
				sem.Release()
				sem.Release()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sem := NewDefaultSemaphore(tt.size)
			if tt.checkFunc != nil {
				tt.checkFunc(t, sem)
			}
		})
	}
}

func TestDefaultSemaphore_Concurrent(t *testing.T) {
	sem := NewDefaultSemaphore(2)
	done := make(chan bool, 3)

	// Try to acquire 3 permits with capacity 2
	go func() {
		sem.Acquire()
		done <- true
	}()
	go func() {
		sem.Acquire()
		done <- true
	}()
	go func() {
		sem.Acquire()
		done <- true
	}()

	// First two should complete quickly
	<-done
	<-done

	// Release one to allow third
	sem.Release()

	// Third should complete
	<-done
}
