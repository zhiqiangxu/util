package util

import "sync"

// Closer holds the two things we need to close a goroutine and wait for it to finish: a chan
// to tell the goroutine to shut down, and a WaitGroup with which to wait for it to finish shutting
// down.
type Closer struct {
	closed  chan struct{}
	waiting sync.WaitGroup
}

// NewCloser is ctor for Closer
func NewCloser() *Closer {
	ret := &Closer{closed: make(chan struct{})}
	return ret
}

// AddRunning for add delta to WaitGroup
func (lc *Closer) AddRunning(delta int) {
	lc.waiting.Add(delta)
}

// Signal closes the chan
func (lc *Closer) Signal() {
	close(lc.closed)
}

// HasBeenClosed gets signaled when Signal() is called.
func (lc *Closer) HasBeenClosed() <-chan struct{} {
	return lc.closed
}

// Done calls Done() on the WaitGroup.
func (lc *Closer) Done() {
	lc.waiting.Done()
}

// Wait on the WaitGroup.
func (lc *Closer) Wait() {
	lc.waiting.Wait()
}

// SignalAndWait calls Signal(), then Wait().
func (lc *Closer) SignalAndWait() {
	lc.Signal()
	lc.Wait()
}
