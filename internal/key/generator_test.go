package key

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKey_KeyGenerator(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]string
		wantResult  string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid prefix and key",
			data: map[string]string{
				"prefix": "cache",
				"key":    "user:123",
			},
			wantResult: "cache:user:123.gob",
			wantErr:    false,
		},
		{
			name: "empty prefix",
			data: map[string]string{
				"prefix": "",
				"key":    "user:123",
			},
			wantResult: ":user:123.gob",
			wantErr:    false,
		},
		{
			name: "empty key",
			data: map[string]string{
				"prefix": "cache",
				"key":    "",
			},
			wantResult: "cache:.gob",
			wantErr:    false,
		},
		{
			name: "both empty",
			data: map[string]string{
				"prefix": "",
				"key":    "",
			},
			wantResult: ":.gob",
			wantErr:    false,
		},
		{
			name: "missing prefix",
			data: map[string]string{
				"key": "user:123",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name:        "empty data map",
			data:        map[string]string{},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name:        "nil data map",
			data:        nil,
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name: "extra fields ignored",
			data: map[string]string{
				"prefix": "cache",
				"key":    "user:123",
				"extra":  "ignored",
			},
			wantResult: "cache:user:123.gob",
			wantErr:    false,
		},
		{
			name: "special characters in prefix and key",
			data: map[string]string{
				"prefix": "cache-v1",
				"key":    "user:123:data",
			},
			wantResult: "cache-v1:user:123:data.gob",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := KeyGenerator(tt.data)

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

func TestKey_KeyVersionGenerator(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]string
		wantResult  string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid all parameters",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "1",
				"key":       "123",
			},
			wantResult: "cache:users:v1.gob",
			wantErr:    false,
		},
		{
			name: "version with multiple digits",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "123",
				"key":       "456",
			},
			wantResult: "cache:users:v123.gob",
			wantErr:    false,
		},
		{
			name: "empty prefix",
			data: map[string]string{
				"prefix":    "",
				"namespace": "users",
				"version":   "1",
				"key":       "123",
			},
			wantResult: ":users:v1.gob",
			wantErr:    false,
		},
		{
			name: "empty namespace",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "",
				"version":   "1",
				"key":       "123",
			},
			wantResult: "cache::v1.gob",
			wantErr:    false,
		},
		{
			name: "empty version",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "",
				"key":       "123",
			},
			wantResult: "cache:users:v.gob",
			wantErr:    false,
		},
		{
			name: "key field is ignored",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "1",
				"key":       "this-is-ignored",
			},
			wantResult: "cache:users:v1.gob",
			wantErr:    false,
		},
		{
			name: "missing prefix",
			data: map[string]string{
				"namespace": "users",
				"version":   "1",
				"key":       "123",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name: "missing namespace",
			data: map[string]string{
				"prefix":  "cache",
				"version": "1",
				"key":     "123",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'namespace' in key generator data",
		},
		{
			name: "missing version",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"key":       "123",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'version' in key generator data",
		},
		{
			name:        "empty data map",
			data:        map[string]string{},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name:        "nil data map",
			data:        nil,
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name: "extra fields ignored",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "1",
				"key":       "123",
				"extra":     "ignored",
			},
			wantResult: "cache:users:v1.gob",
			wantErr:    false,
		},
		{
			name: "special characters in all fields",
			data: map[string]string{
				"prefix":    "cache-v1",
				"namespace": "users:admin",
				"version":   "2.0",
				"key":       "user:123:data",
			},
			wantResult: "cache-v1:users:admin:v2.0.gob",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := KeyVersionGenerator(tt.data)

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

func TestKey_VersionGenerator(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]string
		wantResult  string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid prefix and namespace",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
			},
			wantResult: "cache:users:version",
			wantErr:    false,
		},
		{
			name: "empty prefix",
			data: map[string]string{
				"prefix":    "",
				"namespace": "users",
			},
			wantResult: ":users:version",
			wantErr:    false,
		},
		{
			name: "empty namespace",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "",
			},
			wantResult: "cache::version",
			wantErr:    false,
		},
		{
			name: "both empty",
			data: map[string]string{
				"prefix":    "",
				"namespace": "",
			},
			wantResult: "::version",
			wantErr:    false,
		},
		{
			name: "missing prefix",
			data: map[string]string{
				"namespace": "users",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name: "missing namespace",
			data: map[string]string{
				"prefix": "cache",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'namespace' in key generator data",
		},
		{
			name:        "empty data map",
			data:        map[string]string{},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name:        "nil data map",
			data:        nil,
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name: "extra fields ignored",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"extra":     "ignored",
			},
			wantResult: "cache:users:version",
			wantErr:    false,
		},
		{
			name: "special characters in prefix and namespace",
			data: map[string]string{
				"prefix":    "cache-v1",
				"namespace": "users:admin",
			},
			wantResult: "cache-v1:users:admin:version",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VersionGenerator(tt.data)

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

func TestKey_LockGenerator(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]string
		wantResult  string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid prefix and namespace",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
			},
			wantResult: "cache:users:lock",
			wantErr:    false,
		},
		{
			name: "empty prefix",
			data: map[string]string{
				"prefix":    "",
				"namespace": "users",
			},
			wantResult: ":users:lock",
			wantErr:    false,
		},
		{
			name: "empty namespace",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "",
			},
			wantResult: "cache::lock",
			wantErr:    false,
		},
		{
			name: "both empty",
			data: map[string]string{
				"prefix":    "",
				"namespace": "",
			},
			wantResult: "::lock",
			wantErr:    false,
		},
		{
			name: "missing prefix",
			data: map[string]string{
				"namespace": "users",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name: "missing namespace",
			data: map[string]string{
				"prefix": "cache",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'namespace' in key generator data",
		},
		{
			name:        "empty data map",
			data:        map[string]string{},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name:        "nil data map",
			data:        nil,
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name: "extra fields ignored",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"extra":     "ignored",
			},
			wantResult: "cache:users:lock",
			wantErr:    false,
		},
		{
			name: "special characters in prefix and namespace",
			data: map[string]string{
				"prefix":    "cache-v1",
				"namespace": "users:admin",
			},
			wantResult: "cache-v1:users:admin:lock",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := LockGenerator(tt.data)

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

func TestKey_KeyHashedVersionGenerator(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]string
		wantResult  string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid all parameters",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "1",
				"key":       "123",
			},
			wantResult: "cache:users:v1-123.gob",
			wantErr:    false,
		},
		{
			name: "version with multiple digits",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "123",
				"key":       "456",
			},
			wantResult: "cache:users:v123-456.gob",
			wantErr:    false,
		},
		{
			name: "empty prefix",
			data: map[string]string{
				"prefix":    "",
				"namespace": "users",
				"version":   "1",
				"key":       "123",
			},
			wantResult: ":users:v1-123.gob",
			wantErr:    false,
		},
		{
			name: "empty namespace",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "",
				"version":   "1",
				"key":       "123",
			},
			wantResult: "cache::v1-123.gob",
			wantErr:    false,
		},
		{
			name: "empty version",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "",
				"key":       "123",
			},
			wantResult: "cache:users:v-123.gob",
			wantErr:    false,
		},
		{
			name: "empty key",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "1",
				"key":       "",
			},
			wantResult: "cache:users:v1-.gob",
			wantErr:    false,
		},
		{
			name: "all empty",
			data: map[string]string{
				"prefix":    "",
				"namespace": "",
				"version":   "",
				"key":       "",
			},
			wantResult: "::v-.gob",
			wantErr:    false,
		},
		{
			name: "missing prefix",
			data: map[string]string{
				"namespace": "users",
				"version":   "1",
				"key":       "123",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name: "missing namespace",
			data: map[string]string{
				"prefix":  "cache",
				"version": "1",
				"key":     "123",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'namespace' in key generator data",
		},
		{
			name: "missing version",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"key":       "123",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'version' in key generator data",
		},
		{
			name: "missing key",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "1",
			},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'key' in key generator data",
		},
		{
			name:        "empty data map",
			data:        map[string]string{},
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name:        "nil data map",
			data:        nil,
			wantResult:  "",
			wantErr:     true,
			errContains: "missing 'prefix' in key generator data",
		},
		{
			name: "extra fields ignored",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "1",
				"key":       "123",
				"extra":     "ignored",
			},
			wantResult: "cache:users:v1-123.gob",
			wantErr:    false,
		},
		{
			name: "special characters in all fields",
			data: map[string]string{
				"prefix":    "cache-v1",
				"namespace": "users:admin",
				"version":   "2.0",
				"key":       "user:123:data",
			},
			wantResult: "cache-v1:users:admin:v2.0-user:123:data.gob",
			wantErr:    false,
		},
		{
			name: "complex key with colons",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "1",
				"key":       "session:abc123:token:xyz789",
			},
			wantResult: "cache:users:v1-session:abc123:token:xyz789.gob",
			wantErr:    false,
		},
		{
			name: "numeric only values",
			data: map[string]string{
				"prefix":    "123",
				"namespace": "456",
				"version":   "789",
				"key":       "012",
			},
			wantResult: "123:456:v789-012.gob",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := KeyHashedVersionGenerator(tt.data)

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

func TestKey_KeyHashedVersionGenerator_ErrorTypes(t *testing.T) {
	tests := []struct {
		name      string
		data      map[string]string
		wantError error
	}{
		{
			name:      "missing prefix returns correct error",
			data:      map[string]string{"namespace": "users", "version": "1", "key": "123"},
			wantError: ErrKeyGeneratorParamsPrefixRequired,
		},
		{
			name:      "missing namespace returns correct error",
			data:      map[string]string{"prefix": "cache", "version": "1", "key": "123"},
			wantError: ErrKeyGeneratorParamsNamespaceRequired,
		},
		{
			name:      "missing version returns correct error",
			data:      map[string]string{"prefix": "cache", "namespace": "users", "key": "123"},
			wantError: ErrKeyGeneratorParamsVersionRequired,
		},
		{
			name:      "missing key returns correct error",
			data:      map[string]string{"prefix": "cache", "namespace": "users", "version": "1"},
			wantError: ErrKeyGeneratorParamsKeyRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := KeyHashedVersionGenerator(tt.data)
			require.Error(t, err)
			assert.Equal(t, tt.wantError, err)
		})
	}
}

