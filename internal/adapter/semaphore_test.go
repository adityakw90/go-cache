package adapter

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapter_NewSemaphore(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		checkFunc func(t *testing.T, sem *semaphore)
	}{
		{
			name: "valid size",
			size: 5,
			checkFunc: func(t *testing.T, sem *semaphore) {
				assert.NotNil(t, sem)
				assert.Equal(t, 5, sem.Size())
			},
		},
		{
			name: "zero size defaults to 1",
			size: 0,
			checkFunc: func(t *testing.T, sem *semaphore) {
				assert.NotNil(t, sem)
				assert.Equal(t, 1, sem.Size())
			},
		},
		{
			name: "negative size defaults to 1",
			size: -1,
			checkFunc: func(t *testing.T, sem *semaphore) {
				assert.NotNil(t, sem)
				assert.Equal(t, 1, sem.Size())
			},
		},
		{
			name: "large size",
			size: 1000,
			checkFunc: func(t *testing.T, sem *semaphore) {
				assert.NotNil(t, sem)
				assert.Equal(t, 1000, sem.Size())
			},
		},
		{
			name: "size of 1",
			size: 1,
			checkFunc: func(t *testing.T, sem *semaphore) {
				assert.NotNil(t, sem)
				assert.Equal(t, 1, sem.Size())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sem := NewSemaphore(tt.size)
			if tt.checkFunc != nil {
				tt.checkFunc(t, sem)
			}
		})
	}
}

func TestAdapter_Semaphore_AcquireRelease(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		checkFunc func(t *testing.T, sem *semaphore)
	}{
		{
			name: "acquire and release single permit",
			size: 1,
			checkFunc: func(t *testing.T, sem *semaphore) {
				// Should not block when capacity available
				sem.Acquire()
				// Release should work
				sem.Release()
			},
		},
		{
			name: "acquire and release multiple permits",
			size: 3,
			checkFunc: func(t *testing.T, sem *semaphore) {
				// Acquire all permits
				sem.Acquire()
				sem.Acquire()
				sem.Acquire()

				// Release all permits
				sem.Release()
				sem.Release()
				sem.Release()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sem := NewSemaphore(tt.size)
			if tt.checkFunc != nil {
				tt.checkFunc(t, sem)
			}
		})
	}
}

func TestAdapter_Semaphore_BlockingBehavior(t *testing.T) {
	sem := NewSemaphore(2)

	// Acquire all permits
	sem.Acquire()
	sem.Acquire()

	// Try to acquire third permit - should block
	done := make(chan bool, 1)
	go func() {
		sem.Acquire()
		done <- true
	}()

	// Wait a bit to ensure goroutine is blocked
	time.Sleep(50 * time.Millisecond)
	select {
	case <-done:
		t.Fatal("semaphore should have blocked")
	default:
		// Expected - goroutine is blocked
	}

	// Release one permit
	sem.Release()

	// Now the goroutine should complete
	select {
	case <-done:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("semaphore should have unblocked")
	}

	// Clean up
	sem.Release()
}

func TestAdapter_Semaphore_Concurrent(t *testing.T) {
	sem := NewSemaphore(2)
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	// Try to acquire more permits than available
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			sem.Acquire()
			done <- true
			// Hold permit briefly
			time.Sleep(10 * time.Millisecond)
			sem.Release()
		}(i)
	}

	// Wait for all goroutines to complete
	timeout := time.After(5 * time.Second)
	completed := 0
	for completed < numGoroutines {
		select {
		case <-done:
			completed++
		case <-timeout:
			t.Fatalf("timeout waiting for goroutines, completed: %d/%d", completed, numGoroutines)
		}
	}
}

func TestAdapter_Semaphore_ConcurrentAcquireRelease(t *testing.T) {
	sem := NewSemaphore(5)
	const numGoroutines = 20
	var wg sync.WaitGroup

	// Concurrent acquire and release
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()

			// Simulate some work
			time.Sleep(10 * time.Millisecond)
		}(i)
	}

	wg.Wait()
}

func TestAdapter_Semaphore_MultipleAcquireRelease(t *testing.T) {
	sem := NewSemaphore(3)

	// Acquire multiple times
	sem.Acquire()
	sem.Acquire()
	sem.Acquire()

	// Release multiple times
	sem.Release()
	sem.Release()
	sem.Release()

	// Should be able to acquire again
	sem.Acquire()
	sem.Release()
}

func TestAdapter_Semaphore_ZeroSizeDefaultsToOne(t *testing.T) {
	sem := NewSemaphore(0)

	// Should only allow one permit at a time
	done := make(chan bool, 1)
	sem.Acquire()

	go func() {
		sem.Acquire()
		done <- true
	}()

	// Wait a bit to ensure goroutine is blocked
	time.Sleep(50 * time.Millisecond)
	select {
	case <-done:
		t.Fatal("semaphore should have blocked (size should default to 1)")
	default:
		// Expected
	}

	// Release should unblock
	sem.Release()

	select {
	case <-done:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("semaphore should have unblocked")
	}
}

func TestAdapter_Semaphore_NegativeSizeDefaultsToOne(t *testing.T) {
	sem := NewSemaphore(-5)

	// Should only allow one permit at a time
	done := make(chan bool, 1)
	sem.Acquire()

	go func() {
		sem.Acquire()
		done <- true
	}()

	// Wait a bit to ensure goroutine is blocked
	time.Sleep(50 * time.Millisecond)
	select {
	case <-done:
		t.Fatal("semaphore should have blocked (size should default to 1)")
	default:
		// Expected
	}

	// Release should unblock
	sem.Release()

	select {
	case <-done:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("semaphore should have unblocked")
	}
}

func TestAdapter_Semaphore_StressTest(t *testing.T) {
	sem := NewSemaphore(10)
	const iterations = 1000
	var wg sync.WaitGroup

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()
			// Minimal work
		}()
	}

	wg.Wait()
}

func TestAdapter_Semaphore_Ordering(t *testing.T) {
	sem := NewSemaphore(1)
	const numGoroutines = 5
	order := make(chan int, numGoroutines)
	var wg sync.WaitGroup

	// All goroutines try to acquire at once
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()
			order <- id
		}(i)
	}

	wg.Wait()
	close(order)

	// Verify all goroutines completed
	ids := make([]int, 0, numGoroutines)
	for id := range order {
		ids = append(ids, id)
	}

	require.Len(t, ids, numGoroutines, "all goroutines should have completed")
}
