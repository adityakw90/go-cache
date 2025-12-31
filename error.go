package cache

import (
	"github.com/adityakw90/go-cache/internal/errs"
	"github.com/adityakw90/go-cache/internal/lock"
	"github.com/adityakw90/go-cache/internal/serialize"
)

// re-export errors
type InvalidConfigError = errs.InvalidConfigError

var (
	// lock errors
	ErrLockAcquireFailed    = lock.ErrLockAcquireFailed
	ErrLockReleaseUnlocked  = lock.ErrLockReleaseUnlocked
	ErrLockReleaseForbidden = lock.ErrLockReleaseForbidden

	// serialize errors
	ErrSerializeNilValue    = serialize.ErrSerializeNilValue
	ErrDeserializeEmptyData = serialize.ErrDeserializeEmptyData
	ErrDeserializeResultNil = serialize.ErrDeserializeResultNil
)
