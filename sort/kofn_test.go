package sort

import (
	"testing"

	"gotest.tools/assert"
)

func TestPartition(t *testing.T) {

	{
		data := [][]int{
			[]int{5, 4, 3, 2, 1},
			[]int{1, 2, 3, 4, 5},
			[]int{1, 3, 2, 5, 4},
		}

		for _, ns := range data {
			i := 3
			v := ns[i]
			cmp := func(i, j int) int {
				return ns[i] - ns[j]
			}
			pos := PartitionLT(ns, i, cmp)
			verifyLT(t, ns, v, pos)

		}
	}

	{
		data := [][]int{
			[]int{5, 4, 3, 2, 1},
			[]int{1, 2, 3, 4, 5},
			[]int{1, 3, 2, 5, 4},
		}

		for _, ns := range data {
			i := 3
			v := ns[i]
			cmp := func(i, j int) int {
				return ns[i] - ns[j]
			}
			pos := PartitionGT(ns, i, cmp)
			verifyGT(t, ns, v, pos)
		}
	}

	{
		data := [][]int{
			[]int{5, 4, 3, 2, 1},
			[]int{1, 2, 3, 4, 5},
			[]int{1, 3, 2, 5, 4},
		}

		for _, ns := range data {
			cmp := func(i, j int) int {
				return ns[i] - ns[j]
			}
			ks := KSmallest(ns, 3, cmp)
			verifyKS(t, ks.([]int), ns, 3)
		}
	}

	{
		data := [][]int{
			[]int{5, 4, 3, 2, 1},
			[]int{1, 2, 3, 4, 5},
			[]int{1, 3, 2, 5, 4},
		}

		for _, ns := range data {
			cmp := func(i, j int) int {
				return ns[i] - ns[j]
			}
			kl := KLargest(ns, 3, cmp)
			verifyKL(t, kl.([]int), ns, 3)
		}
	}

}

func verifyKL(t *testing.T, ks, ns []int, k int) {
	assert.Assert(t, len(ks) == k, "len(ks) != k")
	m := make(map[int]bool)
	for _, v := range ns {
		m[v] = true
	}

	for _, v := range ks {
		delete(m, v)
	}

	for _, v1 := range ks {
		for v2 := range m {
			assert.Assert(t, v2 < v1, "v2>v1")
		}
	}
}

func verifyKS(t *testing.T, ks, ns []int, k int) {
	assert.Assert(t, len(ks) == k, "len(ks) != k")
	m := make(map[int]bool)
	for _, v := range ns {
		m[v] = true
	}

	for _, v := range ks {
		delete(m, v)
	}

	for _, v1 := range ks {
		for v2 := range m {
			assert.Assert(t, v2 > v1, "v2<v1")
		}
	}
}

func verifyLT(t *testing.T, ns []int, v int, pos int) {
	assert.Assert(t, v == ns[pos], "v != ns[pos]")
	for i := 0; i < len(ns); i++ {
		if i < pos {
			assert.Assert(t, ns[i] < v, "left element not smaller")
		} else if i > pos {
			assert.Assert(t, ns[i] > v, "right element not larger")
		}
	}
}

func verifyGT(t *testing.T, ns []int, v int, pos int) {
	assert.Assert(t, v == ns[pos], "v != ns[pos]")
	for i := 0; i < len(ns); i++ {
		if i < pos {
			assert.Assert(t, ns[i] > v, "left element not larger")
		} else if i > pos {
			assert.Assert(t, ns[i] < v, "right element not smaller")
		}
	}
}
