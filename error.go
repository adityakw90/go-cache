package cache

import (
	"github.com/adityakw90/go-cache/internal/cache"
	"github.com/adityakw90/go-cache/internal/errs"
	"github.com/adityakw90/go-cache/internal/key"
	"github.com/adityakw90/go-cache/internal/lock"
	"github.com/adityakw90/go-cache/internal/serialize"
)

// re-export errors
type InvalidConfigError = errs.InvalidConfigError

// NewInvalidConfigError creates a new InvalidConfigError error.
func NewInvalidConfigError(field string, message string) error {
	return errs.NewInvalidConfigError(field, message)
}

var (
	// cache operation errors
	ErrGetCacheMiss = cache.ErrGetCacheMiss

	// lock errors
	ErrLockAcquireFailed    = lock.ErrLockAcquireFailed
	ErrLockReleaseUnlocked  = lock.ErrLockReleaseUnlocked
	ErrLockReleaseForbidden = lock.ErrLockReleaseForbidden

	// serialize errors
	ErrSerializeNilValue    = serialize.ErrSerializeNilValue
	ErrDeserializeEmptyData = serialize.ErrDeserializeEmptyData
	ErrDeserializeResultNil = serialize.ErrDeserializeResultNil

	// custom key function errors
	ErrNameRequired                        = key.ErrNameRequired
	ErrCallableRequired                    = key.ErrCallableRequired
	ErrParamsRequired                      = key.ErrParamsRequired
	ErrKeyGeneratorParamsPrefixRequired    = key.ErrKeyGeneratorParamsPrefixRequired
	ErrKeyGeneratorParamsKeyRequired       = key.ErrKeyGeneratorParamsKeyRequired
	ErrKeyGeneratorParamsNamespaceRequired = key.ErrKeyGeneratorParamsNamespaceRequired
	ErrKeyGeneratorParamsVersionRequired   = key.ErrKeyGeneratorParamsVersionRequired
)
