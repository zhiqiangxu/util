package math

import (
	"fmt"
	"math/big"
	"strings"
)

// Polynomial is used for educational purpose only, not optimized for sparse polynomials
type Polynomial struct {
	maxOrder     int
	coefficients []*big.Int
}

func NewPolynomialWithMaxOrder(maxOrder int, coefficients []*big.Int) *Polynomial {
	return &Polynomial{maxOrder: maxOrder, coefficients: coefficients}
}

func NewPolynomial(coefficients []*big.Int) *Polynomial {
	return &Polynomial{coefficients: coefficients}
}

var zero = big.NewInt(0)
var one = big.NewInt(1)

func (p1 *Polynomial) String() string {
	var parts []string
	for i, c := range p1.coefficients {
		if c.Cmp(zero) != 0 {
			if i == 0 {
				parts = append(parts, c.String())
			} else {
				if c.Cmp(one) == 0 {
					parts = append(parts, fmt.Sprintf("x^%d", i))
				} else {
					parts = append(parts, fmt.Sprintf("%sx^%d", c.String(), i))
				}
			}
		}
	}

	return strings.Join(parts, " + ")
}

func (p1 *Polynomial) Mul(p2 *Polynomial) *Polynomial {
	dim := len(p1.coefficients) + len(p2.coefficients) - 1
	if dim <= 0 {
		return &Polynomial{}
	}

	if p1.maxOrder > 0 && dim > p1.maxOrder+1 {
		dim = p1.maxOrder + 1
	}
	coefficients := make([]*big.Int, dim)
	for i, c1 := range p1.coefficients {
		for j, c2 := range p2.coefficients {
			if p1.maxOrder > 0 && p1.maxOrder < i+j {
				break
			}
			c3 := coefficients[i+j]
			if c3 == nil {
				c3 = big.NewInt(0)
			}

			coefficients[i+j] = c3.Add(c3, big.NewInt(0).Mul(c1, c2))
		}
	}

	return &Polynomial{maxOrder: p1.maxOrder, coefficients: coefficients}
}
