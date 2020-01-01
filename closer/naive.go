package closer

import "sync"

// Naive holds the two things we need to close a goroutine and wait for it to finish: a chan
// to tell the goroutine to shut down, and a WaitGroup with which to wait for it to finish shutting
// down.
// The happens before relationship between Add and Wait is guaranteed by the caller side.
type Naive struct {
	waiting sync.WaitGroup
	done    chan struct{}
}

// NewNaive is ctor for Naive
func NewNaive() *Naive {
	return &Naive{done: make(chan struct{})}
}

// WaitGroupRef returns the reference of WaitGroup
func (n *Naive) WaitGroupRef() *sync.WaitGroup {
	return &n.waiting
}

// Add delta to WaitGroup
func (n *Naive) Add(delta int) {
	n.waiting.Add(delta)
}

// CloseSignal gets signaled when Wait() is called.
func (n *Naive) CloseSignal() <-chan struct{} {
	return n.done
}

// Done calls Done() on the WaitGroup.
func (n *Naive) Done() {
	n.waiting.Done()
}

// SignalAndWait closes chan and wait on the WaitGroup.
// Call it more than once will panic
func (n *Naive) SignalAndWait() {
	close(n.done)
	n.waiting.Wait()
}
