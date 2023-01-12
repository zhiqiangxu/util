package wnaf

import (
	"math/big"
	"reflect"
	"testing"
)

func TestToWnafLR(t *testing.T) {
	testCases := []struct {
		D    *big.Int
		Wnaf []int64
		W    uint8
	}{
		// quoted from https://rd.springer.com/content/pdf/10.1007/978-3-540-68914-0_26.pdf Page 11
		{
			D:    big.NewInt(0b11101001100100010101110101010111),
			Wnaf: []int64{0, 0, 7, 0, 0, 0, 5, 0, 0, 0, 0, -7, 0, 0, 0, 0, 0, 5, 0, 0, 0, 7, 0, 0, 0, 5, 0, 0, 3, 0, 0, 0, -2},
			W:    4,
		},
		{
			D:    big.NewInt(0b11101001100100010101110101010111),
			Wnaf: []int64{2, 0, 0, 0, -3, 0, 0, 0, 3, 0, 0, 1, 0, 0, 0, 0, 3, 0, -1, 0, 0, 0, 0, -3, 0, 0, 3, 0, -1, 0, 0, 0, -2},
			W:    3,
		},
		{
			D:    big.NewInt(0b1110),
			Wnaf: []int64{0, 0, 7, 0},
			W:    4,
		},
		{
			D:    big.NewInt(0b111),
			Wnaf: []int64{0, 0, 7},
			W:    4,
		},
	}

	for _, testCase := range testCases {
		ret := ToWnafLR(testCase.D, testCase.W)
		if !reflect.DeepEqual(ret, testCase.Wnaf) {
			t.Fatalf("actual %v vs expected %v, w %d", ret, testCase.Wnaf, testCase.W)
		}
	}

}
