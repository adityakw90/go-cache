package version

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/internal/adapter"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion_InvalidateVersion(t *testing.T) {
	tracer := &adapter.NoOpTracer{}
	logger := &adapter.NoOpLogger{}
	semaphore := adapter.NewSemaphore(10)
	versionGenerator := key.VersionGenerator
	versionExpire := 1 * time.Hour
	keyPrefix := "test-key-prefix"
	namespace := "test-invalidate"
	prefix := "test-prefix"

	tests := []struct {
		name            string
		setupMock       func(*testing.T, redismock.ClientMock, string)
		prefix          string
		namespace       string
		customGenerator key.KeyGeneratorFunc // Optional: if nil, uses default versionGenerator
		wantErr         bool
		errContains     string
		checkFunc       func(*testing.T, redismock.ClientMock)
	}{
		{
			name: "invalidate new version",
			setupMock: func(t *testing.T, mock redismock.ClientMock, key string) {
				// Pipeline Incr
				mock.ExpectIncr(key).SetVal(1)
				// Async TTL check
				mock.ExpectTTL(key).SetVal(-1)
				mock.ExpectExpire(key, versionExpire).SetVal(true)
			},
			prefix:    prefix,
			namespace: namespace + "-new",
			wantErr:   false,
			checkFunc: func(t *testing.T, mock redismock.ClientMock) {
				time.Sleep(50 * time.Millisecond)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name: "invalidate existing version increments it",
			setupMock: func(t *testing.T, mock redismock.ClientMock, key string) {
				// Pipeline Incr (from 5 to 6)
				mock.ExpectIncr(key).SetVal(6)
				// Async TTL check
				mock.ExpectTTL(key).SetVal(-1)
				mock.ExpectExpire(key, versionExpire).SetVal(true)
			},
			prefix:    prefix,
			namespace: namespace + "-existing",
			wantErr:   false,
			checkFunc: func(t *testing.T, mock redismock.ClientMock) {
				time.Sleep(50 * time.Millisecond)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name: "empty prefix uses keyPrefix",
			setupMock: func(t *testing.T, mock redismock.ClientMock, key string) {
				// Pipeline Incr
				mock.ExpectIncr(key).SetVal(1)
				// Async TTL check
				mock.ExpectTTL(key).SetVal(-1)
				mock.ExpectExpire(key, versionExpire).SetVal(true)
			},
			prefix:    "",
			namespace: namespace + "-empty-prefix",
			wantErr:   false,
			checkFunc: func(t *testing.T, mock redismock.ClientMock) {
				time.Sleep(50 * time.Millisecond)
				assert.NoError(t, mock.ExpectationsWereMet())
			},
		},
		{
			name: "key generator error",
			setupMock: func(t *testing.T, mock redismock.ClientMock, key string) {
				// No Redis interactions expected - error occurs before any Redis calls
			},
			prefix:    prefix,
			namespace: namespace,
			customGenerator: func(data map[string]string) (string, error) {
				return "", errors.New("key generation failed")
			},
			wantErr:     true,
			errContains: "failed to increment version: failed to generate version key",
		},
		{
			name: "redis Incr error",
			setupMock: func(t *testing.T, mock redismock.ClientMock, key string) {
				// IncrementCacheVersion uses direct Redis calls, not pipeline
				// So the error happens in IncrementCacheVersion, not during Exec()
				mock.ExpectIncr(key).SetErr(errors.New("redis connection error"))
			},
			prefix:      prefix,
			namespace:   namespace + "-incr-error",
			wantErr:     true,
			errContains: "failed to increment version: failed to increment version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			defer client.Close()

			// Use custom generator if provided, otherwise use default
			testGenerator := versionGenerator
			if tt.customGenerator != nil {
				testGenerator = tt.customGenerator
			}

			// Generate key for mock setup (only if using default generator)
			var key string
			if tt.customGenerator == nil {
				testPrefix := tt.prefix
				if testPrefix == "" {
					testPrefix = keyPrefix
				}
				data := map[string]string{
					"prefix":    testPrefix,
					"namespace": tt.namespace,
				}
				var err error
				key, err = versionGenerator(data)
				require.NoError(t, err)
			}

			tt.setupMock(t, mock, key)
			ctx := context.Background()

			getSession := func() redis.Pipeliner {
				return client.Pipeline()
			}

			err := InvalidateVersion(
				ctx,
				client,
				tracer,
				logger,
				semaphore,
				testGenerator,
				versionExpire,
				keyPrefix,
				tt.namespace,
				tt.prefix,
				getSession,
			)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, mock)
				}
			}
		})
	}
}

