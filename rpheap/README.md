# rpheap

An easy to use rank pairing heap, which takes **O(1)** for `Insert`, `FindMin`, `Meld`, and **O(log n)** for `DeleteMin`. Rank pairing heap is one of the most efficient heaps so far!

```golang

package rpheap

import (
	"sort"
	"testing"

    "gotest.tools/assert"
    "github.com/zhiqiangxu/rpheap"
)

func TestRPHeap(t *testing.T) {
	heap := rpheap.New()
	numbers := []int{10, 4, 3, 2, 5, 1}
	for _, number := range numbers {
		rpheap.Insert(int64(number))
	}

	sort.Ints(numbers)

	for _, number := range numbers {
		m := heap.DeleteMin()
		assert.Assert(t, int64(number) == m, "number:%v m:%v", number, m)
	}

	assert.Assert(t, heap.Size() == 0, "heap not empty")
}

```