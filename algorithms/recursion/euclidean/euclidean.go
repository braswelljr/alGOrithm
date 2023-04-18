package euclidean

// Euclidean algorithm - is an efficient method for computing the greatest common divisor (GCD) of two numbers.
//
//	@param a - non-negative integer
//	@param b - non-negative integer
//	@return - non-negative integer - the greatest common divisor of a and b
func Euclidean(a, b int) int {
	// return a as the greatest common divisor if b == 0
	if b == 0 {
		return a
	}
	// Recur the function with parameters b and (a modulo b)
	return Euclidean(b, a%b)
}
