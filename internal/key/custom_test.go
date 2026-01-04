package key

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKey_CustomKeyFunction_Name(t *testing.T) {
	tests := []struct {
		name         string
		functionName string
		wantName     string
	}{
		{
			name:         "returns correct name",
			functionName: "getUser",
			wantName:     "getUser",
		},
		{
			name:         "returns name with special characters",
			functionName: "getUserById",
			wantName:     "getUserById",
		},
		{
			name:         "returns name with underscores",
			functionName: "get_user_by_id",
			wantName:     "get_user_by_id",
		},
		{
			name:         "returns name with numbers",
			functionName: "getUser123",
			wantName:     "getUser123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ckf, err := NewCustomKeyFunction(
				tt.functionName,
				func(args ...interface{}) string {
					return "test"
				},
				[]string{"param"},
			)
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, ckf.Name())
		})
	}
}

func TestKey_CustomKeyFunction_Call(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (CustomKeyFunction, error)
		params      map[string]interface{}
		wantResult  string
		wantErr     bool
		errContains string
	}{
		{
			name: "single string parameter",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getUser",
					func(args ...interface{}) string {
						return "user:" + args[0].(string)
					},
					[]string{"uid"},
				)
			},
			params: map[string]interface{}{
				"uid": "123",
			},
			wantResult: "user:123",
			wantErr:    false,
		},
		{
			name: "multiple parameters",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getUserByType",
					func(args ...interface{}) string {
						return "user:" + args[0].(string) + ":" + args[1].(string)
					},
					[]string{"uid", "type"},
				)
			},
			params: map[string]interface{}{
				"uid":  "123",
				"type": "admin",
			},
			wantResult: "user:123:admin",
			wantErr:    false,
		},
		{
			name: "mixed parameter types",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getUserById",
					func(args ...interface{}) string {
						return fmt.Sprintf("user:%d:%s", args[0].(int), args[1].(string))
					},
					[]string{"id", "name"},
				)
			},
			params: map[string]interface{}{
				"id":   65,
				"name": "test",
			},
			wantResult: "user:65:test",
			wantErr:    false,
		},
		{
			name: "missing required parameter",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getUser",
					func(args ...interface{}) string {
						return "user:" + args[0].(string)
					},
					[]string{"uid"},
				)
			},
			params: map[string]interface{}{
				"other": "value",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "required param uid missing",
		},
		{
			name: "missing one of multiple required parameters",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getUserByType",
					func(args ...interface{}) string {
						return "user:" + args[0].(string) + ":" + args[1].(string)
					},
					[]string{"uid", "type"},
				)
			},
			params: map[string]interface{}{
				"uid": "123",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "required param type missing",
		},
		{
			name: "extra parameters ignored",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getUser",
					func(args ...interface{}) string {
						return "user:" + args[0].(string)
					},
					[]string{"uid"},
				)
			},
			params: map[string]interface{}{
				"uid":   "123",
				"extra": "ignored",
			},
			wantResult: "user:123",
			wantErr:    false,
		},
		{
			name: "nil params map with required params",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getUser",
					func(args ...interface{}) string {
						return "user:" + args[0].(string)
					},
					[]string{"uid"},
				)
			},
			params:      nil,
			wantResult:  "",
			wantErr:     true,
			errContains: "required param uid missing",
		},
		{
			name: "empty params map with required params",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getUser",
					func(args ...interface{}) string {
						return "user:" + args[0].(string)
					},
					[]string{"uid"},
				)
			},
			params:      map[string]interface{}{},
			wantResult:  "",
			wantErr:     true,
			errContains: "required param uid missing",
		},
		{
			name: "parameter order preserved",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"orderedParams",
					func(args ...interface{}) string {
						return args[0].(string) + "-" + args[1].(string) + "-" + args[2].(string)
					},
					[]string{"first", "second", "third"},
				)
			},
			params: map[string]interface{}{
				"third":  "c",
				"first":  "a",
				"second": "b",
			},
			wantResult: "a-b-c",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ckf, err := tt.setupFunc()
			require.NoError(t, err)

			result, err := ckf.Call(tt.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}

func TestKey_CustomKeyFunction_Call_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func() (CustomKeyFunction, error)
		params     map[string]interface{}
		wantErr    bool
		wantResult string
		checkFunc  func(t *testing.T, result string, err error)
	}{
		{
			name: "empty string param",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getUser",
					func(args ...interface{}) string {
						return "user:" + args[0].(string)
					},
					[]string{"uid"},
				)
			},
			params: map[string]interface{}{
				"uid": "",
			},
			wantErr:    false,
			wantResult: "user:",
			checkFunc: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				assert.Equal(t, "user:", result)
			},
		},
		{
			name: "zero value params",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getById",
					func(args ...interface{}) string {
						id := args[0].(int)
						return fmt.Sprintf("id:%d", id)
					},
					[]string{"id"},
				)
			},
			params: map[string]interface{}{
				"id": 0,
			},
			wantErr:    false,
			wantResult: "id:0",
			checkFunc: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				// Zero value should be passed through
				assert.Equal(t, "id:0", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ckf, err := tt.setupFunc()
			require.NoError(t, err)

			result, err := ckf.Call(tt.params)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result, err)
			}
		})
	}
}

func TestKey_CustomKeyFunction_Callable(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func() (CustomKeyFunction, error)
		args       []interface{}
		wantResult string
	}{
		{
			name: "single string argument",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getUser",
					func(args ...interface{}) string {
						return "user:" + args[0].(string)
					},
					[]string{"uid"},
				)
			},
			args:       []interface{}{"123"},
			wantResult: "user:123",
		},
		{
			name: "multiple arguments",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getUserByType",
					func(args ...interface{}) string {
						return "user:" + args[0].(string) + ":" + args[1].(string)
					},
					[]string{"uid", "type"},
				)
			},
			args:       []interface{}{"123", "admin"},
			wantResult: "user:123:admin",
		},
		{
			name: "mixed type arguments",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getUserById",
					func(args ...interface{}) string {
						return fmt.Sprintf("user:%d:%s", args[0].(int), args[1].(string))
					},
					[]string{"id", "name"},
				)
			},
			args:       []interface{}{65, "test"},
			wantResult: "user:65:test",
		},
		{
			name: "empty string argument",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getUser",
					func(args ...interface{}) string {
						return "user:" + args[0].(string)
					},
					[]string{"uid"},
				)
			},
			args:       []interface{}{""},
			wantResult: "user:",
		},
		{
			name: "zero value integer argument",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getById",
					func(args ...interface{}) string {
						return fmt.Sprintf("id:%d", args[0].(int))
					},
					[]string{"id"},
				)
			},
			args:       []interface{}{0},
			wantResult: "id:0",
		},
		{
			name: "no arguments",
			setupFunc: func() (CustomKeyFunction, error) {
				return NewCustomKeyFunction(
					"getConstant",
					func(args ...interface{}) string {
						return "constant"
					},
					[]string{"dummy"},
				)
			},
			args:       []interface{}{},
			wantResult: "constant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ckf, err := tt.setupFunc()
			require.NoError(t, err)

			result := ckf.Callable(tt.args...)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
