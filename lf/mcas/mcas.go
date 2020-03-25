package mcas

import (
	"sort"
	"sync/atomic"
	"unsafe"
)

type mcDesc struct {
	a []*unsafe.Pointer
	e []unsafe.Pointer
	n []unsafe.Pointer
	s uint32
}

const (
	undecided uint32 = iota
	failed
	successful
)

func isMCDesc(v unsafe.Pointer) bool {
	return uintptr(v)&addrMask == mcDescAddr
}

const (
	mcDescAddr = 1
	ccDescAddr = 2
	addrMask   = 3
)

func mcfromPointer(v unsafe.Pointer) *mcDesc {
	ptr := uintptr(v)
	ptr = ptr & ^uintptr(addrMask)
	return (*mcDesc)(unsafe.Pointer(ptr))
}

func (d *mcDesc) toPointer() unsafe.Pointer {
	return unsafe.Pointer(uintptr(unsafe.Pointer(d)) + uintptr(mcDescAddr))
}

func (d *mcDesc) sortAddr() {
	sort.Slice(d.a, func(i, j int) bool {
		return uintptr(unsafe.Pointer(d.a[i])) < uintptr(unsafe.Pointer(d.a[j]))
	})
}

func (d *mcDesc) status() uint32 {
	return atomic.LoadUint32(&d.s)
}

func (d *mcDesc) mcasHelp() (suc bool) {
	ds := failed
	var (
		v unsafe.Pointer
	)
	/* PHASE 1: Attempt to acquire each location in turn. */
	for i := range d.a {
		for {
			ccas(d.a[i], d.e[i], d.toPointer(), &d.s)
			v = atomic.LoadPointer(d.a[i])
			if v == d.e[i] && d.status() == undecided {
				continue
			}
			if v == d.toPointer() {
				break
			}
			if !isMCDesc(v) {
				goto decision_point
			}
			mcfromPointer(v).mcasHelp()
		}
	}

	ds = successful
decision_point:

	atomic.CompareAndSwapUint32(&d.s, undecided, ds)

	/* PHASE 2: Release each location that we hold. */
	suc = atomic.LoadUint32(&d.s) == successful
	for i := range d.a {
		if suc {
			atomic.CompareAndSwapPointer(d.a[i], d.toPointer(), d.n[i])
		} else {
			atomic.CompareAndSwapPointer(d.a[i], d.toPointer(), d.e[i])
		}
	}

	return
}
