package bytes

import (
	"testing"
	"unsafe"

	"gotest.tools/assert"
)

func TestAligned(t *testing.T) {
	s1 := AlignedTo8(1)
	assert.Assert(t, uintptr(unsafe.Pointer(&s1[0]))%8 == 0 && len(s1) == 1)
	s2 := AlignedTo4(1)
	assert.Assert(t, uintptr(unsafe.Pointer(&s2[0]))%4 == 0 && len(s2) == 1)
}
