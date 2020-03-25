package mcas

import (
	"sync/atomic"
	"unsafe"
)

type ccDesc struct {
	a  *unsafe.Pointer
	e  unsafe.Pointer
	n  unsafe.Pointer
	sp *uint32
}

func ccasRead(a *unsafe.Pointer) (v unsafe.Pointer) {
	for v = atomic.LoadPointer(a); isCCDesc(v); v = atomic.LoadPointer(a) {
		ccfromPointer(v).ccasHelp()
	}
	return
}

func isCCDesc(v unsafe.Pointer) bool {
	return uintptr(v)&addrMask == ccDescAddr
}

func ccfromPointer(v unsafe.Pointer) *ccDesc {
	ptr := uintptr(v)
	ptr = ptr & ^uintptr(addrMask)
	return (*ccDesc)(unsafe.Pointer(ptr))
}

func (d *ccDesc) toPointer() unsafe.Pointer {
	return unsafe.Pointer(uintptr(unsafe.Pointer(d)) + uintptr(ccDescAddr))
}

func ccas(a *unsafe.Pointer, e, n unsafe.Pointer, sp *uint32) (ok, swapped, isn bool) {
	d := &ccDesc{a: a, e: e, n: n, sp: sp}
	var v unsafe.Pointer
	for !atomic.CompareAndSwapPointer(d.a, d.e, d.toPointer()) {
		v = atomic.LoadPointer(d.a)
		if !isCCDesc(v) {
			return
		}
		ccfromPointer(v).ccasHelp()
	}
	ok = true
	swapped, isn = d.ccasHelp()
	return
}

func (d *ccDesc) ccasHelp() (swapped, isn bool) {
	s := atomic.LoadUint32(d.sp)
	var v unsafe.Pointer
	if s == undecided {
		isn = true
		v = d.n
	} else {
		v = d.e
	}
	swapped = atomic.CompareAndSwapPointer(d.a, d.toPointer(), v)
	return
}
