package xorlayer

import (
	"reflect"
	"testing"

	"sort"

	"math/rand"

	"gotest.tools/assert"
)

func TestXTM(t *testing.T) {
	nodeID := NodeID(1)
	k, h := 4, 1
	x := NewXTM(k, h, nodeID, nil)

	assert.Assert(t, x.getBucketIdx(nodeID) == bitSize)

	total := 10000
	cookie := uint64(0)
	neighbours := make(map[NodeID]bool)
	for i := 0; i < total; i++ {
		n := NodeID(rand.Uint64())
		if n == nodeID {
			continue
		}
		if neighbours[n] {
			continue
		}
		neighbours[n] = true
		x.AddNeighbour(n, cookie)
	}

	kclosest := x.KClosest(nodeID)
	sort.Slice(kclosest, func(i, j int) bool {
		return kclosest[i] < kclosest[j]
	})

	neighbourSlice := make([]NodeID, 0, len(neighbours))
	for n := range neighbours {
		neighbourSlice = append(neighbourSlice, n)
	}
	sort.Slice(neighbourSlice, func(i, j int) bool {
		return neighbourSlice[i]^nodeID < neighbourSlice[j]^nodeID
	})

	expected := []NodeID{nodeID}
	expected = append(expected, neighbourSlice[0:k-1]...)
	assert.Assert(t, reflect.DeepEqual(kclosest, expected))
}
