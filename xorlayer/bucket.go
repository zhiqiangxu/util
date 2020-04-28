package xorlayer

import (
	"container/list"

	"github.com/zhiqiangxu/util/sort"
)

type bucket struct {
	m map[NodeID]*list.Element
	l *list.List
}

type nodeWithCookie struct {
	N      NodeID
	Cookie uint64
}

func newBucket() *bucket {
	return &bucket{m: make(map[NodeID]*list.Element), l: list.New()}
}

func (b *bucket) insert(n NodeID, cookie uint64) {
	e := b.m[n]
	if e != nil {
		e.Value.(*nodeWithCookie).Cookie = cookie
		b.l.MoveToFront(e)
	} else {
		nwc := &nodeWithCookie{
			N:      n,
			Cookie: cookie,
		}
		e = b.l.PushFront(nwc)
		b.m[n] = e
	}
}

func (b *bucket) remove(n NodeID, cookie uint64) (removed bool) {
	e := b.m[n]
	if e != nil {
		if e.Value.(*nodeWithCookie).Cookie != cookie {
			return
		}
		removed = true
		b.l.Remove(e)
		delete(b.m, n)
	}
	return
}

func (b *bucket) reduceTo(max int) {
	for b.size() > max {
		e := b.l.Back()
		nwc := e.Value.(*nodeWithCookie)
		delete(b.m, nwc.N)
		b.l.Remove(e)
	}
}

func (b *bucket) refresh(n NodeID) (exists bool) {
	e := b.m[n]
	if e != nil {
		b.l.MoveToFront(e)
		exists = true
	}
	return
}

func (b *bucket) size() int {
	return len(b.m)
}

func (b *bucket) appendXClosest(r []NodeID, x int, target NodeID) []NodeID {
	if b.size() <= x {
		return b.appendAll(r)
	}

	// find k closest NodeID to target
	all := make([]NodeID, 0, b.size())
	for n := range b.m {
		all = append(all, n)
	}

	kclosest := sort.KSmallest(all, x, func(j, k int) int {
		dj := all[j] ^ target
		dk := all[k] ^ target
		switch {
		case dj < dk:
			return -1
		case dj > dk:
			return 1
		}

		return 0
	}).([]NodeID)

	for _, n := range kclosest {
		r = append(r, n)
	}
	return r
}

func (b *bucket) appendAll(r []NodeID) []NodeID {
	for n := range b.m {
		r = append(r, n)
	}
	return r
}

func (b *bucket) oldest() nodeWithCookie {
	return *b.l.Back().Value.(*nodeWithCookie)
}
