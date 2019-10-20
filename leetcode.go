package util

import "math"

// Max for int
func Max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// MaxLengthOfUniqueSubstring returns the max length of unique sub string
func MaxLengthOfUniqueSubstring(s string) (l int) {

	var indexes [math.MaxUint8] /*byte是uint8*/ int
	n := len(s)

	var i, j int
	// 基于的观察：
	// 如果s[j]跟[i,j)有重复j'，那么可以跳过[i,j']的元素，i直接变为j'+1
	for ; j < n; j++ {
		byteJ := s[j]
		// 假如indexes的元素非0，那么必定是这个byte上一次出现的位置j'+1
		// 且j'的位置必须在[i,j)之间才有效
		if indexes[byteJ] != 0 && indexes[byteJ] > i {
			i = indexes[byteJ]
		}

		l = Max(l, j-i+1)
		indexes[byteJ] = j + 1
	}
	return
}
