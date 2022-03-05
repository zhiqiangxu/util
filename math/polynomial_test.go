package math

import (
	"math/big"
	"testing"

	"gotest.tools/assert"
)

func intArray2BigArray(coefficients []int) []*big.Int {
	result := make([]*big.Int, len(coefficients))
	for i, c := range coefficients {
		result[i] = big.NewInt(int64(c))
	}
	return result
}

func TestPolynomial(t *testing.T) {

	// euler formula to calculate Partition(5)
	p1 := NewPolynomial(intArray2BigArray([]int{1, 1, 1, 1, 1, 1}))
	p2 := NewPolynomial(intArray2BigArray([]int{1, 0, 1, 0, 1}))
	p3 := NewPolynomial(intArray2BigArray([]int{1, 0, 0, 1}))
	p4 := NewPolynomial(intArray2BigArray([]int{1, 0, 0, 0, 1}))
	p5 := NewPolynomial(intArray2BigArray([]int{1, 0, 0, 0, 0, 1}))

	p := p1.Mul(p2).Mul(p3).Mul(p4).Mul(p5)
	assert.Equal(t, p.coefficients[5].Uint64(), uint64(7))
	t.Log(p)
}
