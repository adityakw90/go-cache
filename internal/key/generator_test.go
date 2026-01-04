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
			name: "missing key",
			data: map[string]string{
				"prefix": "cache",
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
