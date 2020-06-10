package mmr

import (
	"errors"
	"fmt"
)

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
	if hasher == nil {
		hasher = defaultHasher
	}
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

// Push a leaf
func (m *MMR) Push(leaf []byte, wantAP bool) []HashType {
	h := m.hasher.Leaf(leaf)
	return m.PushHash(h, wantAP)
}

// Push a hash
func (m *MMR) PushHash(h HashType, wantAP bool) (ap []HashType) {
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

// InclusionProof returns the audit path of ti wrt size
func (m *MMR) InclusionProof(leafIdx, size uint64) (hashes []HashType, err error) {
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
		lpLeaf := leftPeakLeaf(size - 1) // -1 for a proper one
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

func (m *MMR) VerifyInclusion(leafHash, rootHash HashType, leafIdx, size uint64, proof []HashType) (err error) {
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

// FYI: https://tools.ietf.org/id/draft-ietf-trans-rfc6962-bis-27.html#rfc.section.2.1.4
func (m *MMR) ConsistencyProof(l, n uint64) (hashes []HashType, err error) {
	if m.store == nil {
		err = ErrHashStoreNotAvailable
		return
	}

	hashes, err = m.subproof(l, n, true)
	return
}

func (m *MMR) subproof(l, n uint64, compeleteST bool) (hashes []HashType, err error) {

	var hash HashType
	offset := uint64(0)
	for l < n {
		k := leftPeakLeaf(n - 1)
		if l <= k {
			rightPeaks := getMoutainPeaks(n - k)
			rightHashes := make([]HashType, len(rightPeaks), len(rightPeaks))
			for i := range rightPeaks {
				rightPeaks[i] = offset + 2*k - 1
				rightHashes[i], err = m.store.GetHash(rightPeaks[i] - 1)
				if err != nil {
					return
				}
			}
			baggedRightHash := bagPeaks(m.hasher, rightHashes)
			hashes = append(hashes, baggedRightHash)
			n = k
		} else {
			offset += k*2 - 1
			hash, err = m.store.GetHash(offset - 1)
			if err != nil {
				return
			}
			hashes = append(hashes, hash)
			l -= k
			n -= k
			compeleteST = false
		}
	}

	if !compeleteST {
		peaks := getMoutainPeaks(l)
		if len(peaks) != 1 {
			panic("bug in subproof")
		}
		hash, err = m.store.GetHash(peaks[0] + offset - 1)
		if err != nil {
			return
		}
		hashes = append(hashes, hash)
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

func (m *MMR) VerifyConsistency(oldTreeSize, newTreeSize uint64, oldRoot, newRoot HashType, proof []HashType) (err error) {
	if oldTreeSize > newTreeSize {
		err = fmt.Errorf("oldTreeSize > newTreeSize")
		return
	}

	if oldTreeSize == newTreeSize {
		return
	}

	if oldTreeSize == 0 {
		return
	}

	first := oldTreeSize - 1
	last := newTreeSize - 1

	for first%2 == 1 {
		first /= 2
		last /= 2
	}

	lenp := len(proof)
	if lenp == 0 {
		err = errors.New("Wrong proof length")
		return
	}

	pos := 0
	var newHash, oldHash HashType

	if first != 0 {
		newHash = proof[pos]
		oldHash = proof[pos]
		pos += 1
	} else {
		newHash = oldRoot
		oldHash = oldRoot
	}

	for first != 0 {
		if first%2 == 1 {
			if pos >= lenp {
				err = errors.New("Wrong proof length")
				return
			}
			// node is a right child: left sibling exists in both trees
			nextNode := proof[pos]
			pos += 1
			oldHash = m.hasher.Node(nextNode, oldHash)
			newHash = m.hasher.Node(nextNode, newHash)
		} else if first < last {
			if pos >= lenp {
				err = errors.New("Wrong proof length")
				return
			}
			// node is a left child: right sibling only exists in the newer tree
			nextNode := proof[pos]
			pos += 1
			newHash = m.hasher.Node(nextNode, newHash)
		}

		first /= 2
		last /= 2
	}

	for last != 0 {
		if pos >= lenp {
			err = errors.New("Wrong proof length")
			return
		}
		nextNode := proof[pos]
		pos += 1
		newHash = m.hasher.Node(nextNode, newHash)
		last /= 2
	}

	if newHash != newRoot {
		err = errors.New(fmt.Sprintf(`Bad Merkle proof: second root hash does not match. 
			Expected hash:%x, computed hash: %x`, newRoot, newHash))
		return
	} else if oldHash != oldRoot {
		err = errors.New(fmt.Sprintf(`Inconsistency: first root hash does not match."
			"Expected hash: %x, computed hash:%x`, oldRoot, oldHash))
		return
	}

	if pos != lenp {
		err = errors.New("Proof too long")
		return
	}

	return
}
