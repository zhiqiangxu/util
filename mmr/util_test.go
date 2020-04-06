package mmr

import (
	"crypto/sha256"
	"testing"

	"gotest.tools/assert"
)

func TestMMR(t *testing.T) {
	size := uint64(0b00111110)

	completeSize := uint64(0b10000)
	assert.Assert(t, minPeakHeight(completeSize) == 4 && maxPeakHeight(completeSize) == 4)
	assert.Assert(t, minPeakHeight(size) == 1)
	assert.Assert(t, maxPeakHeight(size) == 5)
	assert.Assert(t, peakCount(size) == 5)
	assert.Assert(t, leftPeak(size) == 2*0b100000-1 && leftPeakLeaf(size) == 0b100000)

	sizes := getMoutainSizes(size)
	assert.Assert(t,
		sizes[0] == 2*0b100000-1 &&
			sizes[1] == 2*0b10000-1 &&
			sizes[2] == 2*0b1000-1 &&
			sizes[3] == 2*0b100-1 &&
			sizes[4] == 2*0b10-1)

	indices := getMoutainPeaks(size)
	assert.Assert(t,
		indices[0] == 2*0b100000-1 &&
			indices[1] == 2*0b100000-1+2*0b10000-1 &&
			indices[2] == 2*0b100000-1+2*0b10000-1+2*0b1000-1 &&
			indices[3] == 2*0b100000-1+2*0b10000-1+2*0b1000-1+2*0b100-1 &&
			indices[4] == 2*0b100000-1+2*0b10000-1+2*0b1000-1+2*0b100-1+2*0b10-1)

	store, err := NewFileHashStore("/tmp/hs.log", 0)
	assert.Assert(t, err == nil)
	defer store.Close()

	mmr := NewMMR(0, nil, NewHasher([]byte{1}), store)
	h1 := sha256.Sum256([]byte{1})
	mmr.Push(h1, false)
	h1Idx := mmr.Size() - 1

	h2 := sha256.Sum256([]byte{2})
	ap2 := mmr.Push(h2, true)
	h2Idx := mmr.Size() - 1
	assert.Assert(t, len(ap2) == proofLength(h2Idx, mmr.Size()) && ap2[0] == h1)

	rootHash2 := mmr.Root()

	h3 := sha256.Sum256([]byte{3})
	ap3 := mmr.Push(h3, true)
	h3Idx := mmr.Size() - 1
	assert.Assert(t, len(ap3) == proofLength(h3Idx, mmr.Size()))

	// h2's proof is returned by Push
	err = mmr.VerifyExists(h2, rootHash2, h2Idx, h2Idx+1, ap2)
	assert.Assert(t, err == nil)

	// h3's proof is returned by Push
	err = mmr.VerifyExists(h3, mmr.Root(), h3Idx, h3Idx+1, ap3)
	assert.Assert(t, err == nil)

	// generate proof for h1 wrt current root
	proof, err := mmr.GenProof(h1Idx, mmr.Size())
	assert.Assert(t, err == nil)
	err = mmr.VerifyExists(h1, mmr.Root(), h1Idx, mmr.Size(), proof)
	assert.Assert(t, err == nil)
}
