package euclidean

// Euclidean algorithm | The Greatest common divisor
// https://en.wikipedia.org/wiki/Euclidean_algorithm
func Euclidean(a, b int) int {
	if b == 0 {
		return a
	}
  // Recur the function with parameters b and (a modulo b)
	return Euclidean(b, a%b)
}
