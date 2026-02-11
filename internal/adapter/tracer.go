package adapter

// No Operation Tracer implementation.
type NoOpSpan struct{}

// End does nothing.
func (n *NoOpSpan) End() {}