func TestKey_KeyHashedVersionGenerator_Vs_KeyVersionGenerator(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]string
		wantHashed  string
		wantNonHash string
	}{
		{
			name: "simple key",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "1",
				"key":       "123",
			},
			wantHashed:  "cache:users:v1-123.gob",
			wantNonHash: "cache:users:v1.gob",
		},
		{
			name: "key with special characters",
			data: map[string]string{
				"prefix":    "cache",
				"namespace": "users",
				"version":   "2",
				"key":       "user:123:profile",
			},
			wantHashed:  "cache:users:v2-user:123:profile.gob",
			wantNonHash: "cache:users:v2.gob",
		},
		{
			name: "key with underscores and hyphens",
			data: map[string]string{
				"prefix":    "my-app",
				"namespace": "user_sessions",
				"version":   "10",
				"key":       "session-abc-123",
			},
			wantHashed:  "my-app:user_sessions:v10-session-abc-123.gob",
			wantNonHash: "my-app:user_sessions:v10.gob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedResult, err := KeyHashedVersionGenerator(tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.wantHashed, hashedResult)

			nonHashedResult, err := KeyVersionGenerator(tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.wantNonHash, nonHashedResult)

			// Verify hashed version includes the key while non-hashed does not
			assert.Contains(t, hashedResult, tt.data["key"])
			assert.NotContains(t, nonHashedResult, tt.data["key"])
		})
	}
}

