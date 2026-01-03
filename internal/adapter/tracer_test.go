package adapter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdapter_NoOpTracer_StartSpan(t *testing.T) {
	tracer := &NoOpTracer{}
	ctx := context.Background()

	tests := []struct {
		name      string
		spanName  string
		checkFunc func(t *testing.T, newCtx context.Context, span Span)
	}{
		{
			name:     "start span with valid name",
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
		{
			name:     "start span with empty name",
			spanName: "",
			checkFunc: func(t *testing.T, newCtx context.Context, span Span) {
				assert.NotNil(t, newCtx)
				assert.NotNil(t, span)
				assert.IsType(t, &NoOpSpan{}, span)
			},
		},
		{
			name:     "start span with nil context",
			spanName: "test-span",
			checkFunc: func(t *testing.T, newCtx context.Context, span Span) {
				assert.NotNil(t, span)
				assert.IsType(t, &NoOpSpan{}, span)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCtx := ctx
			if tt.name == "start span with nil context" {
				testCtx = nil
			}
			newCtx, span := tracer.StartSpan(testCtx, tt.spanName)
			if tt.checkFunc != nil {
				tt.checkFunc(t, newCtx, span)
			}
		})
	}
}

func TestAdapter_NoOpTracer_NewSpanFromSpan(t *testing.T) {
	tracer := &NoOpTracer{}
	ctx := context.Background()
	parentSpan := &NoOpSpan{}

	tests := []struct {
		name      string
		spanName  string
		parent    Span
		checkFunc func(t *testing.T, newCtx context.Context, span Span)
	}{
		{
			name:     "new span from parent",
			spanName: "child-span",
			parent:   parentSpan,
			checkFunc: func(t *testing.T, newCtx context.Context, span Span) {
				assert.NotNil(t, newCtx)
				assert.NotNil(t, span)
				assert.IsType(t, &NoOpSpan{}, span)
			},
		},
		{
			name:     "new span from parent with empty name",
			spanName: "",
			parent:   parentSpan,
			checkFunc: func(t *testing.T, newCtx context.Context, span Span) {
				assert.NotNil(t, newCtx)
				assert.NotNil(t, span)
				assert.IsType(t, &NoOpSpan{}, span)
			},
		},
		{
			name:     "new span with nil parent",
			spanName: "child-span",
			parent:   nil,
			checkFunc: func(t *testing.T, newCtx context.Context, span Span) {
				assert.NotNil(t, newCtx)
				assert.NotNil(t, span)
				assert.IsType(t, &NoOpSpan{}, span)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newCtx, span := tracer.NewSpanFromSpan(ctx, tt.spanName, tt.parent)
			if tt.checkFunc != nil {
				tt.checkFunc(t, newCtx, span)
			}
		})
	}
}

func TestAdapter_NoOpTracer_NewStringAttribute(t *testing.T) {
	tracer := &NoOpTracer{}

	tests := []struct {
		name      string
		key       string
		value     string
		checkFunc func(t *testing.T, attr SpanAttribute)
	}{
		{
			name:  "create string attribute",
			key:   "key",
			value: "value",
			checkFunc: func(t *testing.T, attr SpanAttribute) {
				assert.NotNil(t, attr)
				assert.IsType(t, &NoOpSpanAttribute{}, attr)
			},
		},
		{
			name:  "create string attribute with empty key",
			key:   "",
			value: "value",
			checkFunc: func(t *testing.T, attr SpanAttribute) {
				assert.NotNil(t, attr)
				assert.IsType(t, &NoOpSpanAttribute{}, attr)
			},
		},
		{
			name:  "create string attribute with empty value",
			key:   "key",
			value: "",
			checkFunc: func(t *testing.T, attr SpanAttribute) {
				assert.NotNil(t, attr)
				assert.IsType(t, &NoOpSpanAttribute{}, attr)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := tracer.NewStringAttribute(tt.key, tt.value)
			if tt.checkFunc != nil {
				tt.checkFunc(t, attr)
			}
		})
	}
}

func TestAdapter_NoOpSpan_End(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, span *NoOpSpan)
	}{
		{
			name: "end span",
			checkFunc: func(t *testing.T, span *NoOpSpan) {
				// Should not panic
				span.End()
				// Can be called multiple times
				span.End()
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

func TestAdapter_NoOpSpan_AddEvent(t *testing.T) {
	tests := []struct {
		name      string
		event     string
		attrs     []SpanAttribute
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
		{
			name:  "add event with empty attrs",
			event: "test-event",
			attrs: []SpanAttribute{},
			checkFunc: func(t *testing.T, span *NoOpSpan) {
				// Should not panic
				span.AddEvent("test-event")
			},
		},
		{
			name:  "add event with multiple attrs",
			event: "test-event",
			attrs: []SpanAttribute{
				&NoOpSpanAttribute{},
				&NoOpSpanAttribute{},
			},
			checkFunc: func(t *testing.T, span *NoOpSpan) {
				// Should not panic
				span.AddEvent("test-event", &NoOpSpanAttribute{}, &NoOpSpanAttribute{})
			},
		},
		{
			name:  "add event with empty name",
			event: "",
			checkFunc: func(t *testing.T, span *NoOpSpan) {
				// Should not panic
				span.AddEvent("")
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

func TestAdapter_NoOpSpan_SetAttributes(t *testing.T) {
	tests := []struct {
		name      string
		attrs     []SpanAttribute
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
		{
			name:  "set attributes with empty slice",
			attrs: []SpanAttribute{},
			checkFunc: func(t *testing.T, span *NoOpSpan) {
				// Should not panic
				span.SetAttributes()
			},
		},
		{
			name: "set attributes with multiple attrs",
			attrs: []SpanAttribute{
				&NoOpSpanAttribute{},
				&NoOpSpanAttribute{},
			},
			checkFunc: func(t *testing.T, span *NoOpSpan) {
				// Should not panic
				span.SetAttributes(&NoOpSpanAttribute{}, &NoOpSpanAttribute{})
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

func TestAdapter_NoOpSpan_SpanContext(t *testing.T) {
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

func TestAdapter_NoOpSpan_MultipleCalls(t *testing.T) {
	span := &NoOpSpan{}

	// Test that multiple calls don't interfere with each other
	span.AddEvent("event1")
	span.AddEvent("event2")
	span.SetAttributes(&NoOpSpanAttribute{})
	span.SetAttributes(&NoOpSpanAttribute{}, &NoOpSpanAttribute{})
	spanCtx1 := span.SpanContext()
	spanCtx2 := span.SpanContext()
	assert.Equal(t, spanCtx1, spanCtx2)
	span.End()
	span.End()
}

func TestAdapter_NoOpSpanContext_TraceID(t *testing.T) {
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

func TestAdapter_NoOpSpanContext_SpanID(t *testing.T) {
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

func TestNoOpSpanContext_MultipleCalls(t *testing.T) {
	spanCtx := &NoOpSpanContext{}

	// Test that multiple calls return consistent values
	traceID1 := spanCtx.TraceID()
	traceID2 := spanCtx.TraceID()
	spanID1 := spanCtx.SpanID()
	spanID2 := spanCtx.SpanID()

	assert.Equal(t, "", traceID1)
	assert.Equal(t, "", traceID2)
	assert.Equal(t, "", spanID1)
	assert.Equal(t, "", spanID2)
	assert.Equal(t, traceID1, traceID2)
	assert.Equal(t, spanID1, spanID2)
}
