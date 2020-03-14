package mutex

import (
	"context"

	"golang.org/x/sync/semaphore"
)

const rwmutexMaxReaders = 1 << 30

// CRWMutex implements a cancelable rwmutex
type CRWMutex struct {
	sema *semaphore.Weighted
}

// NewCRWMutex is ctor for CRWMutex
func NewCRWMutex() *CRWMutex {
	return &CRWMutex{sema: semaphore.NewWeighted(rwmutexMaxReaders)}
}

// Lock with context
func (rw *CRWMutex) Lock(ctx context.Context) (err error) {
	err = rw.sema.Acquire(ctx, rwmutexMaxReaders)
	return
}

// Unlock should only be called after a successful Lock
func (rw *CRWMutex) Unlock() {
	rw.sema.Release(rwmutexMaxReaders)
	return
}

// RLock with context
func (rw *CRWMutex) RLock(ctx context.Context) (err error) {
	err = rw.sema.Acquire(ctx, 1)
	return
}

// RUnlock should only be called after a successful RLock
func (rw *CRWMutex) RUnlock() {
	rw.sema.Release(1)
}
