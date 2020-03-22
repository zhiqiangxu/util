package rpheap

type node struct {
	item       int64
	left, next *node // 对于root，next用于循环链表，指向下一棵半树；对于非root，next指向右儿子
	rank       int
}

// Heap for rank pairing heap
// zero value is an empty heap
type Heap struct {
	head *node // 指向循环链表的头部
	size int
}

// New is ctor for Heap
func New() *Heap {
	return &Heap{}
}

type heapInterface interface {
	Insert(val int64)
	FindMin() int64
	// 传入的heap将清零
	Meld(*Heap)
	DeleteMin() int64
	Size() int
	Clear()
}

var _ heapInterface = (*Heap)(nil)

// Insert into the heap
func (h *Heap) Insert(val int64) {
	ptr := &node{item: val}
	h.insertRoot(ptr)
	h.size++
}

func (h *Heap) insertRoot(ptr *node /*ptr 是根节点*/) {
	if h.head == nil { // 第一棵半树
		h.head = ptr
		ptr.next = ptr // 循环链表
	} else {
		// 先把ptr串到header后面
		ptr.next = h.head.next
		h.head.next = ptr
		// 如果ptr更小，将header指向ptr即可，这里体现了循环链表的灵活性
		if ptr.item < h.head.item {
			h.head = ptr
		}
	}
}

// FindMin from the heap
func (h *Heap) FindMin() int64 {
	if h.head == nil {
		panic("FindMin on empty heap")
	}
	return h.head.item
}

// Meld two heaps
func (h *Heap) Meld(a *Heap) {
	if h.head == nil {
		h.head = a.head
		h.size = a.size
		a.Clear()
		return
	}
	if a.head == nil {
		return
	}

	// 两个heap都非空
	if h.head.item < a.head.item {
		merge(h, a)
	} else {
		merge(a, h)
		h.head = a.head
		h.size = a.size
	}

	a.Clear()
	return
}

func merge(a, b *Heap) {
	// 前置条件是a、b非空
	// 将b往a合并， 循环链表需要对长度为1的情况特殊处理下
	if a.size == 1 {
		ptr := a.head
		ptr.next = nil
		if ptr.left != nil || ptr.rank != 0 {
			panic("size 1 heap with left or non-zero rank")
		}
		b.insertRoot(ptr)
		b.size++
		a.head = b.head
		a.size = b.size
		return
	} else if b.size == 1 {
		ptr := b.head
		ptr.next = nil
		if ptr.left != nil || ptr.rank != 0 {
			panic("size 1 heap with left or non-zero rank")
		}
		a.insertRoot(ptr)
		a.size++
		return
	}

	// 两个链表长度都大于1
	// 将b的第二个元素串到a的第一个元素后面
	// 同时将a原先的第二个元素串到b的第一个元素后面
	// 这样b的第二个元素最终会回到b的第一个元素，然后再到a原先的第二个元素，直到回到a的头部
	// 便把两个链表合并了
	a.head.next, b.head.next = b.head.next, a.head.next
	a.size += b.size

}

// DeleteMin from the heap
func (h *Heap) DeleteMin() int64 {
	if h.head == nil {
		panic("DeleteMin on empty heap")
	}

	h.size--
	ret := h.head.item

	bucket := make([]*node, h.maxBucketSize())
	// 沿着root的唯一左child一路往右（包括左child本身），将root所在的树分解成若干棵半树，并对相同rank的树进行合并
	for ptr := h.head.left; ptr != nil; {
		// 先记录下右child，因为ptr要作为独立半树进行合并，next必须置为空
		nextPtr := ptr.next
		ptr.next = nil
		mergeSubTreeByRank(bucket, ptr)
		ptr = nextPtr
	}
	// 遍历所有半树串成的循环链表，首棵树除外
	for ptr := h.head.next; ptr != h.head; {
		nextPtr := ptr.next
		ptr.next = nil
		mergeSubTreeByRank(bucket, ptr)
		ptr = nextPtr
	}

	// 将bucket中的所有半树插入到空heap
	h.head = nil
	for _, ptr := range bucket {
		if ptr != nil {
			h.insertRoot(ptr)
		}
	}
	return ret
}

func mergeSubTreeByRank(bucket []*node, ptr *node /*ptr 是根节点*/) {
	for bucket[ptr.rank] != nil {
		rank := ptr.rank
		ptr = link(ptr, bucket[rank])
		bucket[rank] = nil
	}
	bucket[ptr.rank] = ptr
}

func link(a *node, b *node) *node {
	// 前置条件是a、b都非空
	var winner, loser *node
	// 以a、b中较小的根作为合并后的半树的根
	if a.item < b.item {
		winner = a
		loser = b
	} else {
		winner = b
		loser = a
	}

	// 将胜出半树的左child挂到胜败半树的右child上（此前右child为空）
	loser.next = winner.left
	winner.left = loser
	// 在此之前a、b的rank已经保证相等
	winner.rank++
	return winner
}

func (h *Heap) maxBucketSize() int {
	// 对于一n个元素的rpheap，rank最多是log n
	// 对于type 1或type 2的rpheap，rank上限更低，这里简单起见按log n算
	bit, cnt := 0, h.size
	for cnt > 1 {
		cnt /= 2
		bit++
	}
	return bit + 1 // [0, rank]共有rank+1个值
}

// Size of the heap
func (h *Heap) Size() int {
	return h.size
}

// Clear the heap
func (h *Heap) Clear() {
	h.head = nil
	h.size = 0
}
