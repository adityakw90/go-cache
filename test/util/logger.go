package testutil

import (
	"fmt"
	"testing"

	"github.com/adityakw90/go-cache/internal/adapter"
)

type UnitTestLogger struct{}

func (l *UnitTestLogger) Info(msg string, fields map[string]interface{}) {
	fmt.Println(msg, fields)
}

func (l *UnitTestLogger) Error(msg string, fields map[string]interface{}) {
	fmt.Println(msg, fields)
}

func (l *UnitTestLogger) Debug(msg string, fields map[string]interface{}) {
	fmt.Println(msg, fields)
}

func (l *UnitTestLogger) WithSpanContext(spanContext adapter.SpanContext) adapter.Logger {
	return l
}

func CreateUnitTestLogger(t *testing.T) *UnitTestLogger {
	t.Helper()

	return &UnitTestLogger{}
}
