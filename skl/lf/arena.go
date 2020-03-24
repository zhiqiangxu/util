package lf

import (
	"errors"
	"sync/atomic"
	"unsafe"
)

// Arena is lock free
type Arena struct {
	n   uint32
	buf []byte
}

var (
	// ErrOOM used by Arena
	ErrOOM = errors.New("oom")
)

// NewArena is ctor for Arena
func NewArena(n uint32) *Arena {

	// offset 0 is reserved for nil value
	out := &Arena{n: 1, buf: makeAlignedBuf(n)}

	return out
}

func makeAlignedBuf(n uint32) []byte {
	buf := make([]byte, int(n)+nodeAlign)
	buf0Alignment := uint32(uintptr(unsafe.Pointer(&buf[0]))) & uint32(nodeAlign)
	buf = buf[buf0Alignment : buf0Alignment+n]
	return buf
}

func (a *Arena) putKV(k, v []byte) (koff, voff uint32, err error) {
	lk := uint32(len(k))
	lv := uint32(len(v))
	l := lk + lv
	n := atomic.AddUint32(&a.n, l)
	if int(n) > len(a.buf) {
		err = ErrOOM
		return
	}

	koff = n - l
	copy(a.buf[koff:koff+lk], k)
	voff = koff + lk
	copy(a.buf[voff:voff+lv], v)
	return
}

func (a *Arena) putBytes(b []byte) (offset uint32, err error) {
	l := uint32(len(b))
	n := atomic.AddUint32(&a.n, l)
	if int(n) > len(a.buf) {
		err = ErrOOM
		return
	}

	offset = n - l
	copy(a.buf[offset:n], b)
	return

}

const (
	nodeAlign = int(unsafe.Sizeof(uint64(0))) - 1
)

func (a *Arena) putListNode() (offset uint32, err error) {

	// Pad the allocation with enough bytes to ensure pointer alignment.
	l := uint32(ListNodeSize + nodeAlign)
	n := atomic.AddUint32(&a.n, l)
	if int(n) > len(a.buf) {
		err = ErrOOM
		return
	}

	// Return the aligned offset.
	offset = (n - l + uint32(nodeAlign)) & ^uint32(nodeAlign)
	return
}

func (a *Arena) getBytes(offset uint32, size uint16) []byte {
	if offset == 0 {
		return nil
	}

	return a.buf[offset : offset+uint32(size)]
}
func (a *Arena) getListNode(offset uint32) *listNode {
	if offset == 0 {
		return nil
	}

	return (*listNode)(unsafe.Pointer(&a.buf[offset]))
}

func (a *Arena) getListNodeOffset(n *listNode) uint32 {
	if n == nil {
		return 0
	}

	offset := uintptr(unsafe.Pointer(n)) - uintptr(unsafe.Pointer(&a.buf[0]))
	return uint32(offset)
}
