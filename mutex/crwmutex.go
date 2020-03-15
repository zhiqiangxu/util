package mutex

import (
	"context"

	"golang.org/x/sync/semaphore"
)

const rwmutexMaxReaders = 1 << 30

// CRWMutex implements a cancelable rwmutex (in fact also a try-able rwmutex)
type CRWMutex struct {
	sema *semaphore.Weighted
}

// NewCRWMutex is ctor for CRWMutex
func NewCRWMutex() *CRWMutex {
	rw := &CRWMutex{}
	rw.Init()
	return rw
}

// Init for embeded usage
func (rw *CRWMutex) Init() {
	rw.sema = semaphore.NewWeighted(rwmutexMaxReaders)
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

// TryLock returns true if lock acquired
func (rw *CRWMutex) TryLock() bool {
	return rw.sema.TryAcquire(rwmutexMaxReaders)
}

// TryRLock returns true if rlock acquired
func (rw *CRWMutex) TryRLock() bool {
	return rw.sema.TryAcquire(1)
}
