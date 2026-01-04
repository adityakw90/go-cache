package errs

import (
	"fmt"
)

// interface for error
type InvalidConfigError interface {
	Error() string
	Field() string
	Message() string
}

// invalidConfigError represents an invalid configuration error.
type invalidConfigError struct {
	field   string
	message string
}

func (e *invalidConfigError) Error() string {
	return fmt.Sprintf("invalid config %s : %s", e.field, e.message)
}

func (e *invalidConfigError) Field() string {
	return e.field
}

func (e *invalidConfigError) Message() string {
	return e.message
}

// NewInvalidConfigError creates a new InvalidConfigError error.
func NewInvalidConfigError(field string, message string) error {
	return &invalidConfigError{field: field, message: message}
}
