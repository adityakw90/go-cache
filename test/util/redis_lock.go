package testutil

import (
	"fmt"
	"os"
	"syscall"
)

func lockRedisDB(db int) (*os.File, error) {
	path := fmt.Sprintf("/tmp/unit-test-redis-db-%d.lock", db)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, err
	}

	return f, nil
}

func unlockRedisDB(f *os.File) {
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	_ = f.Close()
}
