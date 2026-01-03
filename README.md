# Go Cache Library

A production-ready, framework-agnostic Go caching library with Redis backend, versioning, distributed locking, and decorator pattern support.

## Features

- **Redis-Based Caching**: Fast, distributed caching using Redis
- **Version-Based Invalidation**: Efficient cache invalidation without deleting individual keys
- **Distributed Locking**: Prevents cache stampede with Redis-based locks
- **Decorator Pattern**: Automatic function result caching with clean API
- **Custom Key Generation**: Flexible cache key strategies
- **Optional Observability**: Interface-based design for Tracer/Logger (no forced dependencies)
- **Dynamic TTL**: Support for both fixed and function-based TTL
- **Thread-Safe**: Concurrent access with proper locking

## Installation

```bash
go get github.com/adityakw90/go-cache
```

## Quick Start

```go
package main

import (
    "context"
    "time"
    
    "github.com/adityakw90/go-cache"
    "github.com/go-redis/redis/v8"
)

func main() {
    // Create Redis client
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // Create cache instance
    c, err := cache.NewCache(
        redisClient,
        cache.WithKeyPrefix("myapp"),
        cache.WithExpireDefault(5 * time.Minute),
    )
    if err != nil {
        panic(err)
    }
    
    // Define a function to cache
    getUserFromDB := func(ctx context.Context, args ...interface{}) (interface{}, error) {
        userID := args[0].(string)
        // ... fetch from database
        return &User{ID: userID, Name: "John"}, nil
    }
    
    // Create custom key function
    customKeyFunc, err := cache.NewCustomKeyFunction(
        "getUser",
        func(args ...interface{}) string {
            return "user:" + args[0].(string)
        },
        []string{"userID"},
    )
    if err != nil {
        panic(err)
    }
    
    // Create cached function
    cachedGetUser := c.Cached(
        "getUser",
        10 * time.Minute,
        true, // versioning enabled
        "user",
    )(
        getUserFromDB,
        customKeyFunc,
    )
    
    // Use the cached function
    ctx := context.Background()
    var user *User
    result, err := cachedGetUser(&user, ctx, "user-123")
    if err != nil {
        panic(err)
    }
    
    // Cache hit on subsequent calls
    var user2 *User
    result, err = cachedGetUser(&user2, ctx, "user-123")
    // Returns cached value, doesn't call getUserFromDB
}
```

## Configuration Options

The library uses functional options for flexible configuration:

```go
cache, err := cache.NewCache(
    redisClient,
    cache.WithKeyPrefix("myapp"),              // Cache key prefix
    cache.WithExpireDefault(5 * time.Minute),  // Default TTL
    cache.WithVersionExpire(30 * 24 * time.Hour), // Version key TTL
    cache.WithLockDuration(time.Minute),       // Lock timeout
    cache.WithLockInterval(100 * time.Millisecond), // Lock retry interval
    cache.WithSemaphoreSize(10),               // Concurrency limit
    cache.WithTracer(myTracer),                // Optional tracer
    cache.WithLogger(myLogger),                // Optional logger
)
```

## Cache Invalidation

Invalidate all cache entries for a namespace by incrementing the version:

```go
err := cache.InvalidateVersion(ctx, "getUser", "user")
// All cached getUser entries become invalid
```

## Observability Integration

The library provides interfaces for optional observability integration:

```go
// Tracer interface
type Tracer interface {
    StartSpan(ctx context.Context, name string) (context.Context, Span)
    NewSpanFromSpan(ctx context.Context, name string, parent Span) (context.Context, Span)
}

// Logger interface
type Logger interface {
    Info(msg string, fields map[string]interface{})
    Error(msg string, fields map[string]interface{})
    Debug(msg string, fields map[string]interface{})
    WithSpanContext(spanContext SpanContext) Logger
}
```

If no tracer/logger is provided, the library uses no-op implementations.

## Dynamic TTL

Support for function-based TTL based on result and arguments:

```go
ttlFunc := func(result interface{}, args ...interface{}) time.Duration {
    if user, ok := result.(*User); ok {
        if user.IsPremium {
            return 1 * time.Hour
        }
    }
    return 10 * time.Minute
}

cachedFunc := cache.Cached(
    "getUser",
    ttlFunc, // Dynamic TTL function
    true,
    "user",
)(getUserFromDB, nil)
```

## Architecture

The library follows an interface-based design:

- **Cache Core**: Main cache implementation with Redis operations
- **Version Manager**: Version-based cache invalidation
- **Lock Manager**: Distributed locking to prevent cache stampede
- **Decorator Pattern**: Automatic function result caching
- **Hash Generator**: Consistent hash-based cache keys
- **Serialization**: Gob encoding for complex types

## Testing

Run tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -race -covermode=atomic -coverprofile=coverage.txt ./...
```

## Requirements

- Go 1.22+
- Redis server
- `github.com/go-redis/redis/v8`

## License

[To be determined]

## Contributing

[To be added]

---

*Built with SDD 3.0 - Spec-Driven Development*
