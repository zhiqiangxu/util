package math

import "math/big"

func Partition(n int) *big.Int {
	if n <= 0 {
		panic("invalid input")
	}

	// Euler's method
	var current, next *Polynomial
	for i := 1; i <= n; i++ {
		coefficients := make([]*big.Int, n+1)
		for j := 0; j <= n; j++ {
			if j%i == 0 {
				coefficients[j] = big.NewInt(1)
			} else {
				coefficients[j] = big.NewInt(0)
			}
		}

		next = NewPolynomialWithMaxOrder(n, coefficients)
		if current == nil {
			current = next
		} else {
			current = current.Mul(next)
		}
	}

	return current.coefficients[n]
}
