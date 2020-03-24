package lf

import (
	"bytes"
	"sync/atomic"
	"unsafe"
)

// List is a lock free sorted singly linked list
type List struct {
	head  listNode
	arena *Arena
}

var _ list = (*List)(nil)

// NewListWithArena with specified Arena
func NewListWithArena(arena *Arena) *List {
	l := &List{arena: arena}

	l.head.head = true
	return l
}

// NewList with arenaSize
func NewList(arenaSize uint32) *List {
	arena := NewArena(arenaSize)
	return NewListWithArena(arena)
}

// ListNodeSize is the size of ListNode
const ListNodeSize = int(unsafe.Sizeof(listNode{}))

const (
	bitMask = ^uint32(0x03)
	markBit = uint32(1)
	flagBit = uint32(2)
)

type listNode struct {
	// Multiple parts of the value are encoded as a single uint64 so that it
	// can be atomically loaded and stored:
	//   value offset: uint32 (bits 0-31)
	//   value size  : uint16 (bits 32-63)
	value     uint64
	backlink  uint32 // points to the prev node
	succ      uint32 // contains a right pointer, a mark bit and a flag bit.
	keyOffset uint32 // Immutable. No need to lock to access key.
	keySize   uint16 // Immutable. No need to lock to access key.
	head      bool
}

// Contains checks whether k is in list
func (l *List) Contains(k []byte) bool {
	current, _ := l.searchFrom(k, l.headNode(), true)
	return current.Compare(l.arena, k) == 0
}

// Get v by k if exists
// v is readonly
func (l *List) Get(k []byte) (v []byte, exists bool) {
	current, _ := l.searchFrom(k, l.headNode(), true)
	if current.Compare(l.arena, k) == 0 {
		exists = true
		v = current.Value(l.arena)
	}
	return
}

func (l *List) headNode() *listNode {
	return &l.head
}

// Insert attempts to insert a new node with the supplied key and value.
func (l *List) Insert(k, v []byte) (isNew bool, err error) {
	prev, next := l.searchFrom(k, l.headNode(), true)

	var voffset uint32

	if prev.Compare(l.arena, k) == 0 {

		voffset, err = l.arena.putBytes(v)
		if err != nil {
			return
		}
		prev.UpdateValue(voffset, uint16(len(v)))
		return
	}

	node, err := newListNode(l.arena, k, v)
	if err != nil {
		return
	}
	nodeOffset := l.arena.getListNodeOffset(node)

	for {
		prevSucc := prev.Succ()
		// If the predecessor is flagged, help
		// the corresponding deletion to complete.
		if prevSucc&flagBit != 0 {
			l.helpFlagged(prev, l.arena.getListNode(prevSucc&(^markBit)))
		} else {
			node.succ = l.arena.getListNodeOffset(next)
			// Insertion attempt.
			if atomic.CompareAndSwapUint32(&prev.succ, node.succ, nodeOffset) {
				// Successful insertion.
				isNew = true
				return
			}

			// Failure.

			// Failure due to flagging.
			if prev.Flagged() {
				l.helpFlagged(prev, prev.Next(l.arena))
			}
			// Possibly a failure due to marking. Traverse a
			// chain of backlinks to reach an unmarked node.
			for prev.Marked() {
				prev = l.arena.getListNode(prev.backlink)
			}
		}

		prev, next = l.searchFrom(k, prev, true)
		if prev != nil && prev.Compare(l.arena, k) == 0 {
			prev.UpdateValue(voffset, uint16(len(v)))
			return
		}
	}
}

// Delete sttempts to delete a node with the supplied key
func (l *List) Delete(k []byte) bool {
	prev, del := l.searchFrom(k, l.headNode(), false)
	if del == nil || del.Compare(l.arena, k) != 0 {
		return false
	}

	prev, flagged := l.tryFlag(prev, del)
	if prev != nil {
		l.helpFlagged(prev, del)
	}

	return flagged
}

