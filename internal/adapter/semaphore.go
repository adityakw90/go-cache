package adapter

type semaphore struct {
	sem chan struct{}
}

// Acquire acquires a semaphore permit, blocking if necessary.
func (s *semaphore) Acquire() {
	s.sem <- struct{}{}
}

// Release releases a semaphore permit.
func (s *semaphore) Release() {
	<-s.sem
}
