package closer

import (
	"sync"
	"sync/atomic"
)

// Strict closer is a sync.WaitGroup with state.
// It guarantees no Add with positive delta will ever succeed after Wait.
// Wait can only be called once, but Add and Wait can be called concurrently.
// The happens before relationship between Add and Wait is taken care of automatically.
type Strict struct {
	mu      sync.RWMutex
	waiting sync.WaitGroup
	closed  uint32
	done    chan struct{}
}

// NewStrict is ctor for Strict
func NewStrict() *Strict {
	return &Strict{done: make(chan struct{})}
}

// Add delta to wait group
// Trying to Add positive delta after Wait will panic
func (s *Strict) Add(delta int) {
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
func (s *Strict) HasBeenClosed() bool {
	return atomic.LoadUint32(&s.closed) != 0
}

// ClosedSignal gets signaled when Wait() is called.
func (s *Strict) ClosedSignal() <-chan struct{} {
	return s.done
}

// Done decrements the WaitGroup counter by one.
func (s *Strict) Done() {
	s.waiting.Add(-1)
}

// SignalAndWait updates closed and blocks until the WaitGroup counter is zero.
// Call it more than once will panic
func (s *Strict) SignalAndWait() {
	close(s.done)

	s.mu.Lock()

	atomic.StoreUint32(&s.closed, 1)
	// s.waiting.WaitRelease(func(){
	// 	s.mu.Unlock()
	// })
	s.waiting.Wait()
	s.mu.Unlock()
}
