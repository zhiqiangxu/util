package xorlayer

import (
	"context"
	"math/bits"
	"sync"
	"time"
	"unsafe"

	"github.com/zhiqiangxu/util/sort"
)

// NodeID can be replace by https://github.com/zhiqiangxu/gg
type NodeID uint64

const (
	bitSize = int(unsafe.Sizeof(NodeID(0)) * 8)
)

// Callback is used by XTM
type Callback interface {
	Ping(ctx context.Context, nodeID NodeID) error
}

// XTM manages xor topology
type XTM struct {
	sync.RWMutex
	k       int
	h       int
	theta   int // for delimiting close region
	id      NodeID
	buckets []*bucket
	cb      Callback
}

// NewXTM is ctor for XTM
// caller is responsible for dealing with duplicate NodeID
func NewXTM(k, h int, id NodeID, cb Callback) *XTM {
	buckets := make([]*bucket, bitSize)
	for i := range buckets {
		buckets[i] = newBucket()
	}
	x := &XTM{k: k, h: h, theta: -1, id: id, buckets: buckets, cb: cb}
	return x
}

// AddNeighbours is batch for AddNeighbour
func (x *XTM) AddNeighbours(ns []NodeID, cookies []uint64) {
	x.Lock()

	var unlocked bool

	for i, n := range ns {
		unlocked = x.addNeighbourLocked(n, cookies[i])
		if unlocked {
			unlocked = false
			x.Lock()
		}
	}

	if !unlocked {
		x.Unlock()
	}
}

// AddNeighbour tries to add NodeID with cookie to kbucket
func (x *XTM) AddNeighbour(n NodeID, cookie uint64) {
	x.Lock()

	if !x.addNeighbourLocked(n, cookie) {
		x.Unlock()
	}
}

const (
	pingTimeout = time.Second * 2
)

func (x *XTM) addNeighbourLocked(n NodeID, cookie uint64) (unlocked bool) {
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
			bucket.insert(n, cookie)
		} else {
			if x.cb != nil {
				oldest := bucket.oldest()
				x.Unlock()
				ctx, cancelFunc := context.WithTimeout(context.Background(), pingTimeout)
				defer cancelFunc()

				err := x.cb.Ping(ctx, oldest.N)
				if err != nil {
					x.Lock()
					if bucket.remove(oldest.N, oldest.Cookie) {
						bucket.insert(n, cookie)
					}
				} else {
					unlocked = true
				}
			}
		}
	} else {
		// no limit
		bucket.insert(n, cookie)

		// increment theta if necessary

		for {
			if x.buckets[x.theta+1].size() >= (x.k + x.h) {
				total := 0
				for j := x.theta + 2; j < bitSize; j++ {
					total += x.buckets[j].size()
				}
				if total >= (x.k + x.h) {
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

	return
}

func (x *XTM) delNeighbourLocked(n NodeID, cookie uint64) {
	i := x.getBucketIdx(n)
	if i >= bitSize {
		return
	}

	x.buckets[i].remove(n, cookie)

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

// DelNeighbours is batch for DelNeighbour
func (x *XTM) DelNeighbours(ns []NodeID, cookies []uint64) {
	x.Lock()
	defer x.Unlock()

	for i, n := range ns {
		x.delNeighbourLocked(n, cookies[i])
	}
}

// DelNeighbour by NodeID and cookie
func (x *XTM) DelNeighbour(n NodeID, cookie uint64) {
	x.Lock()
	defer x.Unlock()

	x.delNeighbourLocked(n, cookie)
}

// NeighbourCount returns total neighbour count
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
		return
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
