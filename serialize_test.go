package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_serialize(t *testing.T) {
	redisClient := createTestRedisClient(t)
	cache, err := NewCache(redisClient)
	require.NoError(t, err)

	tests := []struct {
		name      string
		value     interface{}
		wantErr   bool
		checkFunc func(t *testing.T, data []byte, original interface{})
	}{
		{
			name:    "string value",
			value:   "test string",
			wantErr: false,
			checkFunc: func(t *testing.T, data []byte, original interface{}) {
				assert.NotNil(t, data)
				assert.Greater(t, len(data), 0)
			},
		},
		{
			name:    "int value",
			value:   42,
			wantErr: false,
			checkFunc: func(t *testing.T, data []byte, original interface{}) {
				assert.NotNil(t, data)
				assert.Greater(t, len(data), 0)
			},
		},
		{
			name:    "struct value",
			value:   struct{ Name string }{Name: "test"},
			wantErr: false,
			checkFunc: func(t *testing.T, data []byte, original interface{}) {
				assert.NotNil(t, data)
				assert.Greater(t, len(data), 0)
			},
		},
		{
			name:    "slice value",
			value:   []string{"a", "b", "c"},
			wantErr: false,
			checkFunc: func(t *testing.T, data []byte, original interface{}) {
				assert.NotNil(t, data)
				assert.Greater(t, len(data), 0)
			},
		},
		{
			name:    "map value",
			value:   map[string]int{"a": 1, "b": 2},
			wantErr: false,
			checkFunc: func(t *testing.T, data []byte, original interface{}) {
				assert.NotNil(t, data)
				assert.Greater(t, len(data), 0)
			},
		},
		{
			name:    "nil value",
			value:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := cache.serialize(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, data, tt.value)
				}
			}
		})
	}
}

func TestCache_deserialize(t *testing.T) {
	redisClient := createTestRedisClient(t)
	cache, err := NewCache(redisClient)
	require.NoError(t, err)

	tests := []struct {
		name      string
		value     interface{}
		result    interface{}
		wantErr   bool
		checkFunc func(t *testing.T, original, result interface{})
	}{
		{
			name:    "string value",
			value:   "test string",
			result:  new(string),
			wantErr: false,
			checkFunc: func(t *testing.T, original, result interface{}) {
				assert.Equal(t, original, *(result.(*string)))
			},
		},
		{
			name:    "int value",
			value:   42,
			result:  new(int),
			wantErr: false,
			checkFunc: func(t *testing.T, original, result interface{}) {
				assert.Equal(t, original, *(result.(*int)))
			},
		},
		{
			name:    "struct value",
			value:   struct{ Name string }{Name: "test"},
			result:  &struct{ Name string }{},
			wantErr: false,
			checkFunc: func(t *testing.T, original, result interface{}) {
				resultPtr := result.(*struct{ Name string })
				assert.Equal(t, original, *resultPtr)
			},
		},
		{
			name:    "slice value",
			value:   []string{"a", "b", "c"},
			result:  &[]string{},
			wantErr: false,
			checkFunc: func(t *testing.T, original, result interface{}) {
				assert.Equal(t, original, *(result.(*[]string)))
			},
		},
		{
			name:    "empty data",
			value:   "test",
			result:  new(string),
			wantErr: true,
		},
		{
			name:    "nil result",
			value:   "test",
			result:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize first
			data, err := cache.serialize(tt.value)
			if tt.wantErr && tt.name == "empty data" {
				// For empty data test, use empty slice
				data = []byte{}
			} else {
				require.NoError(t, err)
			}

			// Deserialize
			err = cache.deserialize(data, tt.result)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, tt.value, tt.result)
				}
			}
		})
	}
}

func TestCache_serialize_deserialize_RoundTrip(t *testing.T) {
	redisClient := createTestRedisClient(t)
	cache, err := NewCache(redisClient)
	require.NoError(t, err)

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
			data, err := cache.serialize(tt.value)
			require.NoError(t, err)
			require.NotNil(t, data)
			assert.Greater(t, len(data), 0)

			// Deserialize
			err = cache.deserialize(data, tt.result)
			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, tt.value, tt.result)
			}
		})
	}
}
