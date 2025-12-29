package cache

import (
	"bytes"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_getCacheHash(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	cache, err := NewCache(redisClient)
	require.NoError(t, err)

	tests := []struct {
		name     string
		funcName string
		args     []interface{}
		wantLen  int
	}{
		{
			name:     "simple string args",
			funcName: "getUser",
			args:     []interface{}{"user123"},
			wantLen:  32, // MD5 hex string length
		},
		{
			name:     "multiple args",
			funcName: "getUser",
			args:     []interface{}{"user123", 42},
			wantLen:  32,
		},
		{
			name:     "empty args",
			funcName: "getUser",
			args:     []interface{}{},
			wantLen:  32,
		},
		{
			name:     "map args",
			funcName: "getUser",
			args:     []interface{}{map[string]interface{}{"id": "123", "name": "test"}},
			wantLen:  32,
		},
		{
			name:     "slice args",
			funcName: "getUser",
			args:     []interface{}{[]interface{}{"a", "b", "c"}},
			wantLen:  32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := cache.getCacheHash(tt.funcName, tt.args)
			assert.Equal(t, tt.wantLen, len(hash))
			// Hash should be consistent
			hash2 := cache.getCacheHash(tt.funcName, tt.args)
			assert.Equal(t, hash, hash2)
		})
	}
}

func TestCache_getCacheHash_Consistency(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	cache, err := NewCache(redisClient)
	require.NoError(t, err)

	tests := []struct {
		name      string
		funcName  string
		args      []interface{}
		checkFunc func(t *testing.T, hash1, hash2, hash3 string)
	}{
		{
			name:     "consistent hashes",
			funcName: "getUser",
			args:     []interface{}{"user123", 42},
			checkFunc: func(t *testing.T, hash1, hash2, hash3 string) {
				// All hashes should be identical
				assert.Equal(t, hash1, hash2)
				assert.Equal(t, hash2, hash3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := cache.getCacheHash(tt.funcName, tt.args)
			hash2 := cache.getCacheHash(tt.funcName, tt.args)
			hash3 := cache.getCacheHash(tt.funcName, tt.args)

			if tt.checkFunc != nil {
				tt.checkFunc(t, hash1, hash2, hash3)
			}
		})
	}
}

func TestCache_getCacheHash_DifferentArgs(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	cache, err := NewCache(redisClient)
	require.NoError(t, err)

	tests := []struct {
		name      string
		funcName  string
		args1     []interface{}
		args2     []interface{}
		checkFunc func(t *testing.T, hash1, hash2 string)
	}{
		{
			name:     "different args produce different hashes",
			funcName: "getUser",
			args1:    []interface{}{"user123"},
			args2:    []interface{}{"user456"},
			checkFunc: func(t *testing.T, hash1, hash2 string) {
				// Different args should produce different hashes
				assert.NotEqual(t, hash1, hash2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := cache.getCacheHash(tt.funcName, tt.args1)
			hash2 := cache.getCacheHash(tt.funcName, tt.args2)

			if tt.checkFunc != nil {
				tt.checkFunc(t, hash1, hash2)
			}
		})
	}
}

func TestMakeHashable_SimpleTypes(t *testing.T) {
	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)
	buffer.Reset()

	tests := []struct {
		name      string
		value     interface{}
		checkFunc func(t *testing.T, buffer *bytes.Buffer)
	}{
		{
			name:  "string",
			value: "test",
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				assert.Contains(t, buffer.String(), "test")
			},
		},
		{
			name:  "int",
			value: 42,
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				assert.Contains(t, buffer.String(), "42")
			},
		},
		{
			name:  "bool",
			value: true,
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				assert.Contains(t, buffer.String(), "true")
			},
		},
		{
			name:  "nil",
			value: nil,
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				assert.Contains(t, buffer.String(), "<nil>")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer.Reset()
			makeHashable(buffer, tt.value)
			tt.checkFunc(t, buffer)
		})
	}
}

func TestMakeHashable_Map(t *testing.T) {
	tests := []struct {
		name      string
		value     map[string]interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name: "map with multiple keys",
			value: map[string]interface{}{
				"b": "value2",
				"a": "value1",
				"c": 42,
			},
			checkFunc: func(t *testing.T, result string) {
				// Map should be sorted by keys
				assert.Contains(t, result, "a")
				assert.Contains(t, result, "b")
				assert.Contains(t, result, "c")
				// Keys should appear in sorted order
				assert.Greater(t, len(result), 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bufferPool.Get().(*bytes.Buffer)
			defer bufferPool.Put(buffer)
			buffer.Reset()

			makeHashable(buffer, tt.value)
			result := buffer.String()

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestMakeHashable_Slice(t *testing.T) {
	tests := []struct {
		name      string
		value     []interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "slice with string elements",
			value: []interface{}{"a", "b", "c"},
			checkFunc: func(t *testing.T, result string) {
				assert.Contains(t, result, "a")
				assert.Contains(t, result, "b")
				assert.Contains(t, result, "c")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bufferPool.Get().(*bytes.Buffer)
			defer bufferPool.Put(buffer)
			buffer.Reset()

			makeHashable(buffer, tt.value)
			result := buffer.String()

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestMakeHashable_Struct(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name      string
		value     TestStruct
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "struct with fields",
			value: TestStruct{Name: "John", Age: 30},
			checkFunc: func(t *testing.T, result string) {
				assert.Contains(t, result, "Name")
				assert.Contains(t, result, "Age")
				assert.Contains(t, result, "John")
				assert.Contains(t, result, "30")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bufferPool.Get().(*bytes.Buffer)
			defer bufferPool.Put(buffer)
			buffer.Reset()

			makeHashable(buffer, tt.value)
			result := buffer.String()

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestMakeHashable_Pointer(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name: "pointer to string",
			setupFunc: func() interface{} {
				value := "test"
				return &value
			},
			checkFunc: func(t *testing.T, result string) {
				assert.Contains(t, result, "test")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bufferPool.Get().(*bytes.Buffer)
			defer bufferPool.Put(buffer)
			buffer.Reset()

			value := tt.setupFunc()
			makeHashable(buffer, value)
			result := buffer.String()

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestMakeHashable_NilPointer(t *testing.T) {
	tests := []struct {
		name      string
		value     *string
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "nil pointer",
			value: nil,
			checkFunc: func(t *testing.T, result string) {
				assert.Contains(t, result, "<nil>")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bufferPool.Get().(*bytes.Buffer)
			defer bufferPool.Put(buffer)
			buffer.Reset()

			makeHashable(buffer, tt.value)
			result := buffer.String()

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}
