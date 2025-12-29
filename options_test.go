package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultOptions(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, opts *options)
	}{
		{
			name: "default options values",
			checkFunc: func(t *testing.T, opts *options) {
				assert.Equal(t, "CACHE", opts.keyPrefix)
				assert.Equal(t, time.Minute, opts.expireDefault)
				assert.Equal(t, 30*24*time.Hour, opts.versionExpire)
				assert.Equal(t, time.Minute, opts.lockDuration)
				assert.Equal(t, 100*time.Millisecond, opts.lockInterval)
				assert.Equal(t, 10, opts.semaphoreSize)
				assert.NotNil(t, opts.tracer)
				assert.NotNil(t, opts.logger)
				assert.NotNil(t, opts.keyGenerator)
				assert.NotNil(t, opts.keyVersionGenerator)
				assert.NotNil(t, opts.versionGenerator)
				assert.NotNil(t, opts.lockGenerator)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts)
			}
		})
	}
}

func TestWithKeyPrefix(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		checkFunc func(t *testing.T, opts *options)
	}{
		{
			name:   "set key prefix",
			prefix: "myapp",
			checkFunc: func(t *testing.T, opts *options) {
				assert.Equal(t, "myapp", opts.keyPrefix)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			WithKeyPrefix(tt.prefix)(opts)
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts)
			}
		})
	}
}

func TestWithExpireDefault(t *testing.T) {
	tests := []struct {
		name      string
		duration  time.Duration
		checkFunc func(t *testing.T, opts *options)
	}{
		{
			name:     "set expire default",
			duration: 5 * time.Minute,
			checkFunc: func(t *testing.T, opts *options) {
				assert.Equal(t, 5*time.Minute, opts.expireDefault)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			WithExpireDefault(tt.duration)(opts)
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts)
			}
		})
	}
}

func TestWithVersionExpire(t *testing.T) {
	tests := []struct {
		name      string
		duration  time.Duration
		checkFunc func(t *testing.T, opts *options)
	}{
		{
			name:     "set version expire",
			duration: 7 * 24 * time.Hour,
			checkFunc: func(t *testing.T, opts *options) {
				assert.Equal(t, 7*24*time.Hour, opts.versionExpire)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			WithVersionExpire(tt.duration)(opts)
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts)
			}
		})
	}
}

func TestWithLockDuration(t *testing.T) {
	tests := []struct {
		name      string
		duration  time.Duration
		checkFunc func(t *testing.T, opts *options)
	}{
		{
			name:     "set lock duration",
			duration: 2 * time.Minute,
			checkFunc: func(t *testing.T, opts *options) {
				assert.Equal(t, 2*time.Minute, opts.lockDuration)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			WithLockDuration(tt.duration)(opts)
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts)
			}
		})
	}
}

func TestWithLockInterval(t *testing.T) {
	tests := []struct {
		name      string
		duration  time.Duration
		checkFunc func(t *testing.T, opts *options)
	}{
		{
			name:     "set lock interval",
			duration: 200 * time.Millisecond,
			checkFunc: func(t *testing.T, opts *options) {
				assert.Equal(t, 200*time.Millisecond, opts.lockInterval)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			WithLockInterval(tt.duration)(opts)
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts)
			}
		})
	}
}

func TestWithSemaphoreSize(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		checkFunc func(t *testing.T, opts *options)
	}{
		{
			name: "set semaphore size",
			size: 20,
			checkFunc: func(t *testing.T, opts *options) {
				assert.Equal(t, 20, opts.semaphoreSize)
			},
		},
		{
			name: "zero size defaults to 1",
			size: 0,
			checkFunc: func(t *testing.T, opts *options) {
				// Should default to 1
				assert.Equal(t, 1, opts.semaphoreSize)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			WithSemaphoreSize(tt.size)(opts)
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts)
			}
		})
	}
}

func TestWithTracer(t *testing.T) {
	tests := []struct {
		name      string
		tracer    Tracer
		checkFunc func(t *testing.T, opts *options, originalTracer Tracer)
	}{
		{
			name:   "set tracer",
			tracer: &NoOpTracer{},
			checkFunc: func(t *testing.T, opts *options, originalTracer Tracer) {
				assert.Equal(t, &NoOpTracer{}, opts.tracer)
			},
		},
		{
			name:   "nil tracer does not change",
			tracer: nil,
			checkFunc: func(t *testing.T, opts *options, originalTracer Tracer) {
				// Should not change if nil is passed
				assert.Equal(t, originalTracer, opts.tracer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			originalTracer := opts.tracer
			WithTracer(tt.tracer)(opts)
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts, originalTracer)
			}
		})
	}
}

func TestWithLogger(t *testing.T) {
	tests := []struct {
		name      string
		logger    Logger
		checkFunc func(t *testing.T, opts *options, originalLogger Logger)
	}{
		{
			name:   "set logger",
			logger: &NoOpLogger{},
			checkFunc: func(t *testing.T, opts *options, originalLogger Logger) {
				assert.Equal(t, &NoOpLogger{}, opts.logger)
			},
		},
		{
			name:   "nil logger does not change",
			logger: nil,
			checkFunc: func(t *testing.T, opts *options, originalLogger Logger) {
				// Should not change if nil is passed
				assert.Equal(t, originalLogger, opts.logger)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			originalLogger := opts.logger
			WithLogger(tt.logger)(opts)
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts, originalLogger)
			}
		})
	}
}

