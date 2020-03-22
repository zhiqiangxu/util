package skl

// SkipList for skl interface
// TODO extend key type when go supports generic
type SkipList interface {
	Add(key int64, value interface{})
	Get(key int64) (value interface{}, ok bool)
	Remove(key int64)
	Head() (key int64, value interface{}, ok bool)
	NewIterator() SkipListIterator
}

// SkipListIterator for skl iterator
type SkipListIterator interface {
	SeekGE(key int64) (ok bool)
	First() (ok bool)
	Next() (ok bool)
	Valid() bool
	Key() int64
	KeyValue() (int64, interface{})
}
