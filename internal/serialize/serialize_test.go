package serialize

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerialize_Serialize(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		wantErr     bool
		wantErrType error
		wantErrMsg  string
		checkFunc   func(t *testing.T, data []byte, original interface{})
	}{
		{
			name:        "string value",
			value:       "test string",
			wantErr:     false,
			wantErrType: nil,
			wantErrMsg:  "",
			checkFunc: func(t *testing.T, data []byte, original interface{}) {
				assert.NotNil(t, data)
				assert.Greater(t, len(data), 0)
			},
		},
		{
			name:        "int value",
			value:       42,
			wantErr:     false,
			wantErrType: nil,
			wantErrMsg:  "",
			checkFunc: func(t *testing.T, data []byte, original interface{}) {
				assert.NotNil(t, data)
				assert.Greater(t, len(data), 0)
			},
		},
		{
			name:        "struct value",
			value:       struct{ Name string }{Name: "test"},
			wantErr:     false,
			wantErrType: nil,
			wantErrMsg:  "",
			checkFunc: func(t *testing.T, data []byte, original interface{}) {
				assert.NotNil(t, data)
				assert.Greater(t, len(data), 0)
			},
		},
		{
			name:        "slice value",
			value:       []string{"a", "b", "c"},
			wantErr:     false,
			wantErrType: nil,
			wantErrMsg:  "",
			checkFunc: func(t *testing.T, data []byte, original interface{}) {
				assert.NotNil(t, data)
				assert.Greater(t, len(data), 0)
			},
		},
		{
			name:        "map value",
			value:       map[string]int{"a": 1, "b": 2},
			wantErr:     false,
			wantErrType: nil,
			wantErrMsg:  "",
			checkFunc: func(t *testing.T, data []byte, original interface{}) {
				assert.NotNil(t, data)
				assert.Greater(t, len(data), 0)
			},
		},
		{
			name: "map[string]interface{} value",
			value: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": []int{1, 2, 3},
			},
			wantErr:     false,
			wantErrType: nil,
			wantErrMsg:  "",
			checkFunc: func(t *testing.T, data []byte, original interface{}) {
				assert.NotNil(t, data)
				assert.Greater(t, len(data), 0)
			},
		},
		{
			name:        "nil value",
			value:       nil,
			wantErr:     true,
			wantErrType: ErrSerializeNilValue,
			wantErrMsg:  "cannot serialize nil value",
			checkFunc:   nil,
		},
		{
			name:        "channel cannot be serialized",
			value:       make(chan int),
			wantErr:     true,
			wantErrType: nil,
			wantErrMsg:  "failed to serialize value: gob NewTypeObject can't handle type: chan int",
		},
		{
			name:        "function cannot be serialized",
			value:       func() {},
			wantErr:     true,
			wantErrType: nil,
			wantErrMsg:  "failed to serialize value: gob NewTypeObject can't handle type: func()",
		},
		{
			name:        "nil channel",
			value:       (chan int)(nil),
			wantErr:     true,
			wantErrType: nil,
			wantErrMsg:  "failed to serialize value: gob NewTypeObject can't handle type: chan int",
		},
		{
			name:        "nil function",
			value:       (func())(nil),
			wantErr:     true,
			wantErrType: nil,
			wantErrMsg:  "failed to serialize value: gob NewTypeObject can't handle type: func()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Serialize(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, data)
				if tt.wantErrType != nil {
					assert.True(t, errors.Is(err, tt.wantErrType))
				}
				if tt.wantErrMsg != "" {
					assert.Equal(t, tt.wantErrMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, data)
				assert.Greater(t, len(data), 0)
			}
		})
	}
}

