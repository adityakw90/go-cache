package cache

import "github.com/adityakw90/go-cache/internal/errs"

// re-export errors
type InvalidConfigError = errs.InvalidConfigError

var (
	ErrLockAcquireFailed    = errs.ErrLockAcquireFailed
	ErrLockReleaseUnlocked  = errs.ErrLockReleaseUnlocked
	ErrLockReleaseForbidden = errs.ErrLockReleaseForbidden
)
