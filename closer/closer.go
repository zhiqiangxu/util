package closer

// Closer defines the shape of every closer
// Individual closers may have more helper methods
type Closer interface {
	Add(delta int)
	SignalAndWait()
	Done() // alias for Add(-1)
}

var _ Closer = (*State)(nil)

var _ Closer = (*Signal)(nil)
