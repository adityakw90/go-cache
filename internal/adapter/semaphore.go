package adapter

type semaphore struct {
	sem chan struct{}
}

// Size returns the size of the semaphore.
func (s *semaphore) Size() int {
	return cap(s.sem)
}

// Acquire acquires a semaphore permit, blocking if necessary.
func (s *semaphore) Acquire() {
	s.sem <- struct{}{}
}

// Release releases a semaphore permit.
func (s *semaphore) Release() {
	<-s.sem
}

// NewSemaphore creates a new semaphore with the given size.
func NewSemaphore(size int) *semaphore {
	if size <= 0 {
		size = 1
	}
	return &semaphore{
		sem: make(chan struct{}, size),
	}
}
