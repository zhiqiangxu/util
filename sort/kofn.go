package sort

import "reflect"

// KSmallest for k smallest
// mutates ns
// not sorted
func KSmallest(slice interface{}, k int, cmp func(j, k int) int) interface{} {

	v := reflect.ValueOf(slice)
	i := v.Len() / 2

	pos := PartitionLT(slice, i, cmp) + 1
	if pos == k {
		return v.Slice(0, k).Interface()
	}

	if pos < k {
		return reflect.AppendSlice(
			v.Slice(0, pos),
			reflect.ValueOf(KLargest(v.Slice(pos+1, v.Len()).Interface(), k-pos, cmp)),
		).Interface()
	}

	// pos > k

	return KLargest(v.Slice(0, pos-1).Interface(), k, cmp)
}

// KLargest for k largest
// mutates ns
func KLargest(slice interface{}, k int, cmp func(j, k int) int) interface{} {
	v := reflect.ValueOf(slice)
	i := v.Len() / 2

	pos := PartitionGT(slice, i, cmp) + 1
	if pos == k {
		return v.Slice(0, k).Interface()
	}

	if pos < k {
		return reflect.AppendSlice(
			v.Slice(0, pos),
			reflect.ValueOf(KLargest(v.Slice(pos+1, v.Len()).Interface(), k-pos, cmp)),
		).Interface()
	}

	// pos > k

	return KLargest(v.Slice(0, pos-1).Interface(), k, cmp)
}

// PartitionLT partitions array by i-th element
// mutates ns so that all values less than i-th element are on the left
// assume values are distinct
// returns the pos of i-th element
func PartitionLT(slice interface{}, i int, cmp func(j, k int) int) (pos int) {
	v := reflect.ValueOf(slice)
	swp := reflect.Swapper(slice)

	size := v.Len()
	for j := 0; j < size; j++ {
		if cmp(j, i) < 0 {
			pos++
		}
	}

	if i != pos {
		swp(i, pos)
	}

	ri := pos + 1
	if ri == size {
		return
	}
	for li := 0; li < pos; li++ {
		if cmp(li, pos) > 0 {
			for {
				if cmp(ri, pos) < 0 {
					swp(li, ri)
					ri++
					break
				} else {
					ri++
				}
			}
			if ri == size {
				return
			}
		}

	}

	return
}

// PartitionGT places larger ones on the left
func PartitionGT(slice interface{}, i int, cmp func(j, k int) int) (pos int) {
	v := reflect.ValueOf(slice)
	swp := reflect.Swapper(slice)

	size := v.Len()
	for j := 0; j < size; j++ {
		if cmp(j, i) > 0 {
			pos++
		}
	}

	if i != pos {
		swp(i, pos)
	}

	ri := pos + 1
	if ri == size {
		return
	}
	for li := 0; li < pos; li++ {
		if cmp(li, pos) < 0 {
			for {
				if cmp(ri, pos) > 0 {
					swp(li, ri)
					ri++
					break
				} else {
					ri++
				}
			}
			if ri == size {
				return
			}
		}

	}

	return
}
