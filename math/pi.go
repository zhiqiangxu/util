package math

// Pi returns # of primes LTE than n
func Pi(n int) int {
	if n <= 1 {
		return 0
	}

	// store numbers in range [2,n]
	numbers := make([]int, n-1)
	for i := range numbers {
		numbers[i] = i + 2
	}

	count := 0
	// sieve method
	for i, v := range numbers {
		if v == i+2 {
			count++
			// v is a prime
			for j := i + v; j <= n-2; j += v {
				numbers[j] = 0
			}
		}
	}

	return count

}
