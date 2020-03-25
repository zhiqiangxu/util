package mcas

import "unsafe"

// CompareAndSwap for multiple pointer type variables
func CompareAndSwap(a []*unsafe.Pointer, e []unsafe.Pointer, n []unsafe.Pointer) (swapped bool) {
	d := &mcDesc{a: a, e: e, n: n, s: undecided}
	/* Memory locations must be sorted into address order. */
	d.sortAddr()
	swapped = d.mcasHelp()
	return
}

// Read for a mcas consistent view
func Read(a *unsafe.Pointer) (v unsafe.Pointer) {

	for v = ccasRead(a); isMCDesc(v); ccasRead(a) {
		mcfromPointer(v).mcasHelp()
	}

	return
}
