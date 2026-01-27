package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/adapter"
	"github.com/adityakw90/go-cache/internal/cache"
	"github.com/adityakw90/go-cache/internal/hash"
	"github.com/adityakw90/go-cache/internal/key"
	testutil "github.com/adityakw90/go-cache/test/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_KeyGenerator_Format(t *testing.T) {
	tests := []struct {
		name            string
		useHashKey      bool
		versioning      bool
		args            []interface{}
		wantPrefix      string
		wantContains    []string // substrings that should be in the key
		wantNotContains []string // substrings that should NOT be in the key
	}{
		{
			name:            "non-hashed key format",
			useHashKey:      false,
			versioning:      false,
			args:            []interface{}{"arg1"},
			wantPrefix:      "e2e_key_format",
			wantContains:    []string{"e2e_key_format", "testKey"},
			wantNotContains: []string{"-"},
		},
		{
			name:         "hashed key format includes hash",
			useHashKey:   true,
			versioning:   false,
			args:         []interface{}{"arg1"},
			wantPrefix:   "e2e_key_format",
			wantContains: []string{"e2e_key_format", "testKey", "-"},
		},
		{
			name:            "non-hashed key with versioning",
			useHashKey:      false,
			versioning:      true,
			args:            []interface{}{"arg1"},
			wantPrefix:      "e2e_key_format",
			wantContains:    []string{"e2e_key_format", "testKey", "v"},
			wantNotContains: []string{"-"},
		},
		{
			name:         "hashed key with versioning",
			useHashKey:   true,
			versioning:   true,
			args:         []interface{}{"arg1"},
			wantPrefix:   "e2e_key_format",
			wantContains: []string{"e2e_key_format", "testKey", "v"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := testutil.CreateTestRedisClient(t)
			defer client.Close()

			c, err := cache.NewCache(client, cache.Options{
				KeyPrefix:                 tt.wantPrefix,
				ExpireDefault:             1 * time.Hour,
				VersionExpire:             24 * time.Hour,
				LockDuration:              5 * time.Second,
				LockInterval:              100 * time.Millisecond,
				LogProvider:               func(ctx context.Context) adapter.Logger { return adapter.NewNoOpLogger() },
				Semaphore:                 adapter.NewSemaphore(10),
				KeyGenerator:              key.KeyGenerator,
				KeyVersionGenerator:       key.KeyVersionGenerator,
				KeyHashedVersionGenerator: key.KeyHashedVersionGenerator,
				VersionGenerator:          key.VersionGenerator,
				LockGenerator:             key.LockGenerator,
				StartSpan:                 adapter.NoOpStartSpan,
				StartChildSpan:            adapter.NoOpStartChildSpan,
			})
			require.NoError(t, err)

			ctx := context.Background()

			fn := func(ctx context.Context, args ...interface{}) (interface{}, error) {
				return "result", nil
			}

			cachedWrapper := c.Cached("testKey", 1*time.Minute, tt.versioning, "")
			cachedFunc := cachedWrapper(fn, nil, tt.useHashKey)

			var result string
			_, err = cachedFunc(&result, ctx, tt.args...)
			require.NoError(t, err)

			// Get the actual key format from Redis to verify
			// The key format depends on the generator used
			namespace := "testKey"
			version := 0
			if tt.versioning {
				version = 1
			}

			var expectedKey string
			if tt.useHashKey {
				// KeyHashedVersionGenerator format: {prefix}:{namespace}:v{version}-{key}.gob
				hashKey := hash.CacheKey(namespace, tt.args)
				expectedKey, err = c.KeyHashedVersionGenerator(map[string]string{
					"prefix":    tt.wantPrefix,
					"namespace": namespace,
					"version":   fmt.Sprintf("%d", version),
					"key":       hashKey,
				})
			} else {
				// KeyVersionGenerator format: {prefix}:{namespace}:v{version}.gob
				expectedKey, err = c.KeyVersionGenerator(map[string]string{
					"prefix":    tt.wantPrefix,
					"namespace": namespace,
					"version":   fmt.Sprintf("%d", version),
				})
			}
			require.NoError(t, err)

			// Verify the key format matches expected pattern
			assert.Contains(t, expectedKey, ".gob", "key should end with .gob extension")

			for _, want := range tt.wantContains {
				assert.Contains(t, expectedKey, want, "key should contain %s", want)
			}

			for _, notWant := range tt.wantNotContains {
				assert.NotContains(t, expectedKey, notWant, "key should NOT contain %s", notWant)
			}
		})
	}
}
