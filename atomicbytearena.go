package util

import (
	"sync"
	"sync/atomic"
)

// AtomicByteArena is concurrent safe , it's atomic when ballast is enough:)
// it fallbacks to lock when allocating new ballast
type AtomicByteArena struct {
	sync.RWMutex
	ballast   []byte
	offset    int64
	chunkSize int
}

// NewAtomicByteArena is ctor for AtomicByteArena
func NewAtomicByteArena(chunkSize int) *AtomicByteArena {
	return &AtomicByteArena{ballast: make([]byte, chunkSize), chunkSize: chunkSize}
}

// AllocBytes does what it says:)
func (aa *AtomicByteArena) AllocBytes(n int) (bytes []byte) {

	if n > aa.chunkSize/2 {
		return make([]byte, n)
	}

	nint64 := int64(n)
	for {

		aa.RLock()

		newOffset := atomic.AddInt64(&aa.offset, nint64)
		if newOffset <= int64(len(aa.ballast)) {
			bytes = aa.ballast[newOffset-nint64 : newOffset]
			aa.RUnlock()
			return
		}
		aa.RUnlock()

		// need to allocate new ballast

		aa.Lock()

		// double check
		newOffset = atomic.AddInt64(&aa.offset, nint64)
		if newOffset <= int64(len(aa.ballast)) {
			bytes = aa.ballast[newOffset-nint64 : newOffset]
			aa.Unlock()
			return
		}

		aa.ballast = make([]byte, len(aa.ballast))

		bytes = aa.ballast[0:n]
		atomic.StoreInt64(&aa.offset, nint64)
		aa.Unlock()

		return
	}

}
