package mcas

import (
	"sync/atomic"
	"unsafe"

	"github.com/zhiqiangxu/util/bytes"
)

type arena struct {
	offset uint32
	buf    []byte
}

const (
	mcDescSize = uint32(unsafe.Sizeof(mcDesc{}))
	ccDescSize = uint32(unsafe.Sizeof(ccDesc{}))
)

func newArena(size uint32) *arena {
	return &arena{buf: bytes.AlignedTo8(size)}
}

func (a *arena) alloc(size uint32) (offset uint32) {
	if int(size) > len(a.buf) {
		panic("size > buf size")
	}

	// Pad the allocation with enough bytes to ensure pointer alignment.
	l := uint32(size + bytes.Align8Mask)

try:
	n := atomic.AddUint32(&a.offset, l)
	if int(n) > len(a.buf) {
		if atomic.CompareAndSwapUint32(&a.offset, 0, l) {
			n = l
			goto final
		}
		goto try
	}

final:
	// Return the aligned offset.
	offset = (n - l + uint32(bytes.Align8Mask)) & ^uint32(bytes.Align8Mask)
	return
}

func (a *arena) getPointer(ptr uintptr) unsafe.Pointer {
	offset := ptr - uintptr(unsafe.Pointer(&a.buf[0]))
	return unsafe.Pointer(&a.buf[offset])
}

func (a *arena) putMCDesc() *mcDesc {
	offset := a.alloc(mcDescSize)
	return (*mcDesc)(unsafe.Pointer(&a.buf[offset]))
}

func (a *arena) putCCDesc() *ccDesc {
	offset := a.alloc(ccDescSize)
	return (*ccDesc)(unsafe.Pointer(&a.buf[offset]))
}

var are *arena

func init() {
	// enough for 10w concurrency
	are = newArena(1024 * 1024)
}
