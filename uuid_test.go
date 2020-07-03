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

	// test empty interface
	var i interface{}

	type test struct {
		a int
		b string
	}

	var s test
	s.a = 1
	s.b = "1"

	i = s
	s.a = 2
	s.b = "2"
	assert.Assert(t, i.(test).a == 1)

	var i2 interface{}

	i2 = i

	i = s
	s.a = 3
	s.b = "3"

	assert.Assert(t, i.(test).a == 2 && i2.(test).a == 1)

	// // this will error
	// i2.(test).a = 3

	// test slice
	{
		encID := make([]byte, 0, 10)
		_ = append(encID, 'a')
		assert.Assert(t, len(encID) == 0 && cap(encID) == 10)
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
