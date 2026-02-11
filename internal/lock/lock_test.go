package lock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLock_LockData_Fields(t *testing.T) {
	tests := []struct {
		name     string
		lock     *LockData
		validate func(*testing.T, *LockData)
	}{
		{
			name: "all_fields_set",
			lock: &LockData{
				Key:      "test:key",
				Token:    "test-token",
				Acquired: true,
				Released: false,
			},
			validate: func(t *testing.T, lock *LockData) {
				assert.Equal(t, "test:key", lock.Key)
				assert.Equal(t, "test-token", lock.Token)
				assert.True(t, lock.Acquired)
				assert.False(t, lock.Released)
			},
		},
		{
			name: "acquired_and_released",
			lock: &LockData{
				Key:      "test:key:2",
				Token:    "token-123",
				Acquired: true,
				Released: true,
			},
			validate: func(t *testing.T, lock *LockData) {
				assert.Equal(t, "test:key:2", lock.Key)
				assert.Equal(t, "token-123", lock.Token)
				assert.True(t, lock.Acquired)
				assert.True(t, lock.Released)
			},
		},
		{
			name: "not_acquired",
			lock: &LockData{
				Key:      "test:key:3",
				Token:    "token-456",
				Acquired: false,
				Released: false,
			},
			validate: func(t *testing.T, lock *LockData) {
				assert.Equal(t, "test:key:3", lock.Key)
				assert.Equal(t, "token-456", lock.Token)
				assert.False(t, lock.Acquired)
				assert.False(t, lock.Released)
			},
		},
		{
			name: "empty_key_and_token",
			lock: &LockData{
				Key:      "",
				Token:    "",
				Acquired: false,
				Released: false,
			},
			validate: func(t *testing.T, lock *LockData) {
				assert.Empty(t, lock.Key)
				assert.Empty(t, lock.Token)
				assert.False(t, lock.Acquired)
				assert.False(t, lock.Released)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.lock)
		})
	}
}
