package skl

import (
	"testing"

	"gotest.tools/assert"
)

func TestSKL(t *testing.T) {
	// test SkipListIterator
	skl := NewSkipList()
	{
		it := skl.NewIterator()
		ok := it.First()
		assert.Assert(t, !ok)
	}
	total := 10
	for j := 0; j < 2; j++ {
		for i := 0; i < total; i++ {
			skl.Add(int64(i), i)
		}
	}

	assert.Assert(t, skl.Length() == total)

	it := skl.NewIterator()
	ok := it.First()
	assert.Assert(t, ok)
	for i := 0; i < total; i++ {
		assert.Assert(t, it.Valid())
		k, v := it.KeyValue()
		assert.Assert(t, k == int64(i) && v == i)
		if i == total-1 {
			assert.Assert(t, !it.Next())
		} else {
			assert.Assert(t, it.Next())
		}
	}

	for j := 1; j < total-1; j++ {
		ok = it.SeekGE(int64(j))
		assert.Assert(t, ok)
		for i := j; i < total; i++ {
			assert.Assert(t, it.Valid())
			k, v := it.KeyValue()
			assert.Assert(t, k == int64(i) && v == i)
			if i == total-1 {
				assert.Assert(t, !it.Next())
			} else {
				assert.Assert(t, it.Next())
			}
		}
	}

	ok = it.SeekGE(int64(total + 1))
	assert.Assert(t, !ok)
}

func BenchmarkSKL(b *testing.B) {
	skl := NewSkipList()
	for i := 0; i < b.N; i++ {
		i64 := int64(i)
		skl.Add(i64, i)
		skl.Get(i64)
	}
}

func BenchmarkMap(b *testing.B) {
	m := make(map[int64]interface{})
	for i := 0; i < b.N; i++ {
		i64 := int64(i)
		m[i64] = i
		_ = m[i64]
	}
}
