package diskqueue

import (
	"testing"
	"unsafe"

	"gotest.tools/assert"
)

func TestMeta(t *testing.T) {
	size := unsafe.Sizeof(FileMeta{})
	assert.Assert(t, size == 40)
}
