package wnaf

import (
	"math/big"
	"reflect"
	"testing"
)

func TestToWnafRL(t *testing.T) {
	testCases := []struct {
		D    *big.Int
		Wnaf []int64
		W    uint8
	}{
		// quoted from https://rd.springer.com/content/pdf/10.1007/978-3-540-68914-0_26.pdf Page 11
		{
			D:    big.NewInt(0b11101001100100010101110101010111),
			Wnaf: []int64{7, 0, 0, 0, 5, 0, 0, 0, -3, 0, 0, 0, 0, -5, 0, 0, 0, -7, 0, 0, 0, -3, 0, 0, 0, 5, 0, 0, 0, 7},
			W:    4,
		},
		{
			D:    big.NewInt(0b11101001100100010101110101010111),
			Wnaf: []int64{-1, 0, 0, 3, 0, 0, -3, 0, 0, -1, 0, 0, 0, 3, 0, 0, 1, 0, 0, 0, 1, 0, 0, 3, 0, 0, 0, -3, 0, 0, 0, 0, 1},
			W:    3,
		},
		{
			D:    big.NewInt(0),
			Wnaf: []int64{0},
			W:    8,
		},
		{
			D:    big.NewInt(1 << 8),
			Wnaf: []int64{0, 0, 0, 0, 0, 0, 0, 0, 1},
			W:    8,
		},
	}

	for _, testCase := range testCases {
		ret := ToWnafRL(testCase.D, testCase.W)
		if !reflect.DeepEqual(ret, testCase.Wnaf) {
			t.Fatalf("actual %v vs expected %v, w %d", ret, testCase.Wnaf, testCase.W)
		}
	}

}
