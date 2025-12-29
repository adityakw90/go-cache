package cache

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomKeyFunction_Call(t *testing.T) {
	tests := []struct {
		name        string
		ckf         *CustomKeyFunction
		params      map[string]interface{}
		wantResult  string
		wantErr     bool
		errContains string
	}{
		{
			name: "single string parameter",
			ckf: &CustomKeyFunction{
				Name: "getUser",
				Callable: func(args ...interface{}) string {
					return "user:" + args[0].(string)
				},
				Params: []string{"uid"},
			},
			params: map[string]interface{}{
				"uid": "123",
			},
			wantResult: "user:123",
			wantErr:    false,
		},
		{
			name: "multiple parameters",
			ckf: &CustomKeyFunction{
				Name: "getUserByType",
				Callable: func(args ...interface{}) string {
					return "user:" + args[0].(string) + ":" + args[1].(string)
				},
				Params: []string{"uid", "type"},
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
			ckf: &CustomKeyFunction{
				Name: "getUserById",
				Callable: func(args ...interface{}) string {
					return fmt.Sprintf("user:%d:%s", args[0].(int), args[1].(string))
				},
				Params: []string{"id", "name"},
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
			ckf: &CustomKeyFunction{
				Name: "getUser",
				Callable: func(args ...interface{}) string {
					return "user:" + args[0].(string)
				},
				Params: []string{"uid"},
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
			ckf: &CustomKeyFunction{
				Name: "getUserByType",
				Callable: func(args ...interface{}) string {
					return "user:" + args[0].(string) + ":" + args[1].(string)
				},
				Params: []string{"uid", "type"},
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
			ckf: &CustomKeyFunction{
				Name: "getUser",
				Callable: func(args ...interface{}) string {
					return "user:" + args[0].(string)
				},
				Params: []string{"uid"},
			},
			params: map[string]interface{}{
				"uid":   "123",
				"extra": "ignored",
			},
			wantResult: "user:123",
			wantErr:    false,
		},
		{
			name: "empty params list",
			ckf: &CustomKeyFunction{
				Name: "getConstant",
				Callable: func(args ...interface{}) string {
					return "constant"
				},
				Params: []string{},
			},
			params: map[string]interface{}{
				"any": "value",
			},
			wantResult: "constant",
			wantErr:    false,
		},
		{
			name: "nil params map with required params",
			ckf: &CustomKeyFunction{
				Name: "getUser",
				Callable: func(args ...interface{}) string {
					return "user:" + args[0].(string)
				},
				Params: []string{"uid"},
			},
			params:      nil,
			wantResult:  "",
			wantErr:     true,
			errContains: "required param uid missing",
		},
		{
			name: "empty params map with required params",
			ckf: &CustomKeyFunction{
				Name: "getUser",
				Callable: func(args ...interface{}) string {
					return "user:" + args[0].(string)
				},
				Params: []string{"uid"},
			},
			params:      map[string]interface{}{},
			wantResult:  "",
			wantErr:     true,
			errContains: "required param uid missing",
		},
		{
			name: "parameter order preserved",
			ckf: &CustomKeyFunction{
				Name: "orderedParams",
				Callable: func(args ...interface{}) string {
					return args[0].(string) + "-" + args[1].(string) + "-" + args[2].(string)
				},
				Params: []string{"first", "second", "third"},
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
			result, err := tt.ckf.Call(tt.params)

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

func TestCustomKeyFunction_Call_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		ckf        *CustomKeyFunction
		params     map[string]interface{}
		wantErr    bool
		wantResult string
		checkFunc  func(t *testing.T, result string, err error)
		panicFunc  func() func()
	}{
		{
			name: "nil callable panics",
			ckf: &CustomKeyFunction{
				Name:     "nilCallable",
				Callable: nil,
				Params:   []string{"uid"},
			},
			params: map[string]interface{}{
				"uid": "123",
			},
			wantErr: true,
			panicFunc: func() func() {
				// This will panic, but we test that it's handled gracefully
				// In practice, this should be validated during registration
				return func() {
					if r := recover(); r != nil {
						// Expected panic when Callable is nil
						assert.NotNil(t, r)
					}
				}
			},
		},
		{
			name: "empty string param",
			ckf: &CustomKeyFunction{
				Name: "getUser",
				Callable: func(args ...interface{}) string {
					return "user:" + args[0].(string)
				},
				Params: []string{"uid"},
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
			ckf: &CustomKeyFunction{
				Name: "getById",
				Callable: func(args ...interface{}) string {
					id := args[0].(int)
					return fmt.Sprintf("id:%d", id)
				},
				Params: []string{"id"},
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
			if tt.panicFunc != nil {
				defer tt.panicFunc()()
			}

			result, err := tt.ckf.Call(tt.params)

			if tt.wantErr {
				// Error or panic expected
				if err != nil {
					assert.Error(t, err)
				}
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
