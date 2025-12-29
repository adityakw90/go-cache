package hash

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHash_handleMapString tests handleMapString function.
func TestHash_handleMapString(t *testing.T) {
	tests := []struct {
		name      string
		value     map[string]interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name: "map with single key",
			value: map[string]interface{}{
				"a": "value1",
			},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{a:value1}", result)
			},
		},
		{
			name: "map with multiple keys",
			value: map[string]interface{}{
				"b": "value2",
				"a": "value1",
				"c": 42,
			},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{a:value1,b:value2,c:42}", result)
			},
		},
		{
			name:  "empty map",
			value: map[string]interface{}{},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{}", result)
			},
		},
		{
			name: "map with nested values",
			value: map[string]interface{}{
				"user": map[string]interface{}{
					"id":   123,
					"name": "John",
				},
			},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{user:{id:123,name:John}}", result)
			},
		},
		{
			name: "keys are sorted",
			value: map[string]interface{}{
				"z": "last",
				"a": "first",
				"m": "middle",
			},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{a:first,m:middle,z:last}", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			handleMapString(buffer, tt.value)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_handleSlice tests handleSlice function.
func TestHash_handleSlice(t *testing.T) {
	tests := []struct {
		name      string
		value     []interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "slice with string elements",
			value: []interface{}{"a", "b", "c"},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "[a,b,c]", result)
			},
		},
		{
			name:  "empty slice",
			value: []interface{}{},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "[]", result)
			},
		},
		{
			name:  "slice with mixed types",
			value: []interface{}{"a", 1, true, 3.14},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "[a,1,true,3.14]", result)
			},
		},
		{
			name:  "slice order matters",
			value: []interface{}{"a", "b"},
			checkFunc: func(t *testing.T, result string) {
				assert.NotEqual(t, "[b,a]", result)
			},
		},
		{
			name:  "nested slices",
			value: []interface{}{[]interface{}{"a", "b"}, []interface{}{"c", "d"}},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "[[a,b],[c,d]]", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			handleSlice(buffer, tt.value)
			tt.checkFunc(t, buffer.String())
		})
	}
}
