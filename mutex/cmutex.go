package mutex

import (
	"context"

	"golang.org/x/sync/semaphore"
)

// CMutex implements a cancelable mutex
type CMutex struct {
	sema *semaphore.Weighted
}

// NewCMutex is ctor for CMutex
func NewCMutex() *CMutex {
	return &CMutex{sema: semaphore.NewWeighted(1)}
}

// Lock with context
func (m *CMutex) Lock(ctx context.Context) (err error) {
	err = m.sema.Acquire(ctx, 1)
	return
}

// Unlock should only be called after a successful Lock
func (m *CMutex) Unlock() {
	m.sema.Release(1)
}
