package errs

import (
	"errors"
	"testing"
)

func TestErrs_Wrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, "context")

	if wrappedErr == nil {
		t.Fatal("expected non-nil error")
	}

	if !errors.Is(wrappedErr, originalErr) {
		t.Error("wrapped error should wrap original error")
	}
}

func TestErrs_Wrapf(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrapf(originalErr, "context: %s", "test")

	if wrappedErr == nil {
		t.Fatal("expected non-nil error")
	}

	if !errors.Is(wrappedErr, originalErr) {
		t.Error("wrapped error should wrap original error")
	}
}

func TestErrs_WrapNil(t *testing.T) {
	if Wrap(nil, "context") != nil {
		t.Error("wrapping nil should return nil")
	}

	if Wrapf(nil, "context: %s", "test") != nil {
		t.Error("wrapping nil should return nil")
	}
}
