package wnaf

import (
	"math/big"
	"math/bits"
)

// FYI https://rd.springer.com/content/pdf/10.1007/978-3-540-68914-0_26.pdf

// the returned wnaf "bits" are in big endian and in modified wNAF* version
func ToWnafLR(d *big.Int, w uint8) (wnaf []int64) {
	if d.Cmp(zero) < 0 {
		panic("negative d not supported")
	}
	if w >= 64 {
		panic("w >= 64 not supported")
	}

	n := d.BitLen()
	if n == 0 {
		return []int64{0}
	}

	i := n - 1
	c := 0

	mod := big.NewInt(0).Exp(two, big.NewInt(int64(w)), nil)
	w_all1 := big.NewInt(0).Sub(mod, one)
	appendWnaf := func(r int64, w uint8) {
		// the caller ensures that |r| is < 2^(w-1)

		var abs uint64
		if r > 0 {
			abs = uint64(r)
		} else {
			abs = uint64(-r)
		}
		nz := bits.TrailingZeros64(abs)
		// append w-1-nz 0 to wnaf
		wnaf = append(wnaf, make([]int64, w-1-uint8(nz))...)
		wnaf = append(wnaf, r>>nz)
		// append nz 0 to wnaf
		wnaf = append(wnaf, make([]int64, nz)...)
	}

	for {
		if i < -1 {
			break
		}

		var di uint
		// when i < 0, di is 0
		if i >= 0 {
			di = d.Bit(i)
		}

		switch c {
		case 0:
			if di == 0 {
				if i < 0 {
					return
				}

				wnaf = append(wnaf, 0)
				i -= 1
				continue
			}

			// here it's guaranteed that di == 1
			j := i - int(w) + 1
			if j < 0 || d.Bit(j) == 0 {
				if j < 0 {
					mod := big.NewInt(0).Exp(two, big.NewInt(int64(i+1)), nil)
					r := big.NewInt(0).Mod(d, mod).Uint64()
					appendWnaf(int64(r), uint8(i+1))

					i = -1
				} else {
					r := big.NewInt(0).Mod(big.NewInt(0).Rsh(d, uint(j)), mod).Uint64()
					appendWnaf(int64(r), w)

					i -= int(w)
				}
				continue
			}

			// here it's guaranteed that j >= 0 and dj == 1

			r := big.NewInt(0).Mod(big.NewInt(0).Rsh(d, uint(j)), mod)
			if r.Cmp(w_all1) == 0 {
				// all w bits are 1
				wnaf = append(wnaf, 2)
				// append w-1 0 to wnaf
				zeros := make([]int64, w-1)
				wnaf = append(wnaf, zeros...)

				i -= int(w)
				c = -2
			} else {
				// with 0 between i and j

				rPlus1 := r.Uint64() + 1
				appendWnaf(int64(rPlus1), w)

				i -= int(w)
				c = -2
			}
		case -2:
			if di == 1 {
				wnaf = append(wnaf, 0)
				i -= 1
				// c doesn't change
				continue
			}

			// here it's guaranteed that di == 0
			j := i - int(w) + 1
			if j < 0 || d.Bit(j) == 0 {

				if j < 0 {
					if i < 0 {
						wnaf = append(wnaf, -2)
						return
					} else {
						mod := big.NewInt(0).Exp(two, big.NewInt(int64(i+1)), nil)
						r := big.NewInt(0).Mod(d, mod)
						appendWnaf(-int64(big.NewInt(0).Sub(mod, r).Uint64()), uint8(i+1))
						i = -1
					}

				} else {
					r := big.NewInt(0).Mod(big.NewInt(0).Rsh(d, uint(j)), mod)
					appendWnaf(-int64(big.NewInt(0).Sub(mod, r).Uint64()), w)

					i -= int(w)
				}

				c = 0
				continue
			}

			// here it's guaranteed that j >=0 and dj == 1

			r := big.NewInt(0).Mod(big.NewInt(0).Rsh(d, uint(j)), mod)
			appendWnaf(-int64(big.NewInt(0).Sub(mod, big.NewInt(0).Add(r, one)).Uint64()), w) // w < 64, so the type cast is safe

			i -= int(w)
			// c doesn't change
		}

	}

	return
}
