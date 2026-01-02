package cache

import "errors"

var (
	// Operation Get
	ErrGetCacheMiss              = errors.New("cache miss")
	ErrGetCacheFailed            = errors.New("failed to get cache")
	ErrGetCacheDeserializeFailed = errors.New("failed to deserialize cache data")

	// Operation Set
	ErrSetCacheSerializeFailed = errors.New("failed to serialize cache data")
	ErrSetCacheFailed          = errors.New("failed to set cache data")
)
