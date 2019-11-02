package util

// ByteArena for reduce GC pressure
type ByteArena struct {
	alloc []byte
}

// NewByteArena is ctor for ByteArena
func NewByteArena() *ByteArena {
	return &ByteArena{}
}

// AllocBytes for allocate bytes
func (a *ByteArena) AllocBytes(n int) (bytes []byte) {
	if cap(a.alloc)-len(a.alloc) < n {
		a.reserve(n)
	}

	pos := len(a.alloc)
	bytes = a.alloc[pos : pos+n : pos+n]
	a.alloc = a.alloc[:pos+n]

	return
}

// UnsafeReset for reuse
func (a *ByteArena) UnsafeReset() {
	a.alloc = a.alloc[:0]
}

func (a *ByteArena) reserve(n int) {
	const chunkAllocMinSize = 512
	const chunkAllocMaxSize = 16384

	allocSize := cap(a.alloc) * 2
	if allocSize < chunkAllocMinSize {
		allocSize = chunkAllocMinSize
	} else if allocSize > chunkAllocMaxSize {
		allocSize = chunkAllocMaxSize
	}
	if allocSize < n {
		allocSize = n
	}
	a.alloc = make([]byte, 0, allocSize)
}
