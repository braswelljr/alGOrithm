package recursion

import "errors"

// ErrRange is returned for an argument outside what a function computes
// exactly (a factorial above 20! overflows int64).
var ErrRange = errors.New("recursion: argument out of range")

// Factorial returns n! = n * (n-1) * ... * 1, with 0! = 1, recursively. It
// returns ErrRange for n < 0 or n > 20, whose results do not fit in an int64.
//
// Time: O(n). Space: O(n) for the recursion.
func Factorial(n int) (int, error) {
	if n < 0 || n > 20 {
		return 0, ErrRange
	}
	if n <= 1 {
		return 1, nil
	}
	f, _ := Factorial(n - 1)
	return n * f, nil
}
