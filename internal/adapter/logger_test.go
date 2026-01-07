package adapter

import (
	"testing"
)

func TestAdapter_NoOpLogger_Info(t *testing.T) {
	logger := &NoOpLogger{}

	tests := []struct {
		name      string
		message   string
		fields    map[string]interface{}
		checkFunc func(t *testing.T, logger *NoOpLogger)
	}{
		{
			name:    "info with nil fields",
			message: "test message",
			fields:  nil,
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Info("test message", nil)
			},
		},
		{
			name:    "info with empty fields",
			message: "test message",
			fields:  map[string]interface{}{},
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Info("test message", map[string]interface{}{})
			},
		},
		{
			name:    "info with fields",
			message: "test message",
			fields:  map[string]interface{}{"key": "value"},
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Info("test message", map[string]interface{}{"key": "value"})
			},
		},
		{
			name:    "info with multiple fields",
			message: "test message",
			fields: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
				"key3": true,
			},
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Info("test message", map[string]interface{}{
					"key1": "value1",
					"key2": 123,
					"key3": true,
				})
			},
		},
		{
			name:    "info with empty message",
			message: "",
			fields:  nil,
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Info("", nil)
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

func TestAdapter_NoOpLogger_Error(t *testing.T) {
	logger := &NoOpLogger{}

	tests := []struct {
		name      string
		message   string
		fields    map[string]interface{}
		checkFunc func(t *testing.T, logger *NoOpLogger)
	}{
		{
			name:    "error with nil fields",
			message: "test error",
			fields:  nil,
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Error("test error", nil)
			},
		},
		{
			name:    "error with empty fields",
			message: "test error",
			fields:  map[string]interface{}{},
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Error("test error", map[string]interface{}{})
			},
		},
		{
			name:    "error with fields",
			message: "test error",
			fields:  map[string]interface{}{"error": "something went wrong"},
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Error("test error", map[string]interface{}{"error": "something went wrong"})
			},
		},
		{
			name:    "error with multiple fields",
			message: "test error",
			fields: map[string]interface{}{
				"error":   "something went wrong",
				"code":    500,
				"retry":   true,
				"details": map[string]string{"key": "value"},
			},
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Error("test error", map[string]interface{}{
					"error":   "something went wrong",
					"code":    500,
					"retry":   true,
					"details": map[string]string{"key": "value"},
				})
			},
		},
		{
			name:    "error with empty message",
			message: "",
			fields:  nil,
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Error("", nil)
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

func TestAdapter_NoOpLogger_Debug(t *testing.T) {
	logger := &NoOpLogger{}

	tests := []struct {
		name      string
		message   string
		fields    map[string]interface{}
		checkFunc func(t *testing.T, logger *NoOpLogger)
	}{
		{
			name:    "debug with nil fields",
			message: "test debug",
			fields:  nil,
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Debug("test debug", nil)
			},
		},
		{
			name:    "debug with empty fields",
			message: "test debug",
			fields:  map[string]interface{}{},
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Debug("test debug", map[string]interface{}{})
			},
		},
		{
			name:    "debug with fields",
			message: "test debug",
			fields:  map[string]interface{}{"debug": "info"},
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Debug("test debug", map[string]interface{}{"debug": "info"})
			},
		},
		{
			name:    "debug with multiple fields",
			message: "test debug",
			fields: map[string]interface{}{
				"debug":   "info",
				"level":   "verbose",
				"enabled": true,
			},
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Debug("test debug", map[string]interface{}{
					"debug":   "info",
					"level":   "verbose",
					"enabled": true,
				})
			},
		},
		{
			name:    "debug with empty message",
			message: "",
			fields:  nil,
			checkFunc: func(t *testing.T, logger *NoOpLogger) {
				// Should not panic
				logger.Debug("", nil)
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

func TestAdapter_NoOpLogger_MultipleCalls(t *testing.T) {
	logger := &NoOpLogger{}

	// Test that multiple calls don't interfere with each other
	logger.Info("msg1", map[string]interface{}{"key1": "value1"})
	logger.Error("msg2", map[string]interface{}{"key2": "value2"})
	logger.Debug("msg3", map[string]interface{}{"key3": "value3"})
}
