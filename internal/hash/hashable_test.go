package hash

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHash_MakeHashable_SimpleTypes tests makeHashable with primitive types.
func TestHash_MakeHashable_SimpleTypes(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "string",
			value: "test",
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "test", result)
			},
		},
		{
			name:  "int",
			value: 42,
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "42", result)
			},
		},
		{
			name:  "int64",
			value: int64(42),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "42", result)
			},
		},
		{
			name:  "int32",
			value: int32(42),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "42", result)
			},
		},
		{
			name:  "int16",
			value: int16(42),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "42", result)
			},
		},
		{
			name:  "int8",
			value: int8(42),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "42", result)
			},
		},
		{
			name:  "uint",
			value: uint(42),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "42", result)
			},
		},
		{
			name:  "uint64",
			value: uint64(42),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "42", result)
			},
		},
		{
			name:  "uint32",
			value: uint32(42),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "42", result)
			},
		},
		{
			name:  "uint16",
			value: uint16(42),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "42", result)
			},
		},
		{
			name:  "uint8",
			value: uint8(42),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "42", result)
			},
		},
		{
			name:  "float64",
			value: float64(3.14),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "3.14", result)
			},
		},
		{
			name:  "float32",
			value: float32(3.14),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "3.14", result)
			},
		},
		{
			name:  "bool true",
			value: true,
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "true", result)
			},
		},
		{
			name:  "bool false",
			value: false,
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "false", result)
				// true and false should produce different results
				buffer2 := &bytes.Buffer{}
				makeHashable(buffer2, true)
				assert.NotEqual(t, result, buffer2.String())
			},
		},
		{
			name:  "nil",
			value: nil,
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "<nil>", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			makeHashable(buffer, tt.value)
			result := buffer.String()
			tt.checkFunc(t, result)
		})
	}
}

// TestHash_MakeHashable_Collections tests makeHashable with map and slice types.
func TestHash_MakeHashable_Collections(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name: "map[string]interface{}",
			value: map[string]interface{}{
				"a": "value1",
				"b": 42,
			},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{a:value1,b:42}", result)
			},
		},
		{
			name:  "[]interface{}",
			value: []interface{}{"a", "b", "c"},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "[a,b,c]", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			makeHashable(buffer, tt.value)
			result := buffer.String()
			tt.checkFunc(t, result)
		})
	}
}

// TestHash_MakeHashable_Consistency tests that makeHashable produces consistent results.
func TestHash_MakeHashable_Consistency(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "string",
			value: "test",
		},
		{
			name:  "int",
			value: 42,
		},
		{
			name:  "float64",
			value: float64(3.14),
		},
		{
			name:  "bool",
			value: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer1 := &bytes.Buffer{}
			buffer2 := &bytes.Buffer{}
			makeHashable(buffer1, tt.value)
			makeHashable(buffer2, tt.value)
			assert.Equal(t, buffer1.String(), buffer2.String())
		})
	}
}
