package math

func Phi(n int) int {
	if n == 0 {
		panic("invalid input")
	}
	if n == 1 {
		return 1
	}
	// store numbers in range [2,n]
	numbers := make([]int, n-1)
	for i := range numbers {
		numbers[i] = i + 2
	}

	// sieve method
	for i, v := range numbers {
		if v == i+2 {
			// v is a prime
			for j := i; j <= n-2; j += v {
				numbers[j] /= v
				numbers[j] *= v - 1
			}
		}
	}

	return numbers[n-2]
}