func TestSerialize_Deserialize(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		changeData  []byte
		result      interface{}
		wantErr     bool
		wantErrType error
		wantErrMsg  string
		checkFunc   func(t *testing.T, original, result interface{})
	}{
		{
			name:        "string value",
			value:       "test string",
			changeData:  nil,
			result:      new(string),
			wantErr:     false,
			wantErrType: nil,
			wantErrMsg:  "",
			checkFunc: func(t *testing.T, original, result interface{}) {
				assert.Equal(t, original, *(result.(*string)))
			},
		},
		{
			name:        "int value",
			value:       42,
			changeData:  nil,
			result:      new(int),
			wantErr:     false,
			wantErrType: nil,
			wantErrMsg:  "",
			checkFunc: func(t *testing.T, original, result interface{}) {
				assert.Equal(t, original, *(result.(*int)))
			},
		},
		{
			name:        "struct value",
			value:       struct{ Name string }{Name: "test"},
			changeData:  nil,
			result:      &struct{ Name string }{},
			wantErr:     false,
			wantErrType: nil,
			wantErrMsg:  "",
			checkFunc: func(t *testing.T, original, result interface{}) {
				resultPtr := result.(*struct{ Name string })
				assert.Equal(t, original, *resultPtr)
			},
		},
		{
			name:        "slice value",
			value:       []string{"a", "b", "c"},
			changeData:  nil,
			result:      &[]string{},
			wantErr:     false,
			wantErrType: nil,
			wantErrMsg:  "",
			checkFunc: func(t *testing.T, original, result interface{}) {
				assert.Equal(t, original, *(result.(*[]string)))
			},
		},
		{
			name:        "empty data",
			value:       "test",
			changeData:  []byte{},
			result:      new(string),
			wantErr:     true,
			wantErrType: ErrDeserializeEmptyData,
			wantErrMsg:  "cannot deserialize empty data",
		},
		{
			name:        "nil result",
			value:       "test",
			changeData:  nil,
			result:      nil,
			wantErr:     true,
			wantErrType: ErrDeserializeResultNil,
			wantErrMsg:  "result interface cannot be nil",
		},
		{
			name:        "failed deserialize",
			value:       "test",
			changeData:  nil,
			result:      (*string)(nil),
			wantErr:     true,
			wantErrType: nil,
			wantErrMsg:  "failed to deserialize data: gob: DecodeValue of unassignable value",
		},
		{
			name:        "invalid gob data",
			value:       "test",
			changeData:  []byte{0xFF, 0xFF, 0xFF, 0xFF},
			result:      new(string),
			wantErr:     true,
			wantErrType: nil,
			wantErrMsg:  "failed to deserialize data: unexpected EOF",
		},
		{
			name:        "corrupted data",
			value:       "test",
			changeData:  []byte("not a valid gob stream"),
			result:      new(string),
			wantErr:     true,
			wantErrType: nil,
			wantErrMsg:  "failed to deserialize data: unexpected EOF",
		},
		{
			name:        "type mismatch - serialize string, deserialize int",
			value:       "test",
			changeData:  nil,
			result:      new(int),
			wantErr:     true,
			wantErrType: nil,
			wantErrMsg:  "failed to deserialize data: gob: decoding into local type *int, received remote type string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize first
			data, err := Serialize(tt.value)
			if tt.changeData != nil {
				data = tt.changeData
			} else {
				require.NoError(t, err)
			}

			// Deserialize
			err = Deserialize(data, tt.result)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.True(t, errors.Is(err, tt.wantErrType))
				}
				if tt.wantErrMsg != "" {
					assert.Equal(t, tt.wantErrMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tt.result)
				assert.Greater(t, len(data), 0)
			}
		})
	}
}

func TestSerialize_RoundTrip(t *testing.T) {
	type TestStruct struct {
		Name  string
		Age   int
		Items []string
	}

	tests := []struct {
		name      string
		value     interface{}
		result    interface{}
		checkFunc func(t *testing.T, original, result interface{})
	}{
		{
			name:   "complex struct",
			value:  TestStruct{Name: "John", Age: 30, Items: []string{"item1", "item2"}},
			result: &TestStruct{},
			checkFunc: func(t *testing.T, original, result interface{}) {
				resultPtr := result.(*TestStruct)
				assert.Equal(t, original, *resultPtr)
			},
		},
		{
			name:   "string round-trip",
			value:  "round-trip test string",
			result: new(string),
			checkFunc: func(t *testing.T, original, result interface{}) {
				assert.Equal(t, original, *(result.(*string)))
			},
		},
		{
			name:   "int round-trip",
			value:  12345,
			result: new(int),
			checkFunc: func(t *testing.T, original, result interface{}) {
				assert.Equal(t, original, *(result.(*int)))
			},
		},
		{
			name:   "slice round-trip",
			value:  []int{1, 2, 3, 4, 5},
			result: &[]int{},
			checkFunc: func(t *testing.T, original, result interface{}) {
				assert.Equal(t, original, *(result.(*[]int)))
			},
		},
		{
			name:   "map round-trip",
			value:  map[string]string{"key1": "value1", "key2": "value2"},
			result: &map[string]string{},
			checkFunc: func(t *testing.T, original, result interface{}) {
				assert.Equal(t, original, *(result.(*map[string]string)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := Serialize(tt.value)
			require.NoError(t, err)
			require.NotNil(t, data)
			assert.Greater(t, len(data), 0)

			// Deserialize
			err = Deserialize(data, tt.result)
			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, tt.value, tt.result)
			}
		})
	}
}
