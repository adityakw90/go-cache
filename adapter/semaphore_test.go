package adapter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdapter_NewSemaphore(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		checkFunc func(t *testing.T, sem Semaphore)
	}{
		{
			name: "valid size",
			size: 5,
			checkFunc: func(t *testing.T, sem Semaphore) {
				assert.NotNil(t, sem)
				assert.Equal(t, 5, sem.Size())
			},
		},
		{
			name: "zero size defaults to 1",
			size: 0,
			checkFunc: func(t *testing.T, sem Semaphore) {
				assert.NotNil(t, sem)
				assert.Equal(t, 1, sem.Size())
			},
		},
		{
			name: "negative size defaults to 1",
			size: -1,
			checkFunc: func(t *testing.T, sem Semaphore) {
				assert.NotNil(t, sem)
				assert.Equal(t, 1, sem.Size())
			},
		},
		{
			name: "large size",
			size: 1000,
			checkFunc: func(t *testing.T, sem Semaphore) {
				assert.NotNil(t, sem)
				assert.Equal(t, 1000, sem.Size())
			},
		},
		{
			name: "size of 1",
			size: 1,
			checkFunc: func(t *testing.T, sem Semaphore) {
				assert.NotNil(t, sem)
				assert.Equal(t, 1, sem.Size())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sem := NewSemaphore(tt.size)
			if tt.checkFunc != nil {
				tt.checkFunc(t, sem)
			}
		})
	}
}
