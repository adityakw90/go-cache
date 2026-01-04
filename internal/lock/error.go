package lock

import "errors"

// Sentinel errors for lock operations.
var (
	ErrLockAcquireFailed    = errors.New("failed to acquire lock")                       // lock acquisition failed.
	ErrLockReleaseUnlocked  = errors.New("cannot release an unlocked lock")              // release a lock that doesn't exist.
	ErrLockReleaseForbidden = errors.New("cannot release a lock that's no longer owned") // release a lock owned by someone else.
)
