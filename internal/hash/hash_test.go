package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHash_CacheKey tests the main public API function.
func TestHash_CacheKey(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		args     []interface{}
		wantHash string
		wantLen  int
	}{
		{
			name:     "simple string args",
			funcName: "getUser",
			args:     []interface{}{"user123"},
			wantHash: "53263b0e01790b37f53a6427e48f5420",
			wantLen:  32, // MD5 hex string length
		},
		{
			name:     "multiple args",
			funcName: "getUser",
			args:     []interface{}{"user123", 42},
			wantHash: "3ce24f92ca1b1e38b258baa1aae44935",
			wantLen:  32,
		},
		{
			name:     "empty args",
			funcName: "getUser",
			args:     []interface{}{},
			wantHash: "117d99544c66c7b022da4f95e648776a",
			wantLen:  32,
		},
		{
			name:     "map args",
			funcName: "getUser",
			args:     []interface{}{map[string]interface{}{"id": "123", "name": "test"}},
			wantHash: "d67e540e55a306c4e1fc4f484ccdf723",
			wantLen:  32,
		},
		{
			name:     "slice args",
			funcName: "getUser",
			args:     []interface{}{[]interface{}{"a", "b", "c"}},
			wantHash: "fa56bf8abd3d8b9a9c08ba27d2988a33",
			wantLen:  32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := CacheKey(tt.funcName, tt.args)
			assert.Equal(t, tt.wantLen, len(hash))
			assert.Equal(t, tt.wantHash, hash)
		})
	}
}

// TestHash_CacheKey_Consistency verifies that the same inputs produce the same hash.
func TestHash_CacheKey_Consistency(t *testing.T) {
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
			hash1 := CacheKey(tt.funcName, tt.args)
			hash2 := CacheKey(tt.funcName, tt.args)
			hash3 := CacheKey(tt.funcName, tt.args)

			if tt.checkFunc != nil {
				tt.checkFunc(t, hash1, hash2, hash3)
			}
		})
	}
}

// TestHash_CacheKey_DifferentArgs verifies that different inputs produce different hashes.
func TestHash_CacheKey_DifferentArgs(t *testing.T) {
	tests := []struct {
		name      string
		funcName1 string
		funcName2 string
		args1     []interface{}
		args2     []interface{}
		checkFunc func(t *testing.T, hash1, hash2 string)
	}{
		{
			name:      "different args produce different hashes",
			funcName1: "getUser",
			funcName2: "getUser",
			args1:     []interface{}{"user123"},
			args2:     []interface{}{"user456"},
			checkFunc: func(t *testing.T, hash1, hash2 string) {
				// Different args should produce different hashes
				assert.NotEqual(t, hash1, hash2)
			},
		},
		{
			name:      "different function names produce different hashes",
			funcName1: "getUser",
			funcName2: "getProduct",
			args1:     []interface{}{"user123"},
			args2:     []interface{}{"user123"},
			checkFunc: func(t *testing.T, hash1, hash2 string) {
				assert.NotEqual(t, hash1, hash2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := CacheKey(tt.funcName1, tt.args1)
			hash2 := CacheKey(tt.funcName2, tt.args2)

			if tt.checkFunc != nil {
				tt.checkFunc(t, hash1, hash2)
			}
		})
	}
}

// TestHash_CacheKey_EmptyFunctionName tests CacheKey with empty function name.
func TestHash_CacheKey_EmptyFunctionName(t *testing.T) {
	tests := []struct {
		name      string
		funcName  string
		args      []interface{}
		checkFunc func(t *testing.T, hash string)
	}{
		{
			name:     "empty function name with args",
			funcName: "",
			args:     []interface{}{"user123"},
			checkFunc: func(t *testing.T, hash string) {
				assert.Equal(t, 32, len(hash))
				// Should produce consistent results
				hash2 := CacheKey("", []interface{}{"user123"})
				assert.Equal(t, hash, hash2)
			},
		},
		{
			name:     "empty function name with empty args",
			funcName: "",
			args:     []interface{}{},
			checkFunc: func(t *testing.T, hash string) {
				assert.Equal(t, 32, len(hash))
				// Should produce consistent results
				hash2 := CacheKey("", []interface{}{})
				assert.Equal(t, hash, hash2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := CacheKey(tt.funcName, tt.args)
			if tt.checkFunc != nil {
				tt.checkFunc(t, hash)
			}
		})
	}
}
