package errs

import (
	"testing"
)

func TestErrInvalidConfig(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		expectedMsg     string
		expectedField   string
		expectedMessage string
	}{
		{
			name:            "simple message",
			err:             NewInvalidConfigError("redisClient", "cannot be nil"),
			expectedMsg:     "invalid config redisClient : cannot be nil",
			expectedField:   "redisClient",
			expectedMessage: "cannot be nil",
		},
		{
			name:            "empty field",
			err:             NewInvalidConfigError("", "missing configuration"),
			expectedMsg:     "invalid config  : missing configuration",
			expectedField:   "",
			expectedMessage: "missing configuration",
		},
		{
			name:            "empty message",
			err:             NewInvalidConfigError("timeout", ""),
			expectedMsg:     "invalid config timeout : ",
			expectedField:   "timeout",
			expectedMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, ok := tt.err.(*invalidConfigError)
			if !ok {
				t.Errorf("Expected InvalidConfigError, got %T", tt.err)
			}
			if err.Error() != tt.expectedMsg {
				t.Errorf("Error() expected %q, got %q", tt.expectedMsg, err.Error())
			}
			if err.Field() != tt.expectedField {
				t.Errorf("Field() expected %q, got %q", tt.expectedField, err.Field())
			}
			if err.Message() != tt.expectedMessage {
				t.Errorf("Message() expected %q, got %q", tt.expectedMessage, err.Message())
			}
		})
	}
}
