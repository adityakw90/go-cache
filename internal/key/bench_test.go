// Package key provides benchmarks comparing string formatting methods.
//
// These benchmarks compare the performance of fmt.Sprintf vs string concatenation
// for building cache keys in the format: "prefix:namespace:vversion-key.gob"
//
// Expected Results:
//   - String concatenation is typically 2-5x faster than fmt.Sprintf for simple cases
//   - fmt.Sprintf has more overhead due to format parsing and argument handling
//   - For compile-time constant strings, concatenation is optimized by the compiler
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
