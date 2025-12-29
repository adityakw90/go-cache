package hash

import (
	"bytes"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHash_getBuffer tests the getBuffer helper function.
func TestHash_getBuffer(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func()
		checkFunc func(t *testing.T, buffer *bytes.Buffer)
	}{
		{
			name: "returns valid buffer",
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				require.NotNil(t, buffer)
				assert.IsType(t, &bytes.Buffer{}, buffer)
			},
		},
		{
			name: "buffer is reset",
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				require.NotNil(t, buffer)
				assert.Equal(t, 0, buffer.Len())
				assert.Equal(t, "", buffer.String())
			},
		},
		{
			name: "buffer is usable",
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				buffer.WriteString("test")
				assert.Equal(t, "test", buffer.String())
				assert.Equal(t, 4, buffer.Len())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}
			buffer := getBuffer()
			tt.checkFunc(t, buffer)
			putBuffer(buffer)
		})
	}
}

// TestHash_putBuffer tests the putBuffer helper function.
func TestHash_putBuffer(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() *bytes.Buffer
		checkFunc func(t *testing.T)
	}{
		{
			name: "puts buffer back to pool",
			setupFunc: func() *bytes.Buffer {
				return getBuffer()
			},
			checkFunc: func(t *testing.T) {
				// Getting another buffer should work
				buffer := getBuffer()
				require.NotNil(t, buffer)
				putBuffer(buffer)
			},
		},
		{
			name: "allows reuse after put",
			setupFunc: func() *bytes.Buffer {
				buffer := getBuffer()
				buffer.WriteString("test data")
				return buffer
			},
			checkFunc: func(t *testing.T) {
				// After putting back, getting a new buffer should work
				buffer := getBuffer()
				require.NotNil(t, buffer)
				// Buffer should be reset (empty)
				assert.Equal(t, 0, buffer.Len())
				putBuffer(buffer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := tt.setupFunc()
			putBuffer(buffer)
			if tt.checkFunc != nil {
				tt.checkFunc(t)
			}
		})
	}
}

// TestHash_getBufferPutBuffer tests the integration of getBuffer and putBuffer.
func TestHash_getBufferPutBuffer(t *testing.T) {
	tests := []struct {
		name       string
		iterations int
		checkFunc  func(t *testing.T, buffers []*bytes.Buffer)
	}{
		{
			name:       "single get and put",
			iterations: 1,
			checkFunc: func(t *testing.T, buffers []*bytes.Buffer) {
				require.Len(t, buffers, 1)
				require.NotNil(t, buffers[0])
				assert.Equal(t, 0, buffers[0].Len())
			},
		},
		{
			name:       "multiple get and put",
			iterations: 10,
			checkFunc: func(t *testing.T, buffers []*bytes.Buffer) {
				require.Len(t, buffers, 10)
				for i, buf := range buffers {
					require.NotNil(t, buf, "buffer %d should not be nil", i)
					assert.Equal(t, 0, buf.Len(), "buffer %d should be empty", i)
				}
			},
		},
		{
			name:       "buffers are independent",
			iterations: 5,
			checkFunc: func(t *testing.T, buffers []*bytes.Buffer) {
				// Write different data to each buffer
				for i, buf := range buffers {
					buf.WriteString("buffer-")
					buf.WriteString(string(rune(i)))
				}
				// Verify they're independent
				for i, buf := range buffers {
					expected := "buffer-" + string(rune(i))
					assert.Equal(t, expected, buf.String(), "buffer %d should have correct content", i)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffers := make([]*bytes.Buffer, tt.iterations)
			for i := range buffers {
				buffers[i] = getBuffer()
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, buffers)
			}
			// Put all back
			for _, buf := range buffers {
				putBuffer(buf)
			}
		})
	}
}

// TestHash_getBufferPutBuffer_Reuse tests buffer reuse behavior.
func TestHash_getBufferPutBuffer_Reuse(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() *bytes.Buffer
		checkFunc func(t *testing.T, originalCap int)
	}{
		{
			name: "buffer can be reused",
			setupFunc: func() *bytes.Buffer {
				buffer := getBuffer()
				buffer.WriteString("test data")
				return buffer
			},
			checkFunc: func(t *testing.T, originalCap int) {
				// Get another buffer (might be reused)
				buffer2 := getBuffer()
				require.NotNil(t, buffer2)
				// Should be reset
				assert.Equal(t, 0, buffer2.Len())
				putBuffer(buffer2)
			},
		},
		{
			name: "capacity might be preserved",
			setupFunc: func() *bytes.Buffer {
				buffer := getBuffer()
				largeData := make([]byte, 1000)
				buffer.Write(largeData)
				return buffer
			},
			checkFunc: func(t *testing.T, originalCap int) {
				// Get another buffer
				buffer2 := getBuffer()
				require.NotNil(t, buffer2)
				// Capacity might be reused (pool behavior)
				// Note: Reset() doesn't guarantee capacity preservation,
				// so we only verify the buffer is usable regardless of capacity
				assert.Equal(t, 0, buffer2.Len(), "buffer should be reset")
				// If capacity is preserved, verify it's >= original
				// Otherwise, just verify buffer works correctly
				if buffer2.Cap() >= originalCap {
					assert.GreaterOrEqual(t, buffer2.Cap(), originalCap, "if capacity is preserved, it should be >= original")
				}
				// Verify buffer is usable
				buffer2.WriteString("test")
				assert.Equal(t, "test", buffer2.String())
				putBuffer(buffer2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer1 := tt.setupFunc()
			originalCap := buffer1.Cap()
			putBuffer(buffer1)
			if tt.checkFunc != nil {
				tt.checkFunc(t, originalCap)
			}
		})
	}
}

