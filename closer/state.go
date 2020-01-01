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
	done    chan struct{}
}

// NewState is ctor for State
func NewState() *State {
	return &State{done: make(chan struct{})}
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

// CloseSignal gets signaled when Wait() is called.
func (s *State) CloseSignal() <-chan struct{} {
	return s.done
}

// Done decrements the WaitGroup counter by one.
func (s *State) Done() {
	s.waiting.Add(-1)
}

// SignalAndWait updates closed and blocks until the WaitGroup counter is zero.
// Call it more than once will panic
func (s *State) SignalAndWait() {
	close(s.done)

	s.mu.Lock()

	atomic.StoreUint32(&s.closed, 1)
	// s.waiting.WaitRelease(func(){
	// 	s.mu.Unlock()
	// })
	s.waiting.Wait()
	s.mu.Unlock()
}