func TestVersion_InvalidateVersion_PipelineExecution(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	tracer := &adapter.NoOpTracer{}
	logger := &adapter.NoOpLogger{}
	semaphore := adapter.NewSemaphore(10)
	versionGenerator := key.VersionGenerator
	versionExpire := 1 * time.Hour
	keyPrefix := "test-key-prefix"
	namespace := "test-pipeline-exec"
	prefix := "test-prefix"

	ctx := context.Background()

	data := map[string]string{
		"prefix":    prefix,
		"namespace": namespace,
	}
	key, err := versionGenerator(data)
	require.NoError(t, err)

	// Setup mock expectations
	mock.ExpectIncr(key).SetVal(11) // Increment from 10 to 11
	mock.ExpectTTL(key).SetVal(-1)
	mock.ExpectExpire(key, versionExpire).SetVal(true)

	getSession := func() redis.Pipeliner {
		return client.Pipeline()
	}

	err = InvalidateVersion(
		ctx,
		client,
		tracer,
		logger,
		semaphore,
		versionGenerator,
		versionExpire,
		keyPrefix,
		namespace,
		prefix,
		getSession,
	)
	require.NoError(t, err)

	// Wait for async operations
	time.Sleep(50 * time.Millisecond)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_InvalidateVersion_MultipleInvalidations(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	tracer := &adapter.NoOpTracer{}
	logger := &adapter.NoOpLogger{}
	semaphore := adapter.NewSemaphore(10)
	versionGenerator := key.VersionGenerator
	versionExpire := 1 * time.Hour
	keyPrefix := "test-key-prefix"
	namespace := "test-multiple"
	prefix := "test-prefix"

	ctx := context.Background()
	data := map[string]string{
		"prefix":    prefix,
		"namespace": namespace,
	}
	key, err := versionGenerator(data)
	require.NoError(t, err)

	// Setup mock expectations for 5 invalidations
	// Only set up Incr expectations - async TTL checks are tested separately in ttl_test.go
	// redismock has issues with async operations across goroutines
	for i := 1; i <= 5; i++ {
		mock.ExpectIncr(key).SetVal(int64(i))
	}

	getSession := func() redis.Pipeliner {
		return client.Pipeline()
	}

	// Invalidate multiple times
	for i := 0; i < 5; i++ {
		err := InvalidateVersion(
			ctx,
			client,
			tracer,
			logger,
			semaphore,
			versionGenerator,
			versionExpire,
			keyPrefix,
			namespace,
			prefix,
			getSession,
		)
		require.NoError(t, err)
	}

	// Verify expectations were met (only Incr, async TTL is tested separately)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_InvalidateVersion_DifferentNamespaces(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	tracer := &adapter.NoOpTracer{}
	logger := &adapter.NoOpLogger{}
	semaphore := adapter.NewSemaphore(10)
	versionGenerator := key.VersionGenerator
	versionExpire := 1 * time.Hour
	keyPrefix := "test-key-prefix"
	prefix := "test-prefix"

	namespaces := []string{"ns1", "ns2", "ns3"}
	ctx := context.Background()

	// Generate keys and setup mock expectations for each namespace
	keys := make([]string, len(namespaces))
	for i, ns := range namespaces {
		data := map[string]string{
			"prefix":    prefix,
			"namespace": ns,
		}
		key, err := versionGenerator(data)
		require.NoError(t, err)
		keys[i] = key

		// Setup expectations: Incr only (async TTL is tested separately)
		mock.ExpectIncr(key).SetVal(1)
	}

	getSession := func() redis.Pipeliner {
		return client.Pipeline()
	}

	// Invalidate each namespace
	for _, ns := range namespaces {
		err := InvalidateVersion(
			ctx,
			client,
			tracer,
			logger,
			semaphore,
			versionGenerator,
			versionExpire,
			keyPrefix,
			ns,
			prefix,
			getSession,
		)
		require.NoError(t, err)
	}

	// Verify expectations were met (only Incr, async TTL is tested separately)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVersion_InvalidateVersion_PipelineError(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()

	tracer := &adapter.NoOpTracer{}
	logger := &adapter.NoOpLogger{}
	semaphore := adapter.NewSemaphore(10)
	versionGenerator := key.VersionGenerator
	versionExpire := 1 * time.Hour
	keyPrefix := "test-key-prefix"
	namespace := "test-pipeline-error"
	prefix := "test-prefix"

	ctx := context.Background()

	data := map[string]string{
		"prefix":    prefix,
		"namespace": namespace,
	}
	key, err := versionGenerator(data)
	require.NoError(t, err)

	// Setup mock: Incr command fails
	// Note: IncrementCacheVersion uses direct Redis calls, not pipeline
	// So the error happens in IncrementCacheVersion, not during Exec()
	mock.ExpectIncr(key).SetErr(errors.New("redis connection error"))

	getSession := func() redis.Pipeliner {
		return client.Pipeline()
	}

	err = InvalidateVersion(
		ctx,
		client,
		tracer,
		logger,
		semaphore,
		versionGenerator,
		versionExpire,
		keyPrefix,
		namespace,
		prefix,
		getSession,
	)

	// Should get error during IncrementCacheVersion (before Exec is called)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to increment version: failed to increment version")
	assert.NoError(t, mock.ExpectationsWereMet())
}