func TestWithSemaphore(t *testing.T) {
	tests := []struct {
		name      string
		semaphore Semaphore
		checkFunc func(t *testing.T, opts *options, originalSemaphore Semaphore, setSemaphore Semaphore)
	}{
		{
			name:      "set semaphore",
			semaphore: NewDefaultSemaphore(5),
			checkFunc: func(t *testing.T, opts *options, originalSemaphore Semaphore, setSemaphore Semaphore) {
				assert.NotNil(t, opts.semaphore)
				// Verify it's a different instance (not the original)
				assert.NotEqual(t, originalSemaphore, opts.semaphore)
				// Verify it's the one we set
				assert.Equal(t, setSemaphore, opts.semaphore)
			},
		},
		{
			name:      "nil semaphore does not change",
			semaphore: nil,
			checkFunc: func(t *testing.T, opts *options, originalSemaphore Semaphore, setSemaphore Semaphore) {
				// Should not change if nil is passed
				assert.Equal(t, originalSemaphore, opts.semaphore)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			originalSemaphore := opts.semaphore
			WithSemaphore(tt.semaphore)(opts)
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts, originalSemaphore, tt.semaphore)
			}
		})
	}
}

func TestWithKeyGenerator(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, opts *options)
	}{
		{
			name: "set key generator",
			checkFunc: func(t *testing.T, opts *options) {
				assert.NotNil(t, opts.keyGenerator)
				result, err := opts.keyGenerator(map[string]string{"key": "test"})
				assert.NoError(t, err)
				assert.Equal(t, "custom:test", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			customFn := func(data map[string]string) (string, error) {
				return "custom:" + data["key"], nil
			}
			WithKeyGenerator(customFn)(opts)
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts)
			}
		})
	}
}

func TestWithKeyVersionGenerator(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, opts *options)
	}{
		{
			name: "set key version generator",
			checkFunc: func(t *testing.T, opts *options) {
				assert.NotNil(t, opts.keyVersionGenerator)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			customFn := func(data map[string]string) (string, error) {
				return "custom:" + data["version"], nil
			}
			WithKeyVersionGenerator(customFn)(opts)
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts)
			}
		})
	}
}

func TestWithVersionGenerator(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, opts *options)
	}{
		{
			name: "set version generator",
			checkFunc: func(t *testing.T, opts *options) {
				assert.NotNil(t, opts.versionGenerator)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			customFn := func(data map[string]string) (string, error) {
				return "custom:" + data["namespace"], nil
			}
			WithVersionGenerator(customFn)(opts)
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts)
			}
		})
	}
}

func TestWithLockGenerator(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, opts *options)
	}{
		{
			name: "set lock generator",
			checkFunc: func(t *testing.T, opts *options) {
				assert.NotNil(t, opts.lockGenerator)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions()
			customFn := func(data map[string]string) (string, error) {
				return "custom:" + data["namespace"], nil
			}
			WithLockGenerator(customFn)(opts)
			if tt.checkFunc != nil {
				tt.checkFunc(t, opts)
			}
		})
	}
}

func TestDefaultKeyGenerator(t *testing.T) {
	tests := []struct {
		name       string
		data       map[string]string
		wantErr    bool
		wantResult string
		checkFunc  func(t *testing.T, result string, err error)
	}{
		{
			name: "valid data",
			data: map[string]string{
				"prefix": "myapp",
				"key":    "testkey",
			},
			wantErr:    false,
			wantResult: "myapp:testkey",
			checkFunc: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "myapp:testkey", result)
			},
		},
		{
			name: "missing prefix",
			data: map[string]string{
				"key": "testkey",
			},
			wantErr: true,
			checkFunc: func(t *testing.T, result string, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "missing 'prefix'")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := defaultKeyGenerator(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, result, err)
			}
		})
	}
}

func TestDefaultKeyVersionGenerator(t *testing.T) {
	tests := []struct {
		name       string
		data       map[string]string
		wantResult string
		checkFunc  func(t *testing.T, result string, err error)
	}{
		{
			name: "valid data",
			data: map[string]string{
				"prefix":    "myapp",
				"namespace": "getUser",
				"version":   "1",
				"key":       "abc123",
			},
			wantResult: "myapp:getUser:v1-abc123.gob",
			checkFunc: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "myapp:getUser:v1-abc123.gob", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := defaultKeyVersionGenerator(tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, result)
			if tt.checkFunc != nil {
				tt.checkFunc(t, result, err)
			}
		})
	}
}

func TestDefaultVersionGenerator(t *testing.T) {
	tests := []struct {
		name       string
		data       map[string]string
		wantResult string
		checkFunc  func(t *testing.T, result string, err error)
	}{
		{
			name: "valid data",
			data: map[string]string{
				"prefix":    "myapp",
				"namespace": "getUser",
			},
			wantResult: "myapp:getUser:version",
			checkFunc: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "myapp:getUser:version", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := defaultVersionGenerator(tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, result)
			if tt.checkFunc != nil {
				tt.checkFunc(t, result, err)
			}
		})
	}
}

func TestDefaultLockGenerator(t *testing.T) {
	tests := []struct {
		name       string
		data       map[string]string
		wantResult string
		checkFunc  func(t *testing.T, result string, err error)
	}{
		{
			name: "valid data",
			data: map[string]string{
				"prefix":    "myapp",
				"namespace": "getUser",
			},
			wantResult: "myapp:getUser:lock",
			checkFunc: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "myapp:getUser:lock", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := defaultLockGenerator(tt.data)
			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, result)
			if tt.checkFunc != nil {
				tt.checkFunc(t, result, err)
			}
		})
	}
}
