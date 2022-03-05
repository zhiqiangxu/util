package math

import "math/big"

func DivCount(n int) int {

	if n <= 0 {
		panic("invalid input")
	}
	if n == 1 {
		return 1
	}

	// store numbers in range [2,n]
	numbers := make([]bool, n-1)
	for i := range numbers {
		numbers[i] = true
	}

	// sieve method for prime calculation
	for i, v := range numbers {
		p := i + 2
		if v {
			for j := i + p; j <= n-2; j += p {
				numbers[j] = false
			}
		}
		// if n can factor, then there's at least one <= sqrt(n)
		if p*p >= n {
			break
		}
	}

	// Traversing through all prime numbers
	total := 1
	for i, v := range numbers {
		if v {
			// calculate number of divisor
			// with formula total div =
			// (p1+1) * (p2+1) *.....* (pn+1)
			// where n = (a1^p1)*(a2^p2)....
			// *(an^pn) ai being prime divisor
			// for n and pi are their respective
			// power in factorization
			count := 0
			p := i + 2
			for n%p == 0 {
				n /= p
				count++
			}
			if count > 0 {
				total *= 1 + count
			}
		}
	}

	return total
}

// AbelGroups returns the # of abel groups of order n
// it's based on the fundamental theory of algebra that
// every finitely generated abel group is isomorphic to
// direct sum of prime-power cyclic groups.
func AbelGroups(n int) *big.Int {
	if n <= 0 {
		panic("invalid input")
	}
	if n == 1 {
		return big.NewInt(1)
	}

	// store numbers in range [2,n]
	numbers := make([]bool, n-1)
	for i := range numbers {
		numbers[i] = true
	}

	// sieve method for prime calculation
	for i, v := range numbers {
		p := i + 2
		if v {
			for j := i + p; j <= n-2; j += p {
				numbers[j] = false
			}
		}
		// if n can factor, then there's at least one <= sqrt(n)
		if p*p >= n {
			break
		}
	}

	// Traversing through all prime numbers
	total := big.NewInt(1)
	for i, v := range numbers {
		if v {
			count := 0
			p := i + 2
			for n%p == 0 {
				n /= p
				count++
			}
			if count > 0 {
				total.Mul(total, Partition(count))
			}
		}
	}

	return total
}
