package lock

import (
	"errors"
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
				Error:    nil,
			},
			validate: func(t *testing.T, lock *LockData) {
				assert.Equal(t, "test:key", lock.Key)
				assert.Equal(t, "test-token", lock.Token)
				assert.True(t, lock.Acquired)
				assert.False(t, lock.Released)
				assert.NoError(t, lock.Error)
			},
		},
		{
			name: "acquired_and_released",
			lock: &LockData{
				Key:      "test:key:2",
				Token:    "token-123",
				Acquired: true,
				Released: true,
				Error:    nil,
			},
			validate: func(t *testing.T, lock *LockData) {
				assert.Equal(t, "test:key:2", lock.Key)
				assert.Equal(t, "token-123", lock.Token)
				assert.True(t, lock.Acquired)
				assert.True(t, lock.Released)
				assert.NoError(t, lock.Error)
			},
		},
		{
			name: "not_acquired",
			lock: &LockData{
				Key:      "test:key:3",
				Token:    "token-456",
				Acquired: false,
				Released: false,
				Error:    nil,
			},
			validate: func(t *testing.T, lock *LockData) {
				assert.Equal(t, "test:key:3", lock.Key)
				assert.Equal(t, "token-456", lock.Token)
				assert.False(t, lock.Acquired)
				assert.False(t, lock.Released)
				assert.NoError(t, lock.Error)
			},
		},
		{
			name: "with_error",
			lock: &LockData{
				Key:      "test:key:4",
				Token:    "token-789",
				Acquired: false,
				Released: false,
				Error:    errors.New("test error"),
			},
			validate: func(t *testing.T, lock *LockData) {
				assert.Equal(t, "test:key:4", lock.Key)
				assert.Equal(t, "token-789", lock.Token)
				assert.False(t, lock.Acquired)
				assert.False(t, lock.Released)
				assert.Error(t, lock.Error)
				assert.Equal(t, "test error", lock.Error.Error())
			},
		},
		{
			name: "empty_key_and_token",
			lock: &LockData{
				Key:      "",
				Token:    "",
				Acquired: false,
				Released: false,
				Error:    nil,
			},
			validate: func(t *testing.T, lock *LockData) {
				assert.Empty(t, lock.Key)
				assert.Empty(t, lock.Token)
				assert.False(t, lock.Acquired)
				assert.False(t, lock.Released)
				assert.NoError(t, lock.Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.lock)
		})
	}
}
