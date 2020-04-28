package xorlayer

import (
	"math/bits"
	"sync"
	"unsafe"

	"github.com/zhiqiangxu/util/sort"
)

// NodeID can be replace by https://github.com/zhiqiangxu/gg
type NodeID uint64

const (
	bitSize = int(unsafe.Sizeof(NodeID(0)) * 8)
)

// XTM manages xor topology
type XTM struct {
	sync.RWMutex
	k       int
	h       int
	theta   int // for delimiting close region
	id      NodeID
	buckets []*bucket
}

// NewXTM is ctor for XTM
// caller is responsible for dealing with duplicate NodeID
func NewXTM(k, h int, id NodeID) *XTM {
	buckets := make([]*bucket, bitSize)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	x := &XTM{k: k, theta: -1, id: id, buckets: buckets}
	return x
}

func (x *XTM) AddNeighbours(ns []NodeID) {
	x.Lock()
	defer x.Unlock()

	for _, n := range ns {
		x.addNeighbourLocked(n)
	}
}

func (x *XTM) AddNeighbour(n NodeID) {
	x.Lock()
	defer x.Unlock()

	x.addNeighbourLocked(n)
}

func (x *XTM) addNeighbourLocked(n NodeID) {
	i := x.getBucketIdx(n)

	if i >= bitSize {
		return
	}

	bucket := x.buckets[i]
	exists := bucket.refresh(n)
	if exists {
		return
	}

	if i <= x.theta {
		// at most |x.k + x.h|
		if bucket.size() < (x.k + x.h) {
			bucket.insert(n)
		}
	} else {
		// no limit
		bucket.insert(n)

		// increment theta if necessary

		for {
			if x.buckets[x.theta+1].size() >= x.k {
				total := 0
				for j := x.theta + 2; j < bitSize; j++ {
					total += x.buckets[j].size()
				}
				if total >= x.k-1 {
					x.theta++
					x.buckets[x.theta].reduceTo(x.k + x.h)
				} else {
					break
				}
			} else {
				break
			}
		}

	}
}

func (x *XTM) delNeighbourLocked(n NodeID) {
	i := x.getBucketIdx(n)
	if i >= bitSize {
		return
	}

	x.buckets[i].remove(n)

	if i <= x.theta {
		// decrement theta if necessary
		if x.buckets[i].size() < x.k {
			x.theta = i - 1
		}
	}

}

// zero based index
// 0 for all NodeID that's different from id from the first bit, all with the same prefix 1 bit (2^^63)
// 1 for all NodeID that's different from id from the second bit, all with the same prefix 2 bit (2^^62)
// ...
// 63 for all NodeID that's different from id from the 64th bit, all with the same prefix 64 bit(2^^0)
// 64 means n == id
func (x *XTM) getBucketIdx(n NodeID) int {
	xor := uint64(n ^ x.id)
	lz := bits.LeadingZeros64(xor)

	return lz
}

func (x *XTM) DelNeighbours(ns []NodeID) {
	x.Lock()
	defer x.Unlock()

	for _, n := range ns {
		x.delNeighbourLocked(n)
	}
}

func (x *XTM) DelNeighbour(n NodeID) {
	x.Lock()
	defer x.Unlock()

	x.delNeighbourLocked(n)
}

func (x *XTM) NeighbourCount() (total int) {
	x.RLock()
	defer x.RUnlock()

	for _, bucket := range x.buckets {
		total += bucket.size()
	}
	return total
}

// KClosest returns k-closest nodes to target
func (x *XTM) KClosest(target NodeID) (ns []NodeID) {
	x.RLock()
	defer x.RUnlock()

	ns = make([]NodeID, 0, x.k)

	i := x.getBucketIdx(target)
	if i >= bitSize {
		ns = append(ns, x.id)
		remain := x.k - 1
		for j := bitSize - 1; j >= 0; j-- {
			bucket := x.buckets[j]
			ns = bucket.appendXClosest(ns, remain, target)
			remain = x.k - len(ns)
			if remain == 0 {
				return
			}
		}
	}

	ns = x.buckets[i].appendXClosest(ns, x.k, target)
	remain := x.k - len(ns)
	if remain == 0 {
		return
	}

	// search i+1, i+2, ... etc
	var right []NodeID
	for j := i + 1; j < bitSize; j++ {
		right = x.buckets[i].appendAll(right)
	}
	right = append(right, x.id)

	kclosest := sort.KSmallest(right, remain, func(j, k int) int {
		dj := right[j] ^ target
		dk := right[k] ^ target
		switch {
		case dj < dk:
			return -1
		case dj > dk:
			return 1
		}

		return 0
	}).([]NodeID)
	for _, n := range kclosest {
		ns = append(ns, n)
	}

	remain = x.k - len(ns)
	if remain == 0 {
		return
	}

	// search i-1, i-2, ... etc
	for j := i - 1; j >= 0; j-- {
		bucket := x.buckets[j]
		ns = bucket.appendXClosest(ns, remain, target)
		remain = x.k - len(ns)
		if remain == 0 {
			return
		}
	}

	return
}
