package diskqueue

type qfileIteratorInterface interface {
	SeekToOffset(offset int64)
	SeekToTime(ts int64)
	Next()
	CurrentOffset() int64
	Data() []byte
	Valid() bool
}

// qfileIterator for iterate over qfile
type qfileIterator struct {
	qf  *qfile
	cur int64
}

var _ qfileIteratorInterface = (*qfileIterator)(nil)

func (it *qfileIterator) SeekToOffset(offset int64) {
	it.cur = offset
}

func (it *qfileIterator) SeekToTime(ts int64) {
	panic("not impl yet")
}

func (it *qfileIterator) Next() {

}

func (it *qfileIterator) CurrentOffset() int64 {
	return it.cur
}

func (it *qfileIterator) Data() []byte {
	return nil
}

func (it *qfileIterator) Valid() bool {
	return it.qf.IsInRange(it.cur + 4)
}
