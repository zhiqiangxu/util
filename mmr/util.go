package mmr

import (
	"math/bits"
)

// minPeakHeight equals position of the lowest 1 bit
// zero based height
func minPeakHeight(size uint64) int {
	return bits.TrailingZeros64(size)
}

func maxPeakHeight(size uint64) int {
	highBit := 64 - bits.LeadingZeros64(size)
	return highBit - 1
}

// peakCount equals the number of 1 bit
func peakCount(size uint64) int {
	return bits.OnesCount64(size)
}

func bagPeaks(hasher Hasher, peaks []HashType) HashType {
	l := len(peaks)
	accum := peaks[l-1]

	for i := l - 2; i >= 0; i-- {
		accum = hasher.Node(peaks[i], accum)
	}
	return accum
}

// leftPeak returns left peak
// 1 based index
// precondition: size >= 1
func leftPeak(size uint64) uint64 {
	highBit := 64 - bits.LeadingZeros64(size)
	return uint64(1<<highBit) - 1
}

// leftPeakLeaf returns the leaf of left peak
func leftPeakLeaf(size uint64) uint64 {
	highBit := 64 - bits.LeadingZeros64(size)
	return 1 << (highBit - 1)
}

// getMoutainSizes returns moutain sizes from left to right
func getMoutainSizes(size uint64) []uint64 {
	nPeaks := bits.OnesCount64(size)

	moutainSizes := make([]uint64, nPeaks, nPeaks)
	for i, id := nPeaks-1, uint64(1); size != 0; size = size >> 1 {
		id = id * 2
		if size%2 == 1 {
			moutainSizes[i] = id - 1
			i--
		}
	}

	return moutainSizes
}

// getMoutainPeaks returns moutain peaks from left to right
// 1 based index
func getMoutainPeaks(size uint64) []uint64 {
	nPeaks := bits.OnesCount64(size)
	peakPos := make([]uint64, nPeaks, nPeaks)
	for i, id := nPeaks-1, uint64(1); size != 0; size = size >> 1 {
		id = id * 2
		if size%2 == 1 {
			peakPos[i] = id - 1
			i--
		}
	}

	for i := 1; i < nPeaks; i++ {
		peakPos[i] += peakPos[i-1]
	}

	return peakPos
}

func proofLength(index, size uint64) int {
	length := 0
	lastNode := size - 1
	for lastNode > 0 {
		if index%2 == 1 || index < lastNode {
			length += 1
		}
		index /= 2
		lastNode /= 2
	}
	return length
}

var defaultHasher = NewHasher([]byte{0}, []byte{1})

// ComputeRoot ...
func ComputeRoot(hashes []HashType) (root HashType) {
	if len(hashes) == 0 {
		return
	}

	mmr := NewMMR(0, nil, defaultHasher, nil)
	for _, hash := range hashes {
		mmr.PushHash(hash, false)
	}

	root = mmr.Root()

	return
}
