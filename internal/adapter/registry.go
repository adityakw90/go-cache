package adapter

// NewNoOpTracer creates a new no-op tracer.
func NewNoOpTracer() Tracer {
	return &NoOpTracer{}
}

// NewNoOpLogger creates a new no-op logger.
func NewNoOpLogger() Logger {
	return &NoOpLogger{}
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
