package cache

import "errors"

var (
	// Operation Get
	ErrGetCacheMiss = errors.New("cache miss")
)
