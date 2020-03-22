package util

import (
	"testing"

	"golang.org/x/exp/rand"
	"gotest.tools/assert"
)

func TestUUID(t *testing.T) {
	assert.Assert(t, PoorManUUID(true)%2 == 1)
	assert.Assert(t, PoorManUUID(false)%2 == 0)
	assert.Assert(t, FastRandN(1) == 0)
	for i := 2; i < 100; i++ {
		assert.Assert(t, FastRandN(uint32(i)) < uint32(i))
	}

}

func BenchmarkFastRand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FastRand()
	}
}

func BenchmarkPCG(b *testing.B) {
	r := rand.PCGSource{}
	for i := 0; i < b.N; i++ {
		r.Uint64()
	}
}
