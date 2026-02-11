package adapter

import (
	"testing"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/stretchr/testify/assert"
)

func TestAdapter_NewNoOpLogger(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, logger Logger)
	}{
		{
			name: "create new no-op logger",
			checkFunc: func(t *testing.T, logger Logger) {
				assert.NotNil(t, logger)
				assert.IsType(t, &adapter.NoOpLogger{}, logger)
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
	assert.IsType(t, &adapter.NoOpLogger{}, logger1)
	assert.IsType(t, &adapter.NoOpLogger{}, logger2)

	// Each instance should work independently
	logger1.Info("msg1", nil)
	logger2.Error("msg2", nil)
	logger1.Debug("msg3", nil)
	logger2.Info("msg4", nil)
}
