package util

import (
	"math"
	_ "unsafe" // required by go:linkname
)

// FYI: https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
//		https://github.com/golang/go/blob/b5c66de0892d0e9f3f59126eeebc31070e79143b/src/runtime/stubs.go#L115

// FastRand returns a lock free uint32 value.
//go:linkname FastRand runtime.fastrand
func FastRand() uint32

// FastRandN returns a random number in [0, n)
//go:linkname FastRandN runtime.fastrandn
func FastRandN(uint32) uint32

// FastRand64 returns a random uint64 without lock
func FastRand64() (result uint64) {
	a := FastRand()
	b := FastRand()
	result = uint64(a)<<32 + uint64(b)
	return
}

// FastRand64N returns, as an uint64, a pseudo-random number in [0,n).
func FastRand64N(n uint64) uint64 {
	v := FastRand64()
	if n&(n-1) == 0 { // n is power of two, can mask
		return v & (n - 1)
	}
	return v % n
}

// PoorManUUID generate a uint64 uuid
func PoorManUUID(client bool) (result uint64) {

	result = PoorManUUID2()
	if client {
		result |= 1 //odd for client
	} else {
		result &= math.MaxUint64 - 1 //even for server
	}
	return
}

// PoorManUUID2 doesn't care whether client/server side
func PoorManUUID2() uint64 {
	return FastRand64()
}
