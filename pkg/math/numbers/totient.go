package numbers

// EulerTotient returns phi(n), how many of 1..n are coprime to n, from its
// prime factors: phi(n) = n * product of (1 - 1/p) over the distinct primes p
// dividing n. phi(p) = p - 1 for a prime, and Euler's theorem says
// a^phi(n) = 1 (mod n) for a coprime to n, which RSA relies on. 0 for n < 1.
//
// Time: O(sqrt n). Space: O(1).
func EulerTotient(n int) int {
	if n < 1 {
		return 0
	}
	result := n
	for p := 2; p*p <= n; p++ {
		if n%p == 0 {
			for n%p == 0 {
				n /= p
			}
			result -= result / p
		}
	}
	if n > 1 {
		result -= result / n
	}
	return result
}

// TotientSieve returns phi(0..limit), phi(0) being 0, with a sieve: start
// with phi(i) = i, and for each prime p (a value still equal to its index)
// apply the factor (1 - 1/p) to every multiple. nil for limit < 0 or above
// MaxSieve.
//
// Time: O(n log log n). Space: O(n).
func TotientSieve(limit int) []int {
	if limit < 0 || limit > MaxSieve {
		return nil
	}
	phi := make([]int, limit+1)
	for i := range phi {
		phi[i] = i
	}
	for p := 2; p <= limit; p++ {
		if phi[p] == p { // untouched, so prime
			for m := p; m <= limit; m += p {
				phi[m] -= phi[m] / p
			}
		}
	}
	return phi
}
