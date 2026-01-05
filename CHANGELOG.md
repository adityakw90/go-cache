# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> **⚠️ Pre-1.0.0 Notice**: Versions before 1.0.0 are considered unstable. **Backward compatibility is not guaranteed** between minor versions (0.x → 0.y). Breaking changes may be introduced without major version bumps. Once version 1.0.0 is released, this project will follow semantic versioning strictly, and breaking changes will only occur in major version updates.

## [Unreleased]

## [0.2.0] - 2026-01-05

### Changed
- Upgraded Redis client dependency from `github.com/go-redis/redis/v8` to `github.com/redis/go-redis/v9`
- Updated Go version requirement from 1.22+ to 1.24+ to match go.mod
- Updated documentation to reflect Redis client v9 import path
- Replaced version history section with CHANGELOG.md following Keep a Changelog format

## [0.1.0] - Initial Release

### Added
- Initial release of go-cache library
- Redis-based caching with version-based invalidation (using `github.com/go-redis/redis/v8`)
- Distributed locking mechanism to prevent cache stampede
- Cache decorator pattern for automatic function result caching
- Custom key generation with flexible cache key strategies
- Optional observability integration (Tracer/Logger interfaces)
- Dynamic TTL support (fixed and function-based)
- Concurrency control with built-in semaphore
- Thread-safe concurrent access
- Pipeline support for batch Redis operations
- Functional options pattern for configuration
- Comprehensive test coverage

### Features
- **Cache Core**: Main cache implementation with Redis operations (`Get`, `Set`, `CleanCache`)
- **Version Manager**: Version-based cache invalidation without deleting individual keys
- **Lock Manager**: Distributed locking to prevent cache stampede
- **Decorator Pattern**: Automatic function result caching with `Cached()` method
- **Hash Generator**: Consistent hash-based cache keys from function arguments
- **Serialization**: Gob encoding for complex types
- **Key Generator**: Flexible key generation with support for custom generators
- **Adapters**: Pluggable tracer, logger, and semaphore interfaces

### Documentation
- Comprehensive README with examples
- GoDoc comments on all exported functions
- API reference documentation
- Usage examples for common scenarios
- Architecture diagrams

---
