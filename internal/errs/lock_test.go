package errs

import "testing"

func TestErrs_Lock_SentinelErrors(t *testing.T) {
	if ErrLockAcquireFailed == nil {
		t.Error("ErrLockAcquireFailed should not be nil")
	}

	if ErrLockReleaseUnlocked == nil {
		t.Error("ErrLockReleaseUnlocked should not be nil")
	}

	if ErrLockReleaseForbidden == nil {
		t.Error("ErrLockReleaseForbidden should not be nil")
	}

	// Test that sentinel errors are comparable
	err1 := ErrLockAcquireFailed
	err2 := ErrLockAcquireFailed
	if err1 != err2 {
		t.Error("sentinel errors should be comparable")
	}
}
