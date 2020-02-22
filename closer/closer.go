package closer

// Closer defines the shape of every closer
// Specific closers may have more helper methods and different guarantees
type Closer interface {
	Add(delta int)
	SignalAndWait()
	Done() // alias for Add(-1)
}

var _ Closer = (*Strict)(nil)

var _ Closer = (*Naive)(nil)
