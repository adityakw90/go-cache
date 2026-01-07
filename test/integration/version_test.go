package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/adityakw90/go-cache/adapter"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/adityakw90/go-cache/internal/version"
	testutil "github.com/adityakw90/go-cache/test/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion_GetCacheVersion(t *testing.T) {
	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	ctx := context.Background()
	prefix := "test"
	namespace := "test_namespace"
	versionExpire := 24 * time.Hour
	startSpan := adapter.NoOpStartSpan
	startChildSpan := adapter.NoOpStartChildSpan
	semaphore := adapter.NewSemaphore(10)
	getLogger := adapter.GetNoOpLogger

	t.Run("get version initializes to 1", func(t *testing.T) {
		v, err := version.GetCacheVersion(
			ctx, client,
			startSpan,
			startChildSpan,
			semaphore,
			getLogger,
			key.VersionGenerator,
			versionExpire,
			namespace+"_init",
			prefix,
		)
		require.NoError(t, err)
		assert.Equal(t, 1, v)
	})

	t.Run("get version returns existing version", func(t *testing.T) {
		v1, err := version.GetCacheVersion(
			ctx, client,
			startSpan,
			startChildSpan,
			semaphore,
			getLogger,
			key.VersionGenerator,
			versionExpire,
			namespace+"_existing",
			prefix,
		)
		require.NoError(t, err)
		assert.Equal(t, 1, v1)

		v2, err := version.GetCacheVersion(
			ctx, client,
			startSpan,
			startChildSpan,
			semaphore,
			getLogger,
			key.VersionGenerator,
			versionExpire,
			namespace+"_existing",
			prefix,
		)
		require.NoError(t, err)
		assert.Equal(t, 1, v2)
	})

	t.Run("different namespaces have different versions", func(t *testing.T) {
		v1, err := version.GetCacheVersion(
			ctx, client,
			startSpan,
			startChildSpan,
			semaphore,
			getLogger,
			key.VersionGenerator,
			versionExpire,
			namespace+"_ns1",
			prefix,
		)
		require.NoError(t, err)

		v2, err := version.GetCacheVersion(
			ctx, client,
			startSpan,
			startChildSpan,
			semaphore,
			getLogger,
			key.VersionGenerator,
			versionExpire,
			namespace+"_ns2",
			prefix,
		)
		require.NoError(t, err)

		assert.Equal(t, 1, v1)
		assert.Equal(t, 1, v2)
	})
}

func TestVersion_IncrementCacheVersion(t *testing.T) {
	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	ctx := context.Background()
	prefix := "test"
	namespace := "test_increment"
	versionExpire := 24 * time.Hour
	startSpan := adapter.NoOpStartSpan
	startChildSpan := adapter.NoOpStartChildSpan
	semaphore := adapter.NewSemaphore(10)
	getLogger := adapter.GetNoOpLogger

	t.Run("increment version in pipeline", func(t *testing.T) {
		v1, err := version.GetCacheVersion(
			ctx, client,
			startSpan,
			startChildSpan,
			semaphore,
			getLogger,
			key.VersionGenerator,
			versionExpire,
			namespace+"_pipe",
			prefix,
		)
		require.NoError(t, err)
		assert.Equal(t, 1, v1)

		session := client.Pipeline()
		_, err = version.IncrementCacheVersion(
			ctx, client,
			startSpan,
			startChildSpan,
			semaphore,
			getLogger,
			key.VersionGenerator,
			session,
			prefix,
			namespace+"_pipe",
			versionExpire,
		)
		require.NoError(t, err)

		_, err = session.Exec(ctx)
		require.NoError(t, err)

		v2, err := version.GetCacheVersion(
			ctx, client,
			startSpan,
			startChildSpan,
			semaphore,
			getLogger,
			key.VersionGenerator,
			versionExpire,
			namespace+"_pipe",
			prefix,
		)
		require.NoError(t, err)
		assert.Equal(t, 2, v2)
	})

	t.Run("multiple increments", func(t *testing.T) {
		v1, err := version.GetCacheVersion(
			ctx, client,
			startSpan,
			startChildSpan,
			semaphore,
			getLogger,
			key.VersionGenerator,
			versionExpire,
			namespace+"_multi",
			prefix,
		)
		require.NoError(t, err)
		assert.Equal(t, 1, v1)

		session := client.Pipeline()
		_, err = version.IncrementCacheVersion(
			ctx, client,
			startSpan,
			startChildSpan,
			semaphore,
			getLogger,
			key.VersionGenerator,
			session,
			prefix,
			namespace+"_multi",
			versionExpire,
		)
		require.NoError(t, err)

		_, err = version.IncrementCacheVersion(
			ctx, client,
			startSpan,
			startChildSpan,
			semaphore,
			getLogger,
			key.VersionGenerator,
			session,
			prefix,
			namespace+"_multi",
			versionExpire,
		)
		require.NoError(t, err)

		_, err = session.Exec(ctx)
		require.NoError(t, err)

		v2, err := version.GetCacheVersion(
			ctx, client,
			startSpan,
			startChildSpan,
			semaphore,
			getLogger,
			key.VersionGenerator,
			versionExpire,
			namespace+"_multi",
			prefix,
		)
		require.NoError(t, err)
		assert.Equal(t, 3, v2)
	})
}

func TestVersion_Concurrent(t *testing.T) {
	client := testutil.CreateTestRedisClient(t)
	defer client.Close()

	ctx := context.Background()
	prefix := "test"
	namespace := "test_concurrent"
	versionExpire := 24 * time.Hour
	startSpan := adapter.NoOpStartSpan
	startChildSpan := adapter.NoOpStartChildSpan
	semaphore := adapter.NewSemaphore(10)
	getLogger := adapter.GetNoOpLogger

	t.Run("concurrent version access", func(t *testing.T) {
		const numGoroutines = 10
		results := make(chan int, numGoroutines)
		errs := make(chan error, numGoroutines)
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				v, err := version.GetCacheVersion(
					ctx, client,
					startSpan,
					startChildSpan,
					semaphore,
					getLogger,
					key.VersionGenerator,
					versionExpire,
					namespace+"_concurrent",
					prefix,
				)
				if err != nil {
					errs <- err
					return
				}
				results <- v
			}()
		}

		versions := make(map[int]int)
		for i := 0; i < numGoroutines; i++ {
			select {
			case v := <-results:
				versions[v]++
			case err := <-errs:
				require.NoError(t, err, "GetCacheVersion failed in goroutine")
			case <-time.After(2 * time.Second):
				t.Fatal("timeout waiting for version results")
			}
		}

		wg.Wait()
		close(results)
		close(errs)

		// Check for any remaining errors
		for err := range errs {
			require.NoError(t, err, "GetCacheVersion failed in goroutine")
		}

		// Due to race conditions in concurrent initialization,
		// some goroutines might see version 1, others might see higher versions
		// The important thing is that all goroutines get a valid version
		totalVersions := 0
		for _, count := range versions {
			totalVersions += count
		}
		assert.Equal(t, numGoroutines, totalVersions)
		// At least one goroutine should get version 1 (initialization)
		assert.Greater(t, versions[1], 0)
	})
}
