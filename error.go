package cache

import (
	"github.com/adityakw90/go-cache/internal/errs"
	"github.com/adityakw90/go-cache/internal/lock"
)

// re-export errors
type InvalidConfigError = errs.InvalidConfigError

var (
	// lock errors
	ErrLockAcquireFailed    = lock.ErrLockAcquireFailed
	ErrLockReleaseUnlocked  = lock.ErrLockReleaseUnlocked
	ErrLockReleaseForbidden = lock.ErrLockReleaseForbidden

	// serialize errors
	ErrSerializeNilValue    = errs.ErrSerializeNilValue
	ErrDeserializeEmptyData = errs.ErrDeserializeEmptyData
	ErrDeserializeResultNil = errs.ErrDeserializeResultNil
)
