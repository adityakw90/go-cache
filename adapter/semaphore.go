package adapter

import internaladapter "github.com/adityakw90/go-cache/internal/adapter"

// Semaphore interface for concurrency control.
type Semaphore interface {
	Size() int // Size returns the size of the semaphore.
	Acquire()  // Acquire acquires a semaphore permit, blocking if necessary.
	Release()  // Release releases a semaphore permit.
}

// NewSemaphore creates a new semaphore with the given size.
func NewSemaphore(size int) Semaphore {
	return internaladapter.NewSemaphore(size)
}
