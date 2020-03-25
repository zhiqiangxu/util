package mcas

import (
	"testing"
	"unsafe"

	"gotest.tools/assert"
)

func TestMCAS(t *testing.T) {
	v1 := 1
	v2 := 2
	v3 := 3
	v4 := 4

	// try to compare and swap p1 and p2 atomically

	// p1 and p2 initially points to v1 and v2
	p1 := unsafe.Pointer(&v1)
	p2 := unsafe.Pointer(&v2)

	a := []*unsafe.Pointer{&p1, &p2}
	e := []unsafe.Pointer{unsafe.Pointer(&v1), unsafe.Pointer(&v2)}
	n := []unsafe.Pointer{unsafe.Pointer(&v3), unsafe.Pointer(&v4)}

	swapped := CompareAndSwap(a, e, n)
	assert.Assert(t, swapped)

	// assert that p1 and p2 should be swapped to v3 and v4
	p1v := Read(&p1)
	p2v := Read(&p2)
	assert.Assert(t, p1v == unsafe.Pointer(&v3) && p2v == unsafe.Pointer(&v4))
	assert.Assert(t, p1 == p1v && p2 == p2v)

	swapped = CompareAndSwap(a, e, n)
	assert.Assert(t, !swapped)

	p1v = Read(&p1)
	p2v = Read(&p2)
	assert.Assert(t, p1v == unsafe.Pointer(&v3) && p2v == unsafe.Pointer(&v4))
}