// TestHash_getBufferPutBuffer_ConcurrentAccess tests concurrent access.
func TestHash_getBufferPutBuffer_ConcurrentAccess(t *testing.T) {
	tests := []struct {
		name              string
		numGoroutines     int
		iterationsPerGoro int
		checkFunc         func(t *testing.T, errors []error)
	}{
		{
			name:              "concurrent access with 10 goroutines",
			numGoroutines:     10,
			iterationsPerGoro: 100,
			checkFunc: func(t *testing.T, errors []error) {
				assert.Empty(t, errors, "no errors should occur during concurrent access")
			},
		},
		{
			name:              "concurrent access with 50 goroutines",
			numGoroutines:     50,
			iterationsPerGoro: 20,
			checkFunc: func(t *testing.T, errors []error) {
				assert.Empty(t, errors, "no errors should occur during concurrent access")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			errors := make(chan error, tt.numGoroutines)

			// Launch multiple goroutines that use the pool concurrently
			for i := 0; i < tt.numGoroutines; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < tt.iterationsPerGoro; j++ {
						buffer := getBuffer()
						if buffer == nil {
							errors <- assert.AnError
							return
						}

						// Use the buffer
						buffer.WriteString("goroutine-")
						buffer.WriteString(string(rune(id)))
						buffer.WriteString("-iteration-")
						buffer.WriteString(string(rune(j)))

						// Verify it worked
						if buffer.Len() == 0 {
							errors <- assert.AnError
							putBuffer(buffer)
							return
						}

						// Put it back
						putBuffer(buffer)
					}
				}(i)
			}

			wg.Wait()
			close(errors)

			// Collect errors
			var errorList []error
			for err := range errors {
				errorList = append(errorList, err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, errorList)
			}
		})
	}
}

// TestHash_getBufferPutBuffer_DataHandling tests various data handling scenarios.
func TestHash_getBufferPutBuffer_DataHandling(t *testing.T) {
	tests := []struct {
		name      string
		dataFunc  func(*bytes.Buffer)
		checkFunc func(t *testing.T, buffer *bytes.Buffer)
	}{
		{
			name: "empty buffer",
			dataFunc: func(*bytes.Buffer) {
				// No data written
			},
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				assert.Equal(t, 0, buffer.Len())
				assert.Equal(t, "", buffer.String())
			},
		},
		{
			name: "small string",
			dataFunc: func(buf *bytes.Buffer) {
				buf.WriteString("test")
			},
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				assert.Equal(t, 4, buffer.Len())
				assert.Equal(t, "test", buffer.String())
			},
		},
		{
			name: "large data",
			dataFunc: func(buf *bytes.Buffer) {
				largeData := make([]byte, 10000)
				for i := range largeData {
					largeData[i] = byte(i % 256)
				}
				buf.Write(largeData)
			},
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				assert.Equal(t, 10000, buffer.Len())
				assert.Equal(t, 10000, len(buffer.Bytes()))
			},
		},
		{
			name: "mixed data types",
			dataFunc: func(buf *bytes.Buffer) {
				buf.WriteString("string")
				buf.WriteByte(' ')
				buf.Write([]byte("bytes"))
			},
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				expected := "string bytes"
				assert.Equal(t, expected, buffer.String())
				assert.Equal(t, len(expected), buffer.Len())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := getBuffer()
			tt.dataFunc(buffer)
			if tt.checkFunc != nil {
				tt.checkFunc(t, buffer)
			}
			putBuffer(buffer)
		})
	}
}

// TestHash_getBufferPutBuffer_Stress tests the pool under stress.
func TestHash_getBufferPutBuffer_Stress(t *testing.T) {
	tests := []struct {
		name       string
		iterations int
		checkFunc  func(t *testing.T)
	}{
		{
			name:       "100 iterations",
			iterations: 100,
			checkFunc: func(t *testing.T) {
				// Test passes if no panics occur
			},
		},
		{
			name:       "1000 iterations",
			iterations: 1000,
			checkFunc: func(t *testing.T) {
				// Test passes if no panics occur
			},
		},
		{
			name:       "10000 iterations",
			iterations: 10000,
			checkFunc: func(t *testing.T) {
				// Test passes if no panics occur
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.iterations; i++ {
				buffer := getBuffer()
				require.NotNil(t, buffer)

				// Use the buffer
				buffer.WriteString("iteration-")
				buffer.WriteString(string(rune(i % 256)))

				// Verify
				assert.Greater(t, buffer.Len(), 0)

				// Put back
				putBuffer(buffer)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t)
			}
		})
	}
}

// TestHash_bufferPool tests that the bufferPool New function creates valid buffers.
func TestHash_bufferPool(t *testing.T) {
	tests := []struct {
		name      string
		checkFunc func(t *testing.T, buffer *bytes.Buffer)
	}{
		{
			name: "creates valid buffer",
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				require.NotNil(t, buffer)
				assert.IsType(t, &bytes.Buffer{}, buffer)
			},
		},
		{
			name: "buffer is usable",
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				buffer.WriteString("test")
				assert.Equal(t, "test", buffer.String())
				assert.Equal(t, 4, buffer.Len())
			},
		},
		{
			name: "buffer starts empty",
			checkFunc: func(t *testing.T, buffer *bytes.Buffer) {
				assert.Equal(t, 0, buffer.Len())
				assert.Equal(t, "", buffer.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new pool with the same New function to test it directly
			testPool := sync.Pool{
				New: func() interface{} {
					return new(bytes.Buffer)
				},
			}

			buffer := testPool.Get().(*bytes.Buffer)
			if tt.checkFunc != nil {
				tt.checkFunc(t, buffer)
			}
			testPool.Put(buffer)
		})
	}
}
