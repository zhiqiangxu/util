package math

// FYI : https://www.math.sinica.edu.tw/www/file_upload/summer/crypt2017/data/2015/[20150707][carmichael].pdf

func Lambda(n int) int {
	if n == 0 {
		panic("invalid input")
	}
	if n == 1 {
		return 1
	}

	factors := PrimeFactors(n)

	var parts []int
	for prime, power := range factors {
		switch {
		case prime == 2:
			if power == 1 || power == 2 {
				parts = append(parts, PowInts(2, power-1))
			} else {
				parts = append(parts, PowInts(2, power-2))
			}
		default:
			parts = append(parts, PowInts(prime, power)-PowInts(prime, power-1))
		}
	}

	switch {
	case len(parts) == 1:
		return parts[0]
	case len(parts) == 2:
		return LCM(parts[0], parts[1])
	default:
		return LCM(parts[0], parts[1], parts[2:]...)
	}
}

// greatest common divisor (GCD) via Euclidean algorithm
func GCD(a, b int) int {
	for b != 0 {
		t := b
		b = a % b
		a = t
	}
	return a
}

// find Least Common Multiple (LCM) via GCD
func LCM(a, b int, integers ...int) int {
	result := a * b / GCD(a, b)

	for i := 0; i < len(integers); i++ {
		result = LCM(result, integers[i])
	}

	return result
}

// Assumption: n >= 0
func PowInts(x, n int) int {
	if n == 0 {
		return 1
	}

	v := 1
	for n != 0 {
		if n&1 == 1 {
			v *= x
		}
		x *= x
		n /= 2
	}

	return v
}

// Get all prime factors of a given number n
func PrimeFactors(n int) (pfs map[int]int) {
	// Get the number of 2s that divide n

	pfs = make(map[int]int)
	for n%2 == 0 {
		pfs[2] += 1
		n = n / 2
	}

	// n must be odd at this point. so we can skip one element
	// (note i = i + 2)
	for i := 3; i*i <= n; i = i + 2 {
		// while i divides n, append i and divide n
		for n%i == 0 {
			pfs[i] += 1
			n = n / i
		}
	}

	// This condition is to handle the case when n is a prime number
	// greater than 2
	if n > 2 {
		pfs[n] += 1
	}

	return
}