// finds two consecutive nodes n1 and n2
// pre condition:
//		node.key < k
// if equal is true:
// 		n1.key <= k < n2.key.
// if equal is false:
// 		n1.key < k <= n2.key.
func (l *List) searchFrom(k []byte, node *listNode, equal bool) (current, next *listNode) {

	var cmpFunc func(cmp int) bool
	if equal {
		cmpFunc = func(cmp int) bool {
			return cmp <= 0
		}
	} else {
		cmpFunc = func(cmp int) bool {
			return cmp < 0
		}
	}

	current = node
	next = node.Next(l.arena)
	for next != nil && cmpFunc(next.Compare(l.arena, k)) {
		for {
			nextSuc := next.Succ()
			currentSuc := current.Succ()
			currentNext := l.arena.getListNode(currentSuc & bitMask)

			// Ensure that either next node is unmarked,
			// or both curr node and next node are
			// marked and curr node was marked earlier.
			if nextSuc&markBit == 1 && (currentSuc&markBit == 0 || currentNext != next) {
				if currentNext == next {
					l.helpMarked(current, next)
				}
				next = currentNext
			} else {
				break
			}
		}

		if next != nil && cmpFunc(next.Compare(l.arena, k)) {
			current = next
			next = current.Next(l.arena)
		}

	}
	return
}

// Attempts to physically delete the marked
// node del node and unflag prev node.
func (l *List) helpMarked(prev, del *listNode) {
	next := del.Next(l.arena)
	atomic.CompareAndSwapUint32(&prev.succ, l.arena.getListNodeOffset(del)+flagBit, l.arena.getListNodeOffset(next))
}

// Attempts to flag the predecessor of target node. P rev node is the last node known to be the predecessor.
func (l *List) tryFlag(prev, target *listNode) (n *listNode, flagged bool) {

	for {
		// predecessor is already flagged
		if prev.Flagged() {
			n = prev
			return
		}
		targetOffset := l.arena.getListNodeOffset(target)
		if atomic.CompareAndSwapUint32(&prev.succ, targetOffset, targetOffset+flagBit) {
			// c&s was successful
			n = prev
			flagged = true
			return
		}

		if prev.Flagged() {
			// failure due to flagging
			n = prev
			return
		}

		// possibly failure due to marking
		for prev.Marked() {
			prev = l.arena.getListNode(prev.backlink)
		}

		var del *listNode
		prev, del = l.searchFrom(target.Key(l.arena), prev, false)
		// target_node was deleted from the list
		if del != target {
			return
		}
	}

}

// Attempts to mark the node del node.
func (l *List) tryMark(del *listNode) {
	for !del.Marked() {
		right := del.Succ() & (^markBit)
		swapped := atomic.CompareAndSwapUint32(&del.succ, right, right+markBit)
		if !swapped {
			if atomic.LoadUint32(&del.succ)&flagBit != 0 {
				l.helpFlagged(del, l.arena.getListNode(right))
			}
		}
	}
}

// Attempts to mark and physically delete node del node,
// which is the successor of the flagged node prev node.
func (l *List) helpFlagged(prev, del *listNode) {
	del.backlink = l.arena.getListNodeOffset(prev)

	l.tryMark(del)
	l.helpMarked(prev, del)
}

func newListNode(arena *Arena, k, v []byte) (n *listNode, err error) {
	koff, voff, err := arena.putKV(k, v)
	if err != nil {
		return
	}
	noff, err := arena.putListNode()
	if err != nil {
		return
	}
	n = arena.getListNode(noff)
	n.keyOffset = koff
	n.keySize = uint16(len(k))
	n.value = encodeValue(voff, uint16(len(v)))
	return
}

func encodeValue(valOffset uint32, valSize uint16) uint64 {
	return uint64(valSize)<<32 | uint64(valOffset)
}

func decodeValue(value uint64) (valOffset uint32, valSize uint16) {
	valSize = uint16(value >> 32)
	valOffset = uint32(value & 0xffffffff)
	return
}

func (n *listNode) Flagged() bool {
	return atomic.LoadUint32(&n.succ)&flagBit != 0
}

func (n *listNode) Marked() bool {
	return atomic.LoadUint32(&n.succ)&markBit != 0
}

func (n *listNode) Succ() uint32 {
	return atomic.LoadUint32(&n.succ)
}

func (n *listNode) Key(arena *Arena) []byte {
	return arena.getBytes(n.keyOffset, n.keySize)
}

func (n *listNode) Value(arena *Arena) []byte {
	v := atomic.LoadUint64(&n.value)
	voff, vsize := decodeValue(v)
	return arena.getBytes(voff, vsize)
}

func (n *listNode) UpdateValue(offset uint32, size uint16) {
	value := encodeValue(offset, size)
	atomic.StoreUint64(&n.value, value)
}

func (n *listNode) Next(arena *Arena) *listNode {
	succ := n.Succ()
	return arena.getListNode(succ & bitMask)
}

func (n *listNode) Compare(arena *Arena, k []byte) int {
	if n.head {
		return -1
	}
	return bytes.Compare(arena.getBytes(n.keyOffset, n.keySize), k)
}
