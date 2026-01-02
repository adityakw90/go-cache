// Package key provides benchmarks comparing string formatting methods.
//
// These benchmarks compare the performance of fmt.Sprintf vs string concatenation vs strings.Builder
// for building cache keys in the format: "prefix:namespace:vversion-key.gob"
//
// Expected Results:
//   - String concatenation is typically 2-5x faster than fmt.Sprintf for simple cases
//   - fmt.Sprintf has more overhead due to format parsing and argument handling
//   - For compile-time constant strings, concatenation is optimized by the compiler
//   - strings.Builder is efficient for dynamic string building with multiple operations
//
// Run benchmarks with:
//
//	go test -bench=. -benchmem ./internal/key
//
// Example output interpretation:
//   - Lower ns/op (nanoseconds per operation) = faster
//   - Lower B/op (bytes per operation) = less memory allocation
//   - Lower allocs/op = fewer heap allocations
package key

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkKey_printFormatterSprintf benchmarks fmt.Sprintf for string formatting.
//
// This method uses format string parsing and is more flexible but has higher overhead.
// Use this when you need dynamic formatting or complex string templates.
func BenchmarkKey_printFormatterSprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%s:%s:v%s-%s.gob", "prefix", "namespace", "version", "key")
	}
}

// BenchmarkKey_stringConcatenation benchmarks string concatenation using the + operator.
//
// This method is typically faster for simple string building with known values.
// The Go compiler can optimize compile-time constant concatenations.
// Use this when you have a fixed number of string parts to combine.
func BenchmarkKey_stringConcatenation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = "prefix" + ":" + "namespace" + ":v" + "version" + "-" + "key" + ".gob"
	}
}

// BenchmarkKey_stringBuilder benchmarks strings.Builder for string building.
//
// This method is efficient for building strings dynamically, especially when the number
// of concatenations is unknown at compile time or when building strings in loops.
// strings.Builder minimizes allocations by growing the underlying buffer as needed.
// Use this when you need to build strings dynamically or have many concatenations.
func BenchmarkKey_stringBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var sb strings.Builder
		sb.WriteString("prefix")
		sb.WriteString(":")
		sb.WriteString("namespace")
		sb.WriteString(":v")
		sb.WriteString("version")
		sb.WriteString("-")
		sb.WriteString("key")
		sb.WriteString(".gob")
		_ = sb.String()
	}
}

// BenchmarkKeyGenerator benchmarks the KeyGenerator function.
//
// This benchmark measures the performance of generating simple cache keys
// in the format: {prefix}:{key}.gob
func BenchmarkKeyGenerator(b *testing.B) {
	data := map[string]string{
		"prefix": "cache",
		"key":    "user:123",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = KeyGenerator(data)
	}
}

// BenchmarkKeyVersionGenerator benchmarks the KeyVersionGenerator function.
//
// This benchmark measures the performance of generating versioned cache keys
// in the format: {prefix}:{namespace}:v{version}-{key}.gob
func BenchmarkKeyVersionGenerator(b *testing.B) {
	data := map[string]string{
		"prefix":    "cache",
		"namespace": "users",
		"version":   "1.0.0",
		"key":       "user:123",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = KeyVersionGenerator(data)
	}
}

// BenchmarkVersionGenerator benchmarks the VersionGenerator function.
//
// This benchmark measures the performance of generating version keys
// in the format: {prefix}:{namespace}:version
func BenchmarkVersionGenerator(b *testing.B) {
	data := map[string]string{
		"prefix":    "cache",
		"namespace": "users",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = VersionGenerator(data)
	}
}

// BenchmarkLockGenerator benchmarks the LockGenerator function.
//
// This benchmark measures the performance of generating lock keys
// in the format: {prefix}:{namespace}:lock
func BenchmarkLockGenerator(b *testing.B) {
	data := map[string]string{
		"prefix":    "cache",
		"namespace": "users",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = LockGenerator(data)
	}
}
