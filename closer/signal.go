package closer

import "sync"

// Signal holds the two things we need to close a goroutine and wait for it to finish: a chan
// to tell the goroutine to shut down, and a WaitGroup with which to wait for it to finish shutting
// down.
type Signal struct {
	waiting sync.WaitGroup
	done    chan struct{}
}

// NewSignal is ctor for Signal
func NewSignal() *Signal {
	return &Signal{done: make(chan struct{})}
}

// WaitGroupRef returns the reference of WaitGroup
func (s *Signal) WaitGroupRef() *sync.WaitGroup {
	return &s.waiting
}

// Add delta to WaitGroup
func (s *Signal) Add(delta int) {
	s.waiting.Add(delta)
}

// CloseSignal gets signaled when Wait() is called.
func (s *Signal) CloseSignal() <-chan struct{} {
	return s.done
}

// Done calls Done() on the WaitGroup.
func (s *Signal) Done() {
	s.waiting.Done()
}

// SignalAndWait closes chan and wait on the WaitGroup.
// Call it more than once will panic
func (s *Signal) SignalAndWait() {
	close(s.done)
	s.waiting.Wait()
}
