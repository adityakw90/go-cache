package adapter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdapter_NewNoOpTracer(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, tracer Tracer)
	}{
		{
			name: "create new no-op tracer",
			checkFunc: func(t *testing.T, tracer Tracer) {
				assert.NotNil(t, tracer)
				assert.IsType(t, &NoOpTracer{}, tracer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracer := NewNoOpTracer()
			if tt.checkFunc != nil {
				tt.checkFunc(t, tracer)
			}
		})
	}
}

func TestAdapter_NewNoOpTracer_MultipleInstances(t *testing.T) {
	tracer1 := NewNoOpTracer()
	tracer2 := NewNoOpTracer()

	assert.NotNil(t, tracer1)
	assert.NotNil(t, tracer2)
	assert.IsType(t, &NoOpTracer{}, tracer1)
	assert.IsType(t, &NoOpTracer{}, tracer2)

	// Each instance should work independently
	ctx1, span1 := tracer1.StartSpan(context.Background(), "span1")
	ctx2, span2 := tracer2.StartSpan(context.Background(), "span2")

	assert.NotNil(t, span1)
	assert.NotNil(t, span2)
	assert.NotNil(t, ctx1)
	assert.NotNil(t, ctx2)
}

func TestAdapter_NewNoOpLogger(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, logger Logger)
	}{
		{
			name: "create new no-op logger",
			checkFunc: func(t *testing.T, logger Logger) {
				assert.NotNil(t, logger)
				assert.IsType(t, &NoOpLogger{}, logger)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewNoOpLogger()
			if tt.checkFunc != nil {
				tt.checkFunc(t, logger)
			}
		})
	}
}

func TestAdapter_NewNoOpLogger_MultipleInstances(t *testing.T) {
	logger1 := NewNoOpLogger()
	logger2 := NewNoOpLogger()

	assert.NotNil(t, logger1)
	assert.NotNil(t, logger2)
	assert.IsType(t, &NoOpLogger{}, logger1)
	assert.IsType(t, &NoOpLogger{}, logger2)

	// Each instance should work independently
	logger1.Info("msg1", nil)
	logger2.Error("msg2", nil)
	logger1.Debug("msg3", nil)
	logger2.Info("msg4", nil)
}

func TestAdapter_Registry_Integration(t *testing.T) {
	// Test that all factory functions work together
	tracer := NewNoOpTracer()
	logger := NewNoOpLogger()
	semaphore := NewSemaphore(10)

	assert.NotNil(t, tracer)
	assert.NotNil(t, logger)
	assert.NotNil(t, semaphore)

	// Use them together
	ctx, span := tracer.StartSpan(context.Background(), "test-span")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)

	logger.Info("test message", map[string]interface{}{"key": "value"})
	logger.WithSpanContext(span.SpanContext()).Debug("debug message", nil)

	semaphore.Acquire()
	semaphore.Release()
}
