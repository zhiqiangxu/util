package util

import (
	"testing"
)

func BenchmarkAtomicByteArena(b *testing.B) {
	a := NewAtomicByteArena(8 * 1024 * 1024)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.AllocBytes(8)
			// _ = make([]byte, 8)
		}
	})
	// for i := 0; i < b.N; i++ {
	// a.AllocBytes(8)
	// 	_ = make([]byte, 8)
	// }
}