func TestKey_NewCustomKeyFunction_Errors(t *testing.T) {
	tests := []struct {
		name      string
		nameArg   string
		callable  func(args ...interface{}) string
		params    []string
		wantError error
	}{
		{
			name:      "empty name returns error",
			nameArg:   "",
			callable:  func(args ...interface{}) string { return "test" },
			params:    []string{"param"},
			wantError: ErrNameRequired,
		},
		{
			name:      "nil callable returns error",
			nameArg:   "testFunc",
			callable:  nil,
			params:    []string{"param"},
			wantError: ErrCallableRequired,
		},
		{
			name:      "empty params returns error",
			nameArg:   "testFunc",
			callable:  func(args ...interface{}) string { return "test" },
			params:    []string{},
			wantError: ErrParamsRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ckf, err := NewCustomKeyFunction(tt.nameArg, tt.callable, tt.params)
			assert.Nil(t, ckf)
			assert.Error(t, err)
			assert.Equal(t, tt.wantError, err)
		})
	}
}

func TestKey_KeyGenerators_Consistency(t *testing.T) {
	// Test that all generators produce consistent output format
	baseData := map[string]string{
		"prefix":    "app",
		"namespace": "cache",
	}

	// KeyGenerator only needs prefix and key
	keyOnlyData := map[string]string{
		"prefix": "app",
		"key":    "mykey",
	}

	t.Run("KeyGenerator format", func(t *testing.T) {
		result, err := KeyGenerator(keyOnlyData)
		require.NoError(t, err)
		assert.Equal(t, "app:mykey.gob", result)
	})

	t.Run("KeyVersionGenerator format", func(t *testing.T) {
		data := copyMap(baseData)
		data["version"] = "5"
		result, err := KeyVersionGenerator(data)
		require.NoError(t, err)
		assert.Equal(t, "app:cache:v5.gob", result)
	})

	t.Run("KeyHashedVersionGenerator format", func(t *testing.T) {
		data := copyMap(baseData)
		data["version"] = "5"
		data["key"] = "mykey"
		result, err := KeyHashedVersionGenerator(data)
		require.NoError(t, err)
		assert.Equal(t, "app:cache:v5-mykey.gob", result)
	})

	t.Run("VersionGenerator format", func(t *testing.T) {
		result, err := VersionGenerator(baseData)
		require.NoError(t, err)
		assert.Equal(t, "app:cache:version", result)
	})

	t.Run("LockGenerator format", func(t *testing.T) {
		result, err := LockGenerator(baseData)
		require.NoError(t, err)
		assert.Equal(t, "app:cache:lock", result)
	})
}

// copyMap creates a shallow copy of a map
func copyMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
