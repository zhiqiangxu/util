package deadlock

import (
	"context"
	"unsafe"
)

const rwmutexMaxReaders = 1 << 30

// RWMutex is like sync.RWMutex but with builtin deadlock detecting ability
type RWMutex struct {
	sema *Weighted
}

// NewRWMutex is ctor for RWMutex
func NewRWMutex() *RWMutex {
	rw := &RWMutex{}
	rw.Init()
	return rw
}

// Init for embeded usage
func (rw *RWMutex) Init() {
	rw.sema = NewWeighted(rwmutexMaxReaders, rw)
}

// Lock for write lock
func (rw *RWMutex) Lock() {
	rw.sema.Acquire(context.Background(), rwmutexMaxReaders)
	return
}

// Unlock should only be called after a successful Lock
func (rw *RWMutex) Unlock() {
	rw.sema.Release(rwmutexMaxReaders)
	return
}

// RLock for read lock
func (rw *RWMutex) RLock() {
	rw.sema.Acquire(context.Background(), 1)
}

// RUnlock should only be called after a successful RLock
func (rw *RWMutex) RUnlock() {
	rw.sema.Release(1)
}

// TryLock returns true if lock acquired
func (rw *RWMutex) TryLock() bool {
	return rw.sema.TryAcquire(rwmutexMaxReaders)
}

// TryRLock returns true if rlock acquired
func (rw *RWMutex) TryRLock() bool {
	return rw.sema.TryAcquire(1)
}

func (rw *RWMutex) onAcquiredLocked(n int64) {
	d.onAcquiredLocked(uint64(uintptr(unsafe.Pointer(rw))), n == rwmutexMaxReaders)
}

func (rw *RWMutex) onWaitLocked(n int64) {
	d.onWaitLocked(uint64(uintptr(unsafe.Pointer(rw))), n == rwmutexMaxReaders)
}

func (rw *RWMutex) onWaitCanceledLocked(n int64) {
	panic("this should never happen")
}

func (rw *RWMutex) onReleaseLocked(n int64) {
	d.onReleaseLocked(uint64(uintptr(unsafe.Pointer(rw))), n == rwmutexMaxReaders)
}
