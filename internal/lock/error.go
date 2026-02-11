package lock

import "errors"

// Sentinel errors for lock operations.
var (
	// ErrLockAcquireFailed is returned when the lock cannot be acquired.
	ErrLockAcquireFailed = errors.New("failed to acquire lock")
	// ErrLockReleaseUnlocked is returned when the lock does not exist.
	ErrLockReleaseUnlocked = errors.New("lock does not exist")
	// ErrLockReleaseForbidden is returned when the lock exists but is owned by someone else.
	ErrLockReleaseForbidden = errors.New("lock is owned by another process")
)
