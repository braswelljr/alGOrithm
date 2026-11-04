// Package numbers holds number algorithms, one file each: evenness, primality
// by trial division, the sieve of Eratosthenes, prime factorization, fast
// exponentiation, and pagination arithmetic.
//
//	numbers.Sieve(30)        // [2 3 5 7 11 13 17 19 23 29]
//	numbers.Factorize(360)   // [2 2 2 3 3 5]
package numbers

// Even reports whether n is even. Time: O(1).
func Even(n int) bool { return n%2 == 0 }

// Prime reports whether n is prime, by trial division up to sqrt(n): false
// below 2.
//
// Time: O(sqrt n). Space: O(1).
func Prime(n int) bool {
	if n < 2 {
		return false
	}
	for d := 2; d*d <= n; d++ {
		if n%d == 0 {
			return false
		}
	}
	return true
}

// Factorize returns the prime factors of n in ascending order, with
// repetition (360 is [2 2 2 3 3 5]); nil for n < 2.
//
// Time: O(sqrt n). Space: O(log n).
func Factorize(n int) []int {
	var out []int
	for d := 2; d*d <= n; d++ {
		for n%d == 0 {
			out = append(out, d)
			n /= d
		}
	}
	if n > 1 {
		out = append(out, n)
	}
	return out
}

// Power returns base^exp modulo mod by fast exponentiation (square and
// multiply), which takes O(log exp) multiplications instead of exp. mod must
// be at least 1 and small enough that (mod-1)^2 fits in an int64.
//
// Time: O(log exp). Space: O(1).
func Power(base, exp, mod int) int {
	result := 1 % mod
	base = ((base % mod) + mod) % mod
	for ; exp > 0; exp >>= 1 {
		if exp&1 == 1 {
			result = result * base % mod
		}
		base = base * base % mod
	}
	return result
}
