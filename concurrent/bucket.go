package concurrent

type Bucket[K any, E any] struct {
	hash     func(K) uint32
	elements []E
}

func NewBucket[K any, E any](buckets int, genFunc func() E, hash func(K) uint32) *Bucket[K, E] {
	elements := make([]E, 0, buckets)
	for i := 0; i < buckets; i++ {
		elements = append(elements, genFunc())
	}

	return &Bucket[K, E]{elements: elements, hash: hash}
}

func (b *Bucket[K, E]) Element(key K) E {
	idx := b.hash(key)
	return b.elements[int(idx)%len(b.elements)]
}
