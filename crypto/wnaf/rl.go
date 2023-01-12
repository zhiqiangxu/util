package wnaf

import (
	"math/big"
)

// FYI https://en.wikipedia.org/wiki/Elliptic_curve_point_multiplication#w-ary_non-adjacent_form_(wNAF)_method
//	   https://cs.uwaterloo.ca/~shallit/Papers/bbr.pdf

var (
	zero = big.NewInt(0)
	one  = big.NewInt(1)
	two  = big.NewInt(2)
)

// ToWnafRL converts a big int to wnaf form from right to left.
// the returned wnaf "bits" are in little endian.
// each "bit" is in range [-2^(w-1), 2^(w-1))
func ToWnafRL(d *big.Int, w uint8) (wnaf []int64) {
	if d.Cmp(zero) < 0 {
		panic("negative d not supported")
	}
	if w > 64 {
		panic("w > 64 not supported")
	}

	mod := big.NewInt(0).Exp(two, big.NewInt(int64(w)), nil)
	halfMod := big.NewInt(0).Div(mod, two)

	for {
		if d.Cmp(zero) <= 0 {
			break
		}
		if big.NewInt(0).Mod(d, two).Uint64() == 1 {
			// mods
			di := big.NewInt(0).Mod(d, mod)
			if di.Cmp(halfMod) >= 0 {
				di.Sub(di, mod)
			}

			d.Sub(d, di)
			wnaf = append(wnaf, di.Int64())
		} else {
			wnaf = append(wnaf, 0)
		}

		d.Div(d, two)
	}

	if len(wnaf) == 0 {
		wnaf = append(wnaf, 0)
	}

	return
}
