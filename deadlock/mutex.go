package deadlock

import (
	"context"
	"unsafe"
)

// Mutex is like sync.Mutex but with builtin deadlock detecting ability
type Mutex struct {
	sema *Weighted
}

// NewMutex is ctor for Mutex
func NewMutex() *Mutex {
	m := &Mutex{}
	m.Init()
	return m
}

// Init for embeded usage
func (m *Mutex) Init() {
	m.sema = NewWeighted(1, m)
}

// Lock with context
func (m *Mutex) Lock() (err error) {
	err = m.sema.Acquire(context.Background(), 1)
	return
}

// Unlock should only be called after a successful Lock
func (m *Mutex) Unlock() {
	m.sema.Release(1)
}

// TryLock returns true if lock acquired
func (m *Mutex) TryLock() bool {
	return m.sema.TryAcquire(1)
}

func (m *Mutex) onAcquiredLocked(n int64) {
	d.onAcquiredLocked(uint64(uintptr(unsafe.Pointer(m))), true)
}

func (m *Mutex) onWaitLocked(n int64) {
	d.onWaitLocked(uint64(uintptr(unsafe.Pointer(m))), true)
}

func (m *Mutex) onWaitCanceledLocked(n int64) {
	panic("this should never happen")
}

func (m *Mutex) onReleaseLocked(n int64) {
	d.onReleaseLocked(uint64(uintptr(unsafe.Pointer(m))), true)
}
