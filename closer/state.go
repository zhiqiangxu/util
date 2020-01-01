package closer

import (
	"sync"
	"sync/atomic"
)

// State closer is a sync.WaitGroup with state
// It guarantees no Add with positive delta will ever succeed after Wait
type State struct {
	mu      sync.RWMutex
	waiting sync.WaitGroup
	closed  uint32
}

// Add delta to wait group
// Trying to Add positive delta after Wait will panic
func (s *State) Add(delta int) {
	if delta > 0 {
		if s.HasBeenClosed() {
			panic("Add after Wait")
		}

		s.mu.RLock()
		if s.closed != 0 {
			s.mu.RUnlock()
			panic("Add after Wait")
		}
	}

	s.waiting.Add(delta)

	if delta > 0 {
		s.mu.RUnlock()
	}
}

// HasBeenClosed tells whether closed
func (s *State) HasBeenClosed() bool {
	return atomic.LoadUint32(&s.closed) != 0
}

// Done decrements the WaitGroup counter by one.
func (s *State) Done() {
	s.waiting.Add(-1)
}

// SignalAndWait updates closed and blocks until the WaitGroup counter is zero.
// Call it more than once will panic
func (s *State) SignalAndWait() {
	s.mu.Lock()
	if s.closed != 0 {
		s.mu.Unlock()
		panic("Wait more than once")
	}

	atomic.StoreUint32(&s.closed, 1)
	s.waiting.Wait()
	s.mu.Unlock()
}
