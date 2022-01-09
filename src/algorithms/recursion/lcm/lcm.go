package lcm

import (
	"github.com/braswelljr/alGOrithm/src/algorithms/recursion/euclidean"
)

// LCM returns the least common multiple of two integers.
// The least common multiple of two integers a and b is the smallest positive integer that is divisible by both a and b.
// Depends on Euclidean algorithm(The Greatest Common Divisor).
func LCM(a, b int) int {
	return a * b / euclidean.Euclidean(a, b)
}
