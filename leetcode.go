package util

import (
	"math"

	"github.com/zhiqiangxu/qrpc"
)

// Max for int
func Max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// Min for int
func Min(a, b int) int {
	if a > b {
		return b
	}

	return a
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

// ManacherFallback when all byte values are used in s
func ManacherFallback(s string) (ss string) {
	n := len(s)

	for i := 0; i < n; i++ {
		// 奇数的情况
		j := 1
		for ; i-j >= 0 && i+j < n; j++ {
			if s[i-j] != s[i+j] {
				break
			}
		}
		if len(ss) < (j*2 - 1) {
			ss = s[i-j+1 : i+j]
		}
		// 偶数的情况
		if i+1 < n && s[i] == s[i+1] {
			j := 1
			for ; i-j >= 0 && i+1+j < n; j++ {
				if s[i-j] != s[i+1+j] {
					break
				}
			}
			if len(ss) < j*2 {
				ss = s[i-j+1 : i+j+1]
			}
		}
	}
	return
}

// ManacherWithFallback tries Manacher if possible
func ManacherWithFallback(s string) (ss string) {
	var indexes [math.MaxUint8] /*byte是uint8*/ bool
	n := len(s)
	for i := 0; i < n; i++ {
		indexes[s[i]] = true
	}
	canManacher := false
	var manacherByte byte
	for i, exists := range indexes {
		if !exists {
			canManacher = true
			manacherByte = byte(i)
			break
		}
	}
	if !canManacher {
		ss = ManacherFallback(s)
		return
	}

	// preprocess
	bytes := make([]byte, 2*n+1, 2*n+1)
	bytes[0] = manacherByte
	for i := 0; i < n; i++ {
		bytes[2*i+1] = s[i]
		bytes[2*i+2] = manacherByte
	}

	r := make([]int, 2*n+1)
	var maxRightPos, maxRight, maxRPos, maxR int
	r[0] = 1
	r[2*n] = 1
	for i := 1; i < 2*n; i++ {
		if i >= maxRight {
			// 半径包括自己，所以1是最小值
			r[i] = 1
		} else {
			// i在maxRight以内
			// j'坐标为2*maxRightPos-i
			r[i] = Min(maxRight-i, r[2*maxRightPos-i])
		}
		// 尝试扩大半径
		for {
			if i-r[i] >= 0 && i+r[i] <= 2*n && bytes[i-r[i]] == bytes[i+r[i]] {
				r[i]++
			} else {
				break
			}
		}
		if i+r[i]-1 > maxRight {
			maxRight = i + r[i] - 1
			maxRightPos = i
		}
		if maxR < r[i] {
			maxRPos = i
			maxR = r[i]
		}
	}

	targetBytes := make([]byte, 0, r[maxRPos]-1 /*最终结果的长度*/)
	for i := maxRPos - r[maxRPos] + 1; i < maxRPos+r[maxRPos]; i++ {
		if bytes[i] != manacherByte {
			targetBytes = append(targetBytes, bytes[i])
		}
	}

	ss = qrpc.String(targetBytes)
	if len(ss) != r[maxRPos]-1 {
		panic("size != r[maxRPos]-1")
	}

	return
}

// ReverseDigits for reverse digits
func ReverseDigits(n int32) (r int32) {

	if n > 0 {
		for n != 0 {
			pop := n % 10
			// 溢出判断
			if (r > math.MaxInt32/10) || (r == math.MaxInt32/10 && pop > 7) {
				// 上溢出
				return 0
			}

			r = 10*r + pop
			n /= 10
		}
	} else {
		for n != 0 {
			pop := n % 10
			// 溢出判断
			if r < math.MinInt32/10 || (r == math.MinInt32/10 && pop < -8) {
				// 下溢出
				return 0
			}

			r = 10*r + pop
			n /= 10
		}
	}

	return
}

// IsPalindrome checks whether n is palindrome
func IsPalindrome(n int) bool {
	if n < 0 {
		return false
	}

	var reverted int
	// 反转一半即可
	for n > reverted {
		reverted = 10*reverted + n%10
		n /= 10
	}

	return reverted == n || reverted/10 == n
}
