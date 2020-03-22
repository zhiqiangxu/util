package skl

type sklIter struct {
	s    *skl
	node *element
}

func (it *sklIter) First() (ok bool) {
	if it.s.next[0] != nil {
		ok = true
		it.node = it.s.next[0]
	}
	return
}

func (it *sklIter) SeekGE(key int64) (ok bool) {

	prevs := it.s.getPrevLinks(key)

	ele := prevs[0].next[0]
	if ele != nil {
		ok = true
		it.node = ele
	}

	return
}

func (it *sklIter) Next() (ok bool) {
	ele := it.node.next[0]
	if ele != nil {
		ok = true
		it.node = ele
	}
	return
}

func (it *sklIter) Valid() bool {
	return it.node != nil
}

func (it *sklIter) Key() int64 {
	return it.node.key
}

func (it *sklIter) KeyValue() (int64, interface{}) {
	return it.node.key, it.node.value
}
