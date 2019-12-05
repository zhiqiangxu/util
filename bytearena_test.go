package util

import (
	"testing"

	"gotest.tools/assert"
)

func TestByteArena(t *testing.T) {
	a := NewByteArena(512, 16384)
	n := 10
	bytes := a.AllocBytes(n)
	assert.Assert(t, len(bytes) == n && cap(bytes) == n)
}
