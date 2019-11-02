package util

import (
	"testing"

	"gotest.tools/assert"
)

func TestByteArena(t *testing.T) {
	a := NewByteArena()
	n := 10
	bytes := a.AllocBytes(n)
	assert.Assert(t, len(bytes) == n && cap(bytes) == n)
}
