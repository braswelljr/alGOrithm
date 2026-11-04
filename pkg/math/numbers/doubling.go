package numbers

// FibonacciDoubling returns F(n) with the fast-doubling identities
// F(2k) = F(k) * (2F(k+1) - F(k)) and F(2k+1) = F(k)^2 + F(k+1)^2, walking the
// bits of n from the top. It does the same O(log n) work as FibonacciMatrix
// with about half the multiplications. It returns ErrNumber for n outside
// 0..92, beyond which F(n) overflows an int64.
//
// Time: O(log n). Space: O(1).
func FibonacciDoubling(n int) (int, error) {
	if n < 0 || n > 92 {
		return 0, ErrNumber
	}
	var a, b uint64 = 0, 1 // F(k), F(k+1), with k the bits of n read so far
	for bit := 6; bit >= 0; bit-- {
		c := a * (2*b - a)
		d := a*a + b*b
		a, b = c, d
		if n>>bit&1 == 1 {
			a, b = b, a+b
		}
	}
	return int(a), nil
}
