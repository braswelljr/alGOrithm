package euclidean

// Euclidean algorithm | Greatest common divisor
// https://en.wikipedia.org/wiki/Euclidean_algorithm
func Euclidean(a, b int) int {
	if b == 0 {
		return a
	}
	return Euclidean(b, a%b)
}
