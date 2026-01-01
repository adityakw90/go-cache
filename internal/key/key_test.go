package key

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKey_NewCustomKeyFunction_Validation(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (CustomKeyFunction, error)
		wantErr     bool
		errContains string
	}{
		{
			name: "empty name returns error",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"",
					func(args ...interface{}) string {
						return "test"
					},
					[]string{"param"},
				)
			},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "nil callable returns error",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"test",
					nil,
					[]string{"param"},
				)
			},
			wantErr:     true,
			errContains: "callable is required",
		},
		{
			name: "empty params returns error",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"test",
					func(args ...interface{}) string {
						return "test"
					},
					[]string{},
				)
			},
			wantErr:     true,
			errContains: "params is required",
		},
		{
			name: "valid parameters succeed",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"test",
					func(args ...interface{}) string {
						return "test"
					},
					[]string{"param"},
				)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ckf, err := tt.setupFunc()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, ckf)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, ckf)
			}
		})
	}
}
