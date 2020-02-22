package wm

import (
	"context"
	"sync/atomic"
)

// Max watermark model, a tmutex with capacity
// The entire point of this module is to reduce lock contention under fast path
type Max struct {
	count   int64         // total number of participators
	entered int64         // total number of entered
	max     int64         // max number of entered, immutable
	ch      chan struct{} // channel for exit signal from entered
}

// NewMax is ctor for Max
func NewMax(max int64) *Max {
	m := &Max{max: max, ch: make(chan struct{}, max)}
	return m
}

// Enter returns nil if not canceled
// Enter guarantees at most N callers will enter successfully at any moment, and will wait when overflow
func (m *Max) Enter(ctx context.Context) error {
	count := m.incCount()

	// slow path, wait for exit signal before try
	contendFunc := func() error {
		for {
			select {
			case <-m.ch:
				// someone exited, try enter again
				if m.tryAddEntered() {
					return nil
				}
				continue
			case <-ctx.Done():
				// ctx canceled, restore count before return
				m.decCount()
				return ctx.Err()
			}
		}
	}

	if count > m.max {
		// too many participators, wait for someone to exit
		return contendFunc()
	}

	if !m.tryAddEntered() {
		// this means there is a sudden burst of participators that entered in very short instant in other goroutines
		// this should happen very rarely
		return contendFunc()
	}
	return nil
}

//go:nosplit
func (m *Max) tryAddEntered() bool {
	if atomic.AddInt64(&m.entered, 1) <= m.max {
		return true
	}

	// recover before return
	m.decEntered()
	return false
}

//go:nosplit
func (m *Max) decEntered() {
	atomic.AddInt64(&m.entered, -1)
}

// TryEnter returns true if succeed
func (m *Max) TryEnter() bool {
	count := m.incCount()
	if count > m.max {
		m.decCount()
		return false
	}

	if !m.tryAddEntered() {
		m.decCount()
		return false
	}
	return true
}

// Exit should only be called if Enter returns nil
func (m *Max) Exit() {

	m.decEntered()
	count := m.decCount()

	// two cases to consider:
	// 1. remaining count < Max
	// 		can there be any waiters?
	//		prove that if there are X(<=count<Max) waiters, there must be at least X signals in ch.
	//		suppose there are X waiters,
	//		then they must have tried to enter at a moment when there are at least Max participators or exactly Max entered
	//		consider the last waiter, when its Max entered callers exit, the first X will see the X waiters(Max-i+X>=Max),
	//		so at least X signals will be dilivered.
	//
	// 2. ch is full
	//		will some waiters miss the signal?

	// remaining count >= max
	if count >= m.max {
		// try to wake some waiter up by exit signal.
		// buffered channel ensures never miss any signal
		// the waiter may have been canceled by ctx, but it doesn't matter,
		// superfluous signals only matters for performance purpose
		select {
		case m.ch <- struct{}{}:
		default:
		}
	}

}

//go:nosplit
func (m *Max) incCount() int64 {
	return atomic.AddInt64(&m.count, 1)
}

//go:nosplit
func (m *Max) decCount() int64 {
	count := atomic.AddInt64(&m.count, -1)
	if count < 0 {
		panic("Max: count < 0")
	}

	return count
}
