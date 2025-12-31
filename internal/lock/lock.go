package lock

// LockData represents a distributed lock.
type LockData struct {
	Key      string
	Token    string
	Acquired bool
	Released bool
	Error    error
}
