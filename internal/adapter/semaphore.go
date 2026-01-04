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
