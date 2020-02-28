package tmutex

import "sync/atomic"

// TMutex is a mutual exclusion primitive that implements TryLock in addition
// to Lock and Unlock.
type TMutex struct {
	v  int32
	ch chan struct{}
}

// New is ctor for TMutex
func New() *TMutex {
	return &TMutex{v: 1, ch: make(chan struct{}, 1)}
}

// Lock has the same semantics as normal Mutex.Lock
func (m *TMutex) Lock() {
	// Uncontended case.
	if atomic.AddInt32(&m.v, -1) == 0 {
		return
	}

	for {

		// if v < 0, someone is already contending, just wait for lock release
		// otherwise, SwapInt32 can only return one of -1, 0, 1
		// 1 means no contention
		// -1 means someone is already contending
		// 0 means no one else is contending but the lock hasn't been released
		// so for -1 or 0, just wait for lock release
		if v := atomic.LoadInt32(&m.v); v >= 0 && atomic.SwapInt32(&m.v, -1) == 1 {
			return
		}

		// Wait for the mutex to be released before trying again.
		<-m.ch
	}
}

// TryLock returns true on success
func (m *TMutex) TryLock() bool {
	v := atomic.LoadInt32(&m.v)
	if v <= 0 {
		return false
	}
	return atomic.CompareAndSwapInt32(&m.v, 1, 0)
}

// Unlock has the same semantics as normal Mutex.Unlock
func (m *TMutex) Unlock() {
	if atomic.SwapInt32(&m.v, 1) == 0 {
		// There were no pending waiters.
		return
	}

	// Wake some waiter up.
	select {
	case m.ch <- struct{}{}:
	default:
	}
}
