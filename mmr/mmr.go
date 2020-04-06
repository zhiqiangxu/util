package mmr

import (
	"errors"
	"fmt"
)

// HashType can be replaced by https://github.com/zhiqiangxu/gg
type HashType [32]byte

// Hasher for mmr
type Hasher interface {
	Empty() HashType
	Node(left, right HashType) HashType
}

// HashStore for store sequential hashes and fetch by index
type HashStore interface {
	Append(hashes []HashType) error
	Flush() error
	Close()
	GetHash(offset uint64) (HashType, error)
}

// MMR for MerkleMoutainRange
type MMR struct {
	size          uint64
	peaks         []HashType
	rootHash      HashType
	minPeakHeight int
	hasher        Hasher
	store         HashStore
}

var unknownHash HashType

// NewMMR is ctor for MMR
func NewMMR(size uint64, peaks []HashType, hasher Hasher, store HashStore) *MMR {
	m := &MMR{hasher: hasher, store: store}
	m.update(size, peaks)
	return m
}

func (m *MMR) update(size uint64, peaks []HashType) {
	if len(peaks) != peakCount(size) {
		panic("number of peaks != peakCount")
	}
	m.size = size
	m.peaks = peaks
	m.minPeakHeight = minPeakHeight(size)
	m.rootHash = unknownHash
}

// Size of mmr
func (m *MMR) Size() uint64 {
	return m.size
}

// Push a hash
func (m *MMR) Push(h HashType, wantAP bool) (ap []HashType) {
	psize := len(m.peaks)

	if wantAP {
		ap = make([]HashType, psize, psize)
		// reverse
		for i, v := range m.peaks {
			ap[psize-i-1] = v
		}
	}

	newHashes := []HashType{h}
	m.minPeakHeight = 0
	for s := m.size; s%2 == 1; s = s >> 1 {
		m.minPeakHeight++
		h = m.hasher.Node(m.peaks[psize-1], h)
		newHashes = append(newHashes, h)
		psize--
	}

	if m.store != nil {
		m.store.Append(newHashes)
		m.store.Flush()
	}

	m.size++
	m.peaks = m.peaks[0:psize]
	m.peaks = append(m.peaks, h)
	m.rootHash = unknownHash

	return
}

// Root returns root hash
func (m *MMR) Root() HashType {

	if m.rootHash == unknownHash {
		if len(m.peaks) > 0 {
			m.rootHash = bagPeaks(m.hasher, m.peaks)
		} else {
			m.rootHash = m.hasher.Empty()
		}
	}
	return m.rootHash
}

var (
	// ErrRootNotAvailableYet used by MMR
	ErrRootNotAvailableYet = errors.New("not available yet")
	// ErrHashStoreNotAvailable used by MMR
	ErrHashStoreNotAvailable = errors.New("hash store not available")
)

// GenProof returns the audit path of ti wrt size
func (m *MMR) GenProof(leafIdx, size uint64) (hashes []HashType, err error) {
	if leafIdx >= size {
		err = fmt.Errorf("wrong parameters")
		return
	} else if m.size < size {
		err = ErrRootNotAvailableYet
		return
	} else if m.store == nil {
		err = ErrHashStoreNotAvailable
		return
	}

	var (
		offset       uint64
		leftPeakHash HashType
	)
	// need no proof if size is 1
	//
	// for size > 1, we want to:
	// 1. locate the target moutain M leafIdx is in
	// 2. bag the right peaks of moutain M
	// 3. collect the preceding leaks of moutain M
	// 4. collect the proof of leafIdx within moutain M
	for size > 1 {
		// if size is not 2^n, left peak of size/size-1 is the same
		// if size is 2^n, left peak of size-1 decomposes to the left sub peak
		//
		// this trick unifies the process of finding proofs within one moutain and amoung mountains.
		//
		// it's based on the invariant that the graph can always be decomposed into a sub left mountain Msub and right side
		//
		// as long as there're no fewer than 2 leaves, whether it's completely balanced or not.
		//
		// if leafIdx is within Msub, we find the proof for Msub and bag it with the right side
		//
		// if leafIdx is out of Msub, we find the proof for the right side and bag it with the peak of Msub
		lpLeaf := leftPeakLeaf(size - 1)
		if leafIdx < lpLeaf {
			rightPeaks := getMoutainPeaks(size - lpLeaf)
			rightHashes := make([]HashType, len(rightPeaks), len(rightPeaks))
			for i := range rightPeaks {
				rightPeaks[i] += offset + 2*lpLeaf - 1
				rightHashes[i], err = m.store.GetHash(rightPeaks[i] - 1)
				if err != nil {
					return
				}
			}
			baggedRightHash := bagPeaks(m.hasher, rightHashes)
			hashes = append(hashes, baggedRightHash)
			size = lpLeaf
		} else {
			offset += 2*lpLeaf - 1
			leftPeakHash, err = m.store.GetHash(offset - 1)
			if err != nil {
				return
			}
			hashes = append(hashes, leftPeakHash)
			leafIdx -= lpLeaf
			size -= lpLeaf
		}
	}

	// reverse
	// https://github.com/golang/go/wiki/SliceTricks#reversing
	length := len(hashes)
	for i := length/2 - 1; i >= 0; i-- {
		opp := length - 1 - i
		hashes[i], hashes[opp] = hashes[opp], hashes[i]
	}

	return
}

func (m *MMR) VerifyExists(leafHash, rootHash HashType, leafIdx, size uint64, proof []HashType) (err error) {
	if m.size < size {
		err = ErrRootNotAvailableYet
		return
	}

	calculatedHash := leafHash
	lastNode := size - 1
	idx := 0
	proofLen := len(proof)

	for lastNode > 0 {
		if idx >= proofLen {
			err = fmt.Errorf("Proof too short. expected %d, got %d", proofLength(leafIdx, size), proofLen)
			return
		}

		if leafIdx%2 == 1 {
			calculatedHash = m.hasher.Node(proof[idx], calculatedHash)
			idx++
		} else if leafIdx < lastNode {
			calculatedHash = m.hasher.Node(calculatedHash, proof[idx])
			idx++
		}

		leafIdx /= 2
		lastNode /= 2
	}

	if idx < proofLen {
		err = fmt.Errorf("Proof too long")
		return
	}

	if rootHash != calculatedHash {
		err = fmt.Errorf(
			"Constructed root hash differs from provided root hash. Constructed: %x, Expected: %x",
			calculatedHash, rootHash)
		return
	}
	return
}
