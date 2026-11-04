package numbers

import "slices"

// Divisors returns every positive divisor of n in ascending order, pairing d
// with n/d while d*d <= n; nil for n < 1.
//
// Time: O(sqrt n). Space: O(d(n)).
func Divisors(n int) []int {
	if n < 1 {
		return nil
	}
	var small, large []int
	for d := 1; d*d <= n; d++ {
		if n%d == 0 {
			small = append(small, d)
			if d != n/d {
				large = append(large, n/d)
			}
		}
	}
	slices.Reverse(large) // n/d came out largest first
	return append(small, large...)
}

// PerfectNumbers returns the perfect numbers up to limit, those equal to the
// sum of their divisors below themselves (6 = 1 + 2 + 3, 28, 496, 8128). It
// adds each d to the divisor sum of its multiples, a sieve over divisors. nil
// for limit < 1 or above MaxSieve. Every even perfect number is
// 2^(p-1)(2^p - 1) for a Mersenne prime 2^p - 1; no odd one is known.
//
// Time: O(n log n). Space: O(n).
func PerfectNumbers(limit int) []int {
	if limit < 1 || limit > MaxSieve {
		return nil
	}
	sum := make([]int, limit+1)
	for d := 1; d <= limit/2; d++ {
		for m := 2 * d; m <= limit; m += d {
			sum[m] += d
		}
	}
	var out []int
	for n := 2; n <= limit; n++ {
		if sum[n] == n {
			out = append(out, n)
		}
	}
	return out
}
