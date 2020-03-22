package skl

type sklIter struct {
	s    *skl
	node *element
}

func (it *sklIter) First() (ok bool) {
	nele := it.s.links[0].next
	if nele != nil {
		ok = true
		it.node = nele
	}
	return
}

func (it *sklIter) SeekGE(key int64) (ok bool) {

	prevs := it.s.getPrevLinks(key)

	ele := prevs[0].next
	if ele != nil {
		ok = true
		it.node = ele
	}

	return
}

func (it *sklIter) Next() (ok bool) {
	ele := it.node.links[0].next
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
