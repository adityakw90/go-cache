package cache

import "github.com/adityakw90/go-cache/internal/errs"

// re-export errors
type InvalidConfigError = errs.InvalidConfigError

var (
	// lock errors
	ErrLockAcquireFailed    = errs.ErrLockAcquireFailed
	ErrLockReleaseUnlocked  = errs.ErrLockReleaseUnlocked
	ErrLockReleaseForbidden = errs.ErrLockReleaseForbidden

	// serialize errors
	ErrSerializeNilValue    = errs.ErrSerializeNilValue
	ErrDeserializeEmptyData = errs.ErrDeserializeEmptyData
	ErrDeserializeResultNil = errs.ErrDeserializeResultNil
)
