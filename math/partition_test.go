package math

import (
	"testing"

	"gotest.tools/assert"
)

func TestPartition(t *testing.T) {
	p1 := Partition(1)
	p2 := Partition(2)
	p3 := Partition(3)
	p4 := Partition(4)
	p5 := Partition(5)
	p6 := Partition(6)

	assert.Equal(t, p1.Uint64(), uint64(1))
	assert.Equal(t, p2.Uint64(), uint64(2))
	assert.Equal(t, p3.Uint64(), uint64(3))
	assert.Equal(t, p4.Uint64(), uint64(5))
	assert.Equal(t, p5.Uint64(), uint64(7))
	assert.Equal(t, p6.Uint64(), uint64(11))
}
